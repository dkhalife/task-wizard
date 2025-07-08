package ws

import (
	"context"
	"net/http"
	"sync"

	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/services/logging"
	authutil "dkhalife.com/tasks/core/internal/utils/auth"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// connection represents a single websocket connection with associated identity.
type connection struct {
	ws       *websocket.Conn
	identity *models.SignedInIdentity
}

// WSServer keeps track of active websocket connections.
type WSServer struct {
	upgrader    websocket.Upgrader
	mu          sync.Mutex
	connections map[*websocket.Conn]*connection
}

// NewWSServer creates a new websocket server instance.
func NewWSServer() *WSServer {
	return &WSServer{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		connections: make(map[*websocket.Conn]*connection),
	}
}

// HandleConnection upgrades an HTTP request to a WebSocket connection and stores the
// associated SignedInIdentity on success.
func (s *WSServer) HandleConnection(c *gin.Context) {
	identity := authutil.CurrentIdentity(c)
	if identity == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	wsConn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logging.FromContext(c).Errorf("websocket upgrade error: %v", err)
		return
	}

	s.mu.Lock()
	s.connections[wsConn] = &connection{ws: wsConn, identity: identity}
	s.mu.Unlock()

	go s.listen(c, wsConn)
}

// listen waits for messages on a connection and removes it when closed.
func (s *WSServer) listen(ctx context.Context, wsConn *websocket.Conn) {
	defer func() {
		logging.FromContext(ctx).Debugf("cleaning up websocket connection")
		wsConn.Close()
		s.mu.Lock()
		delete(s.connections, wsConn)
		s.mu.Unlock()
	}()

	for {
		if _, _, err := wsConn.ReadMessage(); err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logging.FromContext(ctx).Debugf("websocket connection closed: %v", err)
			} else {
				logging.FromContext(ctx).Errorf("websocket read error: %v", err)
			}
			return
		}
	}
}

// Routes registers WebSocket routes.
func Routes(router *gin.Engine, s *WSServer, auth *jwt.GinJWTMiddleware) {
	router.GET("/api/ws", auth.MiddlewareFunc(), s.HandleConnection)
}

package ws

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/internal/models"
	lRepo "dkhalife.com/tasks/core/internal/repos/label"
	tRepo "dkhalife.com/tasks/core/internal/repos/task"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
	"dkhalife.com/tasks/core/internal/utils/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
)

// connection represents a single websocket connection with associated identity.
type connection struct {
	ws       *websocket.Conn
	identity *models.SignedInIdentity
	writeMu  sync.Mutex // Protects concurrent writes to websocket
}

// WSServer keeps track of active websocket connections.
type WSServer struct {
	upgrader        websocket.Upgrader
	mu              sync.RWMutex
	connections     map[*websocket.Conn]*connection
	userConnections map[int]map[*websocket.Conn]*connection
	handlers        map[string]messageHandler
	pingPeriod      time.Duration
	pongWait        time.Duration
	cfg             *config.Config
	tRepo           *tRepo.TaskRepository
	lRepo           *lRepo.LabelRepository
	uRepo           uRepo.IUserRepo
}

type messageHandler func(ctx context.Context, userID int, msg WSMessage) *WSResponse

// NewWSServer creates a new websocket server instance.
func NewWSServer(cfg *config.Config, tRepo *tRepo.TaskRepository, lRepo *lRepo.LabelRepository, uRepo uRepo.IUserRepo) *WSServer {
	return &WSServer{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		connections:     make(map[*websocket.Conn]*connection),
		userConnections: make(map[int]map[*websocket.Conn]*connection),
		handlers:        make(map[string]messageHandler),
		pongWait:        60 * time.Second,
		pingPeriod:      54 * time.Second,
		cfg:             cfg,
		tRepo:           tRepo,
		lRepo:           lRepo,
		uRepo:           uRepo,
	}
}

// RegisterHandler registers a handler for a specific action. Only one handler
// can be registered per action. Registering multiple handlers for the same
// action will cause a panic.
func (s *WSServer) RegisterHandler(action string, handler messageHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.handlers[action]; ok {
		panic(fmt.Sprintf("handler already registered for action %s", action))
	}
	s.handlers[action] = handler
}

// HandleConnection upgrades an HTTP request to a WebSocket connection and stores the
// associated SignedInIdentity on success.
func (s *WSServer) HandleConnection(c *gin.Context) {
	protocols := c.GetHeader("Sec-WebSocket-Protocol")
	protocolsList := strings.Split(protocols, ",")
	if len(protocolsList) != 2 {
		logging.FromContext(c).Debug("no websocket protocol provided")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	protocol := strings.TrimSpace(protocolsList[0])
	bearerToken := strings.TrimSpace(protocolsList[1])

	if protocol == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(bearerToken, func(t *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod("HS256") != t.Method {
			return nil, errors.New("invalid signing algorithm")
		}

		return []byte(s.cfg.Jwt.Secret), nil
	})

	if err != nil || token == nil || !token.Valid {
		logging.FromContext(c).Debug("token is invalid or missing")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims := token.Claims.(jwt.MapClaims)

	userIdRaw, ok := claims[auth.IdentityKey]
	if !ok {
		logging.FromContext(c).Debugf("user ID not found in claims")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userIdRawStr, ok := userIdRaw.(string)
	if !ok {
		logging.FromContext(c).Debugf("user ID is not a string")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	identityType, ok := claims["type"]
	if !ok {
		logging.FromContext(c).Debugf("identity type not found in claims")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	identityTypeStr, ok := identityType.(string)
	if !ok {
		logging.FromContext(c).Debugf("identity type is not a string")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	identityType = models.IdentityType(identityTypeStr)
	if identityType != models.IdentityTypeUser {
		logging.FromContext(c).Debug("identity type is not user")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(userIdRawStr)
	if err != nil {
		logging.FromContext(c).Debugf("failed to convert user ID to integer: %v", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	scopesRaw, ok := claims["scopes"].([]interface{})
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var scopes []models.ApiTokenScope
	for _, scope := range scopesRaw {
		if s, ok := scope.(string); ok {
			scopes = append(scopes, models.ApiTokenScope(s))
		}
	}

	identity := &models.SignedInIdentity{
		UserID:  userID,
		TokenID: 0,
		Type:    models.IdentityTypeUser,
		Scopes:  scopes,
	}

	wsConn, err := s.upgrader.Upgrade(c.Writer, c.Request, http.Header{
		"Sec-WebSocket-Protocol": []string{protocol},
	})
	if err != nil {
		logging.FromContext(c).Errorf("websocket upgrade error: %v", err)
		return
	}

	conn := &connection{ws: wsConn, identity: identity}

	s.mu.Lock()
	s.connections[wsConn] = conn
	if s.userConnections[identity.UserID] == nil {
		s.userConnections[identity.UserID] = make(map[*websocket.Conn]*connection)
	}
	s.userConnections[identity.UserID][wsConn] = conn
	s.mu.Unlock()

	go s.listen(c, conn)
}

// listen waits for messages on a connection and removes it when closed.
func (s *WSServer) listen(ctx context.Context, conn *connection) {
	wsConn := conn.ws
	if err := wsConn.SetReadDeadline(time.Now().Add(s.pongWait)); err != nil {
		logging.FromContext(ctx).Errorf("set read deadline error: %v", err)
	}
	wsConn.SetPongHandler(func(string) error {
		if err := wsConn.SetReadDeadline(time.Now().Add(s.pongWait)); err != nil {
			logging.FromContext(ctx).Errorf("set read deadline error: %v", err)
		}
		return nil
	})

	ticker := time.NewTicker(s.pingPeriod)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := conn.safeWriteMessage(websocket.PingMessage, nil); err != nil {
					logging.FromContext(ctx).Errorf("websocket ping error: %v", err)
					if cerr := conn.safeClose(); cerr != nil {
						logging.FromContext(ctx).Errorf("websocket close error: %v", cerr)
					}
					return
				}
			case <-done:
				return
			}
		}
	}()

	defer func() {
		close(done)
		ticker.Stop()
		logging.FromContext(ctx).Debugf("cleaning up websocket connection")
		if err := conn.safeClose(); err != nil {
			logging.FromContext(ctx).Errorf("websocket close error: %v", err)
		}
		s.mu.Lock()
		delete(s.connections, wsConn)
		if uMap, ok := s.userConnections[conn.identity.UserID]; ok {
			delete(uMap, wsConn)
			if len(uMap) == 0 {
				delete(s.userConnections, conn.identity.UserID)
			}
		}
		s.mu.Unlock()
	}()

	for {
		var msg WSMessage
		if err := wsConn.ReadJSON(&msg); err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logging.FromContext(ctx).Debugf("websocket connection closed: %v", err)
			} else {
				logging.FromContext(ctx).Errorf("websocket read error: %v", err)
			}
			return
		}

		s.handleMessage(ctx, conn, msg)
	}
}

// safeWriteMessage writes a message to the websocket connection with proper synchronization.
func (c *connection) safeWriteMessage(messageType int, data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.ws.WriteMessage(messageType, data)
}

// safeWriteJSON writes a JSON message to the websocket connection with proper synchronization.
func (c *connection) safeWriteJSON(v interface{}) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.ws.WriteJSON(v)
}

// safeClose closes the websocket connection with proper synchronization.
func (c *connection) safeClose() error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.ws.Close()
}

// Routes registers WebSocket routes.
func Routes(router *gin.Engine, s *WSServer) {
	router.GET("/ws", s.HandleConnection)
}

func (s *WSServer) handleMessage(ctx context.Context, conn *connection, msg WSMessage) {
	s.mu.RLock()
	handler, ok := s.handlers[msg.Action]
	s.mu.RUnlock()

	log := logging.FromContext(ctx)

	if !ok {
		log.Errorf("no handler registered for action %s", msg.Action)
		return
	}

	resp := handler(ctx, conn.identity.UserID, msg)

	if resp == nil {
		return
	}

	resp.Action = msg.Action

	if err := conn.safeWriteJSON(resp); err != nil {
		log.Errorf("failed to write JSON to WebSocket: %v", err)
		return
	}
}

func (s *WSServer) BroadcastToUser(userID int, resp WSResponse) {
	go func() {
		s.mu.RLock()
		conns := s.userConnections[userID]
		s.mu.RUnlock()

		log := logging.FromContext(context.Background())

		for _, c := range conns {
			if err := c.safeWriteJSON(resp); err != nil {
				log.Errorf("Failed to write JSON to WebSocket: %v", err)
			}
		}
	}()
}

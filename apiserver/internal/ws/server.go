package ws

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"dkhalife.com/tasks/core/internal/models"
	lRepo "dkhalife.com/tasks/core/internal/repos/label"
	tRepo "dkhalife.com/tasks/core/internal/repos/task"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
	authMW "dkhalife.com/tasks/core/internal/middleware/auth"
	"dkhalife.com/tasks/core/internal/telemetry"
	"github.com/gin-gonic/gin"
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
	authMiddleware  *authMW.AuthMiddleware
	tRepo           *tRepo.TaskRepository
	lRepo           *lRepo.LabelRepository
	uRepo           uRepo.IUserRepo
}

type messageHandler func(ctx context.Context, userID int, msg WSMessage) *WSResponse

// NewWSServer creates a new websocket server instance.
func NewWSServer(authMiddleware *authMW.AuthMiddleware, tRepo *tRepo.TaskRepository, lRepo *lRepo.LabelRepository, uRepo uRepo.IUserRepo) *WSServer {
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
		authMiddleware:  authMiddleware,
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
		telemetry.TrackWarning(c, "ws_unauthorized", "ws-server", "No websocket protocol provided", nil)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	protocol := strings.TrimSpace(protocolsList[0])
	bearerToken := strings.TrimSpace(protocolsList[1])

	if protocol == "" {
		telemetry.TrackWarning(c, "ws_unauthorized", "ws-server", "Empty websocket protocol", nil)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	identity, err := s.authMiddleware.VerifyWSToken(c.Request.Context(), bearerToken)
	if err != nil {
		logging.FromContext(c).Debugf("token verification failed: %v", err)
		telemetry.TrackWarning(c, "ws_unauthorized", "ws-server", "Token verification failed", nil)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if identity.Type != models.IdentityTypeUser {
		logging.FromContext(c).Debug("identity type is not user")
		telemetry.TrackWarning(c, "ws_unauthorized", "ws-server", "Identity type is not user", nil)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	wsConn, err := s.upgrader.Upgrade(c.Writer, c.Request, http.Header{
		"Sec-WebSocket-Protocol": []string{protocol},
	})
	if err != nil {
		logging.FromContext(c).Errorf("websocket upgrade error: %v", err)
		telemetry.TrackError(c, "ws_upgrade_failed", "ws-server", err, nil)
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
				telemetry.TrackError(context.Background(), "ws_read_failed", "ws-server", err, nil)
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

// wsReadOnlyActions contains WS actions that only read data and are always permitted,
// including for accounts with pending deletion.
var wsReadOnlyActions = map[string]struct{}{
	"get_tasks":           {},
	"get_completed_tasks": {},
	"get_task":            {},
	"get_task_history":    {},
	"get_user_labels":     {},
}

func (s *WSServer) handleMessage(ctx context.Context, conn *connection, msg WSMessage) {
	s.mu.RLock()
	handler, ok := s.handlers[msg.Action]
	s.mu.RUnlock()

	log := logging.FromContext(ctx)

	if !ok {
		log.Errorf("no handler registered for action %s", msg.Action)
		telemetry.TrackWarning(context.Background(), "ws_unknown_action", "ws-server", "No handler for action: "+msg.Action, nil)
		return
	}

	if conn.identity.PendingDeletion {
		if _, readOnly := wsReadOnlyActions[msg.Action]; !readOnly {
			resp := &WSResponse{
				Action:    msg.Action,
				RequestID: msg.RequestID,
				Status:    http.StatusForbidden,
				Data:      map[string]string{"error": "Account is pending deletion"},
			}
			if err := conn.safeWriteJSON(resp); err != nil {
				log.Errorf("failed to write JSON to WebSocket: %v", err)
			}
			return
		}
	}

	resp := handler(ctx, conn.identity.UserID, msg)

	if resp == nil {
		return
	}

	resp.Action = msg.Action
	resp.RequestID = msg.RequestID

	if err := conn.safeWriteJSON(resp); err != nil {
		log.Errorf("failed to write JSON to WebSocket: %v", err)
		telemetry.TrackError(context.Background(), "ws_write_failed", "ws-server", err, nil)
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
				telemetry.TrackError(context.Background(), "ws_broadcast_write_failed", "ws-server", err, nil)
			}
		}
	}()
}

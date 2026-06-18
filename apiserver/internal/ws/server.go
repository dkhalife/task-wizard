package ws

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"taskwiz.app/core/config"
	authMW "taskwiz.app/core/internal/middleware/auth"
	"taskwiz.app/core/internal/models"
	lRepo "taskwiz.app/core/internal/repos/label"
	tRepo "taskwiz.app/core/internal/repos/task"
	uRepo "taskwiz.app/core/internal/repos/user"
	"taskwiz.app/core/internal/services/logging"
	"taskwiz.app/core/internal/telemetry"
)

const (
	// maxMessageBytes caps the size of a single inbound WebSocket message. WS
	// payloads are small JSON mutations, so anything larger is rejected to avoid
	// memory abuse from a flooding client.
	maxMessageBytes = 32 * 1024
	// messageRatePerSecond and messageBurst bound how frequently a single
	// connection may send messages, mirroring the rate limiting applied to REST
	// routes.
	messageRatePerSecond = 20
	messageBurst         = 40
)

// tokenBucket is a small, self-contained token-bucket rate limiter used to bound
// the message rate of a single WebSocket connection.
type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64
	last       time.Time
}

func newTokenBucket(ratePerSecond, burst float64) *tokenBucket {
	return &tokenBucket{
		tokens:     burst,
		maxTokens:  burst,
		refillRate: ratePerSecond,
		last:       time.Now(),
	}
}

// allow reports whether a message may be processed now, consuming one token when
// it can.
func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	b.tokens += now.Sub(b.last).Seconds() * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.last = now

	if b.tokens < 1 {
		return false
	}

	b.tokens--
	return true
}

// connection represents a single websocket connection with associated identity.
type connection struct {
	ws       *websocket.Conn
	identity *models.SignedInIdentity
	writeMu  sync.Mutex // Protects concurrent writes to websocket
	limiter  *tokenBucket
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
func NewWSServer(cfg *config.Config, authMiddleware *authMW.AuthMiddleware, tRepo *tRepo.TaskRepository, lRepo *lRepo.LabelRepository, uRepo uRepo.IUserRepo) *WSServer {
	return &WSServer{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     checkOrigin(cfg.Server.AllowedOrigins),
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

// checkOrigin builds a CheckOrigin function restricting WebSocket upgrades to the
// configured allowed origins, mirroring the HTTP CORS allow-list. When the list is
// empty the check stays permissive to preserve same-origin and local/dev deployments.
// Requests without an Origin header (non-browser clients) are always allowed.
func checkOrigin(allowedOrigins []string) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		if len(allowedOrigins) == 0 {
			return true
		}

		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}

		return false
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

	conn := &connection{
		ws:       wsConn,
		identity: identity,
		limiter:  newTokenBucket(messageRatePerSecond, messageBurst),
	}

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
	wsConn.SetReadLimit(maxMessageBytes)
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
					logging.FromContext(ctx).Warnf("websocket read error: %v", err)
					telemetry.TrackWarning(context.Background(), "ws_read_failed", "ws-server", err.Error(), nil)
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
	"get_tasks":        {},
	"get_activity":     {},
	"get_task":         {},
	"get_task_history": {},
	"get_user_labels":  {},
}

func (s *WSServer) handleMessage(ctx context.Context, conn *connection, msg WSMessage) {
	s.mu.RLock()
	handler, ok := s.handlers[msg.Action]
	s.mu.RUnlock()

	log := logging.FromContext(ctx)

	if conn.limiter != nil && !conn.limiter.allow() {
		log.Warnf("rate limit exceeded for user %d on action %s", conn.identity.UserID, msg.Action)
		telemetry.TrackWarning(context.Background(), "ws_rate_limited", "ws-server", "Too many messages", nil)
		resp := &WSResponse{
			Action:    msg.Action,
			RequestID: msg.RequestID,
			Status:    http.StatusTooManyRequests,
			Data:      map[string]string{"error": "Too many requests"},
		}
		if err := conn.safeWriteJSON(resp); err != nil {
			log.Errorf("failed to write JSON to WebSocket: %v", err)
		}
		return
	}

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

// SetPendingDeletionForUser updates the PendingDeletion flag on all active
// connections for a user so the WS write-guard reflects current deletion state
// without requiring a reconnect.
func (s *WSServer) SetPendingDeletionForUser(userID int, pending bool) {
	s.mu.RLock()
	conns := s.userConnections[userID]
	s.mu.RUnlock()

	for _, c := range conns {
		c.identity.PendingDeletion = pending
	}
}

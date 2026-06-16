package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
	"taskwiz.app/core/config"
	authMW "taskwiz.app/core/internal/middleware/auth"
	"taskwiz.app/core/internal/models"
	uRepo "taskwiz.app/core/internal/repos/user"
)

type mockWSUserRepo struct {
	uRepo.IUserRepo
}

func (m *mockWSUserRepo) EnsureUser(c context.Context, directoryID string, objectID string) (*models.User, error) {
	return &models.User{ID: 1, DirectoryID: directoryID, ObjectID: objectID}, nil
}

func (m *mockWSUserRepo) GetUser(c context.Context, id int) (*models.User, error) {
	return &models.User{ID: id}, nil
}

// WSServerTestSuite defines test suite for websocket server.
type WSServerTestSuite struct {
	suite.Suite
	router *gin.Engine
	server *WSServer
}

func TestWSServerTestSuite(t *testing.T) {
	suite.Run(t, new(WSServerTestSuite))
}

func (s *WSServerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Entra:  config.EntraConfig{Enabled: false},
		Server: config.ServerConfig{Registration: true},
	}

	mockRepo := &mockWSUserRepo{}
	authMiddleware, err := authMW.NewAuthMiddleware(cfg, mockRepo, nil)
	s.Require().NoError(err)

	s.server = NewWSServer(cfg, authMiddleware, nil, nil, mockRepo)
	s.router = gin.New()
	s.router.GET("/ws", s.server.HandleConnection)
}

func (s *WSServerTestSuite) dial(ts *httptest.Server) (*websocket.Conn, *http.Response, error) {
	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	header := http.Header{}
	header.Set("Sec-WebSocket-Protocol", "test-protocol, dummy-token")
	return websocket.DefaultDialer.Dial(url, header)
}

func (s *WSServerTestSuite) waitForConnections(n int) {
	s.Eventually(func() bool {
		s.server.mu.RLock()
		defer s.server.mu.RUnlock()
		return len(s.server.connections) == n
	}, time.Second, 10*time.Millisecond)
}

func (s *WSServerTestSuite) TestHandleConnection_Unauthorized() {
	ts := httptest.NewServer(s.router)
	defer ts.Close()

	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	header := http.Header{}
	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if conn != nil {
		conn.Close()
	}
	s.Error(err)
	s.Equal(0, len(s.server.connections))
}

func (s *WSServerTestSuite) TestHandleConnection_Authorized() {
	ts := httptest.NewServer(s.router)
	defer ts.Close()

	conn, resp, err := s.dial(ts)
	s.Require().NoError(err)
	s.Equal(http.StatusSwitchingProtocols, resp.StatusCode)
	s.waitForConnections(1)
	s.Equal(1, len(s.server.userConnections))

	conn.Close()
	s.waitForConnections(0)
	s.Equal(0, len(s.server.userConnections))
}

func (s *WSServerTestSuite) TestMultipleConnectionsAndCleanup() {
	ts := httptest.NewServer(s.router)
	defer ts.Close()

	conn1, _, err := s.dial(ts)
	s.Require().NoError(err)
	conn2, _, err := s.dial(ts)
	s.Require().NoError(err)

	s.waitForConnections(2)
	s.Equal(1, len(s.server.userConnections))
	s.Equal(2, len(s.server.userConnections[1]))

	conn1.Close()
	conn2.Close()
	s.waitForConnections(0)
	s.Equal(0, len(s.server.userConnections))
}

func (s *WSServerTestSuite) TestPingPongKeepsConnectionAlive() {
	s.server.pingPeriod = 50 * time.Millisecond
	s.server.pongWait = 200 * time.Millisecond

	ts := httptest.NewServer(s.router)
	defer ts.Close()

	conn, _, err := s.dial(ts)
	s.Require().NoError(err)

	conn.SetPongHandler(func(appData string) error {
		return nil
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	time.Sleep(3 * s.server.pingPeriod)
	s.Equal(1, len(s.server.connections))

	conn.Close()
	<-done
	s.waitForConnections(0)
}

func (s *WSServerTestSuite) TestRegisterHandler_DuplicatePanics() {
	handler := func(ctx context.Context, userID int, msg WSMessage) *WSResponse {
		return nil
	}

	s.server.RegisterHandler("dup", handler)
	s.Panics(func() { s.server.RegisterHandler("dup", handler) })
}

func (s *WSServerTestSuite) TestCheckOrigin() {
	newRequest := func(origin string) *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/ws", nil)
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		return r
	}

	allowed := []string{"https://app.example.com", "https://admin.example.com"}

	check := checkOrigin(allowed)
	s.True(check(newRequest("https://app.example.com")), "allowed origin should be accepted")
	s.False(check(newRequest("https://evil.example.com")), "disallowed origin should be rejected")
	s.True(check(newRequest("")), "missing origin should be accepted")

	permissive := checkOrigin(nil)
	s.True(permissive(newRequest("https://anything.example.com")), "empty allow-list should be permissive")
	s.True(permissive(newRequest("")), "empty allow-list with no origin should be permissive")
}

func (s *WSServerTestSuite) TestHandleMessageRoutesResponse() {
	s.server.RegisterHandler("echo", func(ctx context.Context, userID int, msg WSMessage) *WSResponse {
		return &WSResponse{
			Action: "echo",
			Data:   msg.Data,
		}
	})

	ts := httptest.NewServer(s.router)
	defer ts.Close()

	conn, _, err := s.dial(ts)
	s.Require().NoError(err)
	defer conn.Close()

	s.waitForConnections(1)

	payload := WSMessage{RequestID: "1", Action: "echo", Data: json.RawMessage(`"hello"`)}
	s.NoError(conn.WriteJSON(payload))

	var resp WSResponse
	s.NoError(conn.ReadJSON(&resp))
	s.Equal("1", resp.RequestID)
	s.Equal("echo", resp.Action)
	s.Equal("hello", resp.Data)
}

func (s *WSServerTestSuite) TestOversizedMessageClosesConnection() {
	ts := httptest.NewServer(s.router)
	defer ts.Close()

	conn, _, err := s.dial(ts)
	s.Require().NoError(err)
	defer conn.Close()

	s.waitForConnections(1)

	oversized := strings.Repeat("a", maxMessageBytes+1)
	s.NoError(conn.WriteMessage(websocket.TextMessage, []byte(oversized)))

	_, _, readErr := conn.ReadMessage()
	s.Error(readErr)
	s.waitForConnections(0)
}

func (s *WSServerTestSuite) TestRateLimitRejectsFlood() {
	s.server.RegisterHandler("noop", func(ctx context.Context, userID int, msg WSMessage) *WSResponse {
		return &WSResponse{Action: "noop", Status: http.StatusOK}
	})

	ts := httptest.NewServer(s.router)
	defer ts.Close()

	conn, _, err := s.dial(ts)
	s.Require().NoError(err)
	defer conn.Close()

	s.waitForConnections(1)

	rateLimited := false
	for i := 0; i < messageBurst+5; i++ {
		s.Require().NoError(conn.WriteJSON(WSMessage{RequestID: "r", Action: "noop"}))

		var resp WSResponse
		s.Require().NoError(conn.ReadJSON(&resp))
		if resp.Status == http.StatusTooManyRequests {
			rateLimited = true
			break
		}
	}

	s.True(rateLimited, "expected at least one message to be rate limited")
}

func (s *WSServerTestSuite) TestTokenBucketRefills() {
	b := newTokenBucket(1000, 2)
	s.True(b.allow())
	s.True(b.allow())
	s.False(b.allow())

	time.Sleep(10 * time.Millisecond)
	s.True(b.allow(), "tokens should refill over time")
}

package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dkhalife.com/tasks/core/internal/models"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
)

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
	s.server = NewWSServer(nil, nil, nil)
	s.router = gin.New()
	s.router.Use(func(c *gin.Context) {
		if c.GetHeader("X-Test-Auth") == "true" {
			c.Set(auth.IdentityKey, &models.SignedInIdentity{UserID: 1})
		}
	})
	s.router.GET("/ws", s.server.HandleConnection)
}

func (s *WSServerTestSuite) dial(ts *httptest.Server) (*websocket.Conn, *http.Response, error) {
	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	header := http.Header{}
	header.Set("X-Test-Auth", "true")
	return websocket.DefaultDialer.Dial(url, header)
}

func (s *WSServerTestSuite) waitForConnections(n int) {
	s.Eventually(func() bool {
		s.server.mu.Lock()
		defer s.server.mu.Unlock()
		return len(s.server.connections) == n
	}, time.Second, 10*time.Millisecond)
}

func (s *WSServerTestSuite) TestHandleConnection_Unauthorized() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ws", nil)

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusUnauthorized, w.Code)
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

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	time.Sleep(2 * s.server.pongWait)
	s.Equal(1, len(s.server.connections))

	conn.Close()
	s.waitForConnections(0)
}

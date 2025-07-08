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
	s.server = NewWSServer()
	s.router = gin.New()
	s.router.Use(func(c *gin.Context) {
		if c.GetHeader("X-Test-Auth") == "true" {
			c.Set(auth.IdentityKey, &models.SignedInIdentity{UserID: 1})
		}
	})
	s.router.GET("/ws", s.server.HandleConnection)
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

	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	header := http.Header{}
	header.Set("X-Test-Auth", "true")

	conn, resp, err := websocket.DefaultDialer.Dial(url, header)
	s.Require().NoError(err)
	s.Equal(http.StatusSwitchingProtocols, resp.StatusCode)
	s.Equal(1, len(s.server.connections))
	s.Equal(1, len(s.server.userConnections))

	conn.Close()
	time.Sleep(50 * time.Millisecond)

	s.Equal(0, len(s.server.connections))
	s.Equal(0, len(s.server.userConnections))
}

func (s *WSServerTestSuite) TestMultipleConnectionsAndCleanup() {
	ts := httptest.NewServer(s.router)
	defer ts.Close()

	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	header := http.Header{}
	header.Set("X-Test-Auth", "true")

	conn1, _, err := websocket.DefaultDialer.Dial(url, header)
	s.Require().NoError(err)
	conn2, _, err := websocket.DefaultDialer.Dial(url, header)
	s.Require().NoError(err)

	s.Equal(2, len(s.server.connections))
	s.Equal(1, len(s.server.userConnections))
	s.Equal(2, len(s.server.userConnections[1]))

	conn1.Close()
	conn2.Close()
	time.Sleep(50 * time.Millisecond)

	s.Equal(0, len(s.server.connections))
	s.Equal(0, len(s.server.userConnections))
}

func (s *WSServerTestSuite) TestPingPongKeepsConnectionAlive() {
	s.server.pingPeriod = 50 * time.Millisecond
	s.server.pongWait = 200 * time.Millisecond

	ts := httptest.NewServer(s.router)
	defer ts.Close()

	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	header := http.Header{}
	header.Set("X-Test-Auth", "true")

	conn, _, err := websocket.DefaultDialer.Dial(url, header)
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
	time.Sleep(50 * time.Millisecond)

	s.Equal(0, len(s.server.connections))
}

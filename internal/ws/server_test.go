package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dkhalife.com/tasks/core/config"
	auth "dkhalife.com/tasks/core/internal/utils/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
)

// WSServerTestSuite defines test suite for websocket server.
type WSServerTestSuite struct {
	suite.Suite
	router *gin.Engine
	server *WSServer
	cfg    *config.Config
}

func TestWSServerTestSuite(t *testing.T) {
	suite.Run(t, new(WSServerTestSuite))
}

func (s *WSServerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.cfg = &config.Config{
		Jwt: config.JwtConfig{
			Secret:      "test-secret",
			SessionTime: time.Hour,
			MaxRefresh:  time.Hour,
		},
	}
	s.server = NewWSServer(s.cfg, nil, nil, nil)
	s.router = gin.New()
	s.router.GET("/ws", s.server.HandleConnection)
}

func (s *WSServerTestSuite) createTestJWT(userID int) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		auth.IdentityKey: "1",
		"exp":            time.Now().Add(time.Hour).Unix(),
		"type":           "user",
		"scopes":         []string{"read", "write"},
	})

	signedToken, err := token.SignedString([]byte(s.cfg.Jwt.Secret))
	s.Require().NoError(err)
	return signedToken
}

func (s *WSServerTestSuite) dial(ts *httptest.Server) (*websocket.Conn, *http.Response, error) {
	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	header := http.Header{}
	jwtToken := s.createTestJWT(1)
	header.Set("Sec-WebSocket-Protocol", "test-protocol, "+jwtToken)
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
	ts := httptest.NewServer(s.router)
	defer ts.Close()

	// Try to connect without proper protocol header
	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	header := http.Header{}
	// Don't set the required Sec-WebSocket-Protocol header
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

	// Set up proper ping/pong handler
	conn.SetPongHandler(func(appData string) error {
		return nil
	})

	// Set up a goroutine to handle incoming messages
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
			// The ping/pong is handled automatically by the SetPongHandler
		}
	}()

	// Wait for a few ping cycles to ensure the connection stays alive
	time.Sleep(3 * s.server.pingPeriod)
	s.Equal(1, len(s.server.connections))

	conn.Close()
	<-done // Wait for the reading goroutine to finish
	s.waitForConnections(0)
}

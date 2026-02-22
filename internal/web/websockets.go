package web

import (
	"context"
	"net/http"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 5 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{ //nolint:gochecknoglobals
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

var newline = []byte{'\n'} //nolint:gochecknoglobals

type wsHub struct {
	// Registered clients.
	clients map[*wsClient]bool

	// Register requests from the clients.
	register chan *wsClient

	// Unregister requests from clients.
	unregister chan *wsClient

	ctx context.Context //nolint: containedctx
}

func newWSHub() *wsHub {
	return &wsHub{
		register:   make(chan *wsClient),
		unregister: make(chan *wsClient),
		clients:    make(map[*wsClient]bool),
	}
}

func (hub *wsHub) goRun(ctx context.Context, eventChan <-chan types.OutEvent) {
	hub.ctx = ctx

	go func() {
		for ctx.Err() == nil {
			select {
			case client := <-hub.register:
				hub.clients[client] = true
			case client := <-hub.unregister:
				if hub.clients[client] {
					delete(hub.clients, client)
					close(client.send)
				}
			case evt := <-eventChan:
				logger.Trace("[ws] Sending WS event: ", evt.String())

				eventMsg, err := evt.JSON()
				if err != nil {
					logger.Error(err)
				}

				for client := range hub.clients {
					select {
					case client.send <- eventMsg:
					default:
						close(client.send)
						delete(hub.clients, client)
					}
				}
			case <-ctx.Done():
				for client := range hub.clients {
					close(client.send)
				}

				return
			}
		}
	}()
}

// serveWs handles websocket requests from the peer.
func (hub *wsHub) serveWs(c *gin.Context) {
	if !c.IsWebsocket() {
		return
	}

	logger.Tracef("[ws] Registering new WS client from %s", c.Request.RemoteAddr)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Warnf("Failed to upgrade WS connection: %v", err)

		return
	}

	client := &wsClient{conn: conn, send: make(chan []byte, 256)}
	hub.register <- client

	// Allow collection of memory referenced by the caller
	// by doing all work in new goroutines.
	go client.writer(hub.ctx, hub.unregister)
	go client.reader(hub.ctx, hub.unregister)
}

type wsClient struct {
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *wsClient) writer(ctx context.Context, unregister chan *wsClient) {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		logger.Trace("[ws] Closing WS connection from ", c.conn.RemoteAddr())

		_ = c.conn.Close()
		select {
		case unregister <- c:
		case <-ctx.Done():
		}
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})

				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			_, _ = w.Write(message)

			// Add queued messages to the current websocket message.
			n := len(c.send)
			for range n {
				_, _ = w.Write(newline)
				_, _ = w.Write(<-c.send)
			}

			if err = w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				return
			}

			if err = c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *wsClient) reader(ctx context.Context, unregister chan *wsClient) {
	defer func() {
		logger.Trace("[ws] Closing WS reader from ", c.conn.RemoteAddr())

		_ = c.conn.Close()
		select {
		case unregister <- c:
		case <-ctx.Done():
		}
	}()

	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

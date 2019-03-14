package boxwebsock

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/okoshiyoshinori/votebox/logger"
)

const (
	writeWait      = 20 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Hubs *Hubs
	Conn *websocket.Conn
	Box  string
	Send chan []byte
	Stop chan []byte
}

func (c *Client) ReadPump() {
	defer func() {
		s := make(map[string]*Client)
		s[c.Box] = c
		c.Hubs.ClientUnRegister <- s
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Info.Printf("error: %v", err)
			}
			break
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Info.Println(err)
			}
			/*
				w, err := c.Conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				w.Write(message)
				n := len(<-c.Send)
				for i := 0; i < n; i++ {
					w.Write([]byte("\n"))
					w.Write(<-c.Send)
				}
				if err := w.Close(); err != nil {
					log.Println(err)
					return
				}
			*/
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				logger.Info.Println(err)
				return
			}
		case <-c.Stop:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
	}
}

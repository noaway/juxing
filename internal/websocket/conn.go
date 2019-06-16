package websocket

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Base interface
type Base interface {
	init(*websocket.Conn, Dial, ...*logrus.Entry) *Conn
}

// Dial interface
type Dial interface {
	Base

	OnOpen() error
	OnMessage(messageType int, p []byte)
	Send(messageType int, data []byte) error
	OnClose() error
}

// err var
var (
	ErrConnNotInitialized = fmt.Errorf("%v", "connection not initialized")
)

type msg struct {
	messageType int
	p           []byte
}

// Conn struct
type Conn struct {
	sync.Once

	readChan  chan *msg
	writeChan chan *msg
	closeChan chan struct{}
	Logger    *logrus.Entry
	conn      *websocket.Conn
	dial      Dial
	err       error
}

// Init fn
func (c *Conn) init(conn *websocket.Conn, dial Dial, entrys ...*logrus.Entry) *Conn {
	if c == nil {
		c = new(Conn)
	}
	if conn != nil && dial != nil {
		c.conn = conn
		c.dial = dial
	}
	var entry = &logrus.Entry{}
	if len(entrys) > 0 {
		entry = entrys[0]
	}
	c.readChan = make(chan *msg, 100)
	c.writeChan = make(chan *msg, 100)
	c.closeChan = make(chan struct{})
	c.Logger = entry

	go c.read()
	go c.write()

	return c
}

// OnOpen fn
func (c *Conn) OnOpen() error { return nil }

// OnMessage fn
func (c *Conn) OnMessage(messageType int, p []byte) {}

// OnClose fn
func (c *Conn) OnClose() error { return nil }

func (c *Conn) read() {
	for {
		select {
		case <-c.closeChan:
			return
		default:
			t, p, err := c.conn.ReadMessage()
			if err != nil {
				c.err = err
				c.Close()
				return
			}
			c.readChan <- &msg{messageType: t, p: p}
		}
	}
}

func (c *Conn) write() {
	for msg := range c.writeChan {
		if err := c.conn.WriteMessage(msg.messageType, msg.p); err != nil {
			c.Logger.Error("conn.write err ", err)
		}
	}
}

// Send func
func (c *Conn) Send(messageType int, data []byte) error {
	if c.conn == nil {
		return ErrConnNotInitialized
	}
	
	defer func(){if err:=recover();err!=nil{}}()
	c.writeChan <- &msg{messageType: messageType, p: data}
	return nil
}

func (c *Conn) readHandle() {
	if c.conn == nil {
		c.err = fmt.Errorf("conn is nil")
		return
	}
	for msg := range c.readChan {
		c.dial.OnMessage(msg.messageType, msg.p)
	}
}

// Err fn
func (c *Conn) Err() error {
	return c.err
}

// Close fn
func (c *Conn) Close() error {
	c.Do(func() {
		if c.conn != nil {
			c.conn.Close()
		}
		if c.closeChan != nil {
			close(c.closeChan)
		}
		if c.readChan != nil {
			close(c.readChan)
		}
		if c.writeChan != nil {
			close(c.writeChan)
		}
		if c.dial != nil {
			c.dial.OnClose()
		}
	})
	return nil
}

// NewDialer fn
func NewDialer(addr string, dial Dial, entry ...*logrus.Entry) error {
	if dial == nil {
		return fmt.Errorf("%v", "bot is nil")
	}
	dialer := &websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 10*time.Second)
		},
	}
	conn, _, err := dialer.Dial(addr, http.Header{})
	if err != nil {
		return err
	}

	go dial.init(conn, dial, entry...).readHandle()
	return dial.OnOpen()
}

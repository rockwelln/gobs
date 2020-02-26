package main

import (
	"bufio"
	"bytes"
	"container/list"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

// Promise encapsulate a chan to wait for response & a copy of the original related request
type Promise struct {
	Result  chan BSResponse
	Request string
}

// BSConnection holder of BSFT connection
type BSConnection struct {
	Host          string
	Port          string
	conn          net.Conn
	responseQueue *list.List
	closing       chan bool
	closingFlag   bool
	closed        chan bool
	LastError     error
}

// Session wrapper on the BSConnection for a session
type Session struct {
	conn      *BSConnection
	SessionID string
}

// NewConnection instantiate a new BSConnection
func NewConnection(host string, port int) BSConnection {
	return BSConnection{
		host,
		strconv.Itoa(port),
		nil,
		list.New(),
		make(chan bool, 1),
		false,
		make(chan bool, 1),
		nil,
	}
}

// NewSession start a new session
func NewSession(c *BSConnection) Session {
	rand.Seed(time.Now().UnixNano())
	return Session{
		c, strconv.Itoa(rand.Intn(99999999)),
	}
}

// Connect start a plain-text tcp connection
func (c *BSConnection) Connect() error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(c.Host, c.Port), 2*time.Second)
	if err != nil {
		return err
	}

	c.conn = conn

	go c.read()
	return nil
}

// StartSession starts a new Session on an existing connection
func (c *BSConnection) StartSession(username, password string) (*Session, error) {
	s := NewSession(c)
	p, err := s.SendCommand(NewAuthenticationRequest(username))
	if err != nil {
		return nil, err
	}
	resp := <-p.Result
	v, err := resp.Get("BroadsoftDocument.command.nonce")
	if err != nil {
		return nil, err
	}
	nonce := v.(string)
	p2, err := s.SendCommand(NewLoginRequest(username, password, nonce))
	if err != nil {
		return nil, err
	}
	resp = <-p2.Result
	if resp.IsError() {
		details, _ := resp.GetErrorDetails()
		return nil, fmt.Errorf(details.Summary)
	}
	return &s, nil
}

// SendRaw emits a raw string on the connection and return a promise for the return
func (c *BSConnection) SendRaw(request string) (*Promise, error) {
	if c.closingFlag {
		return nil, fmt.Errorf("Connection closed (last error: %v)", c.LastError)
	}
	pc := make(chan BSResponse, 1)
	p := Promise{pc, request}
	writer := bufio.NewWriter(c.conn)
	_, err := writer.WriteString(request)
	if err != nil {
		return nil, err
	}
	err = writer.Flush()
	if err != nil {
		return nil, err
	}
	c.responseQueue.PushBack(p)
	return &p, nil
}

// Send enclose the command in the OCI-P headers / trailers and emits a string on the connection and return a promise for the return
func (c *BSConnection) Send(request string) (*Promise, error) {
	header := "<?xml version='1.0' encoding='UTF-8'?>\n<BroadsoftDocument xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns=\"C\" protocol=\"OCI\">"
	trailer := "</BroadsoftDocument>\n"
	return c.SendRaw(header + request + trailer)
}

// BSCommand is interface for all command with 'Prepare()' function
type BSCommand interface {
	Prepare() string
}

// SendCommand is handling a command respecting BSCommand interface
func (s *Session) SendCommand(comm BSCommand) (*Promise, error) {
	return s.conn.Send("<sessionId xmlns=\"\">" + s.SessionID + "</sessionId>" + comm.Prepare())
}

// SendMultipleCommands is handling a group of commands as 1 message to BSFT
func (s *Session) SendMultipleCommands(comms []BSCommand) (*Promise, error) {
	commsPrepared := make([]string, len(comms))
	for i, comm := range comms {
		commsPrepared[i] = comm.Prepare()
	}
	return s.conn.Send("<sessionId xmlns=\"\">" + s.SessionID + "</sessionId>" + strings.Join(commsPrepared, ""))
}

func (c *BSConnection) read() {
	stopLoop := false
	incoming := make(chan []byte)
	go func() {
		for {
			reader := bufio.NewReader(c.conn)
			msg := make([]byte, 0)
			for {
				c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				b, err := reader.ReadBytes('\n')
				if err != nil { // timeout (but not only!)
					if nerr, ok := err.(net.Error); ok && nerr.Timeout() { // error is a timeout
						continue
					}
					c.LastError = err
					c.Close(false)
					break
				}
				msg = append(msg[:], b[:]...)
				if bytes.HasSuffix(msg, []byte("</BroadsoftDocument>\n")) {
					break
				}
			}
			incoming <- msg
		}
	}()

	for {
		select {
		case <-c.closing:
			stopLoop = true
		case msg := <-incoming:
			// fmt.Println("received:", string(msg))
			resp, err := NewBSResponse(bytes.TrimSpace(msg))
			if err != nil {
				fmt.Println("response parsing error:", err, "original response: ", msg)
			}
			p := c.responseQueue.Front()
			p.Value.(Promise).Result <- *resp
			c.responseQueue.Remove(p)
		}
		if stopLoop {
			break
		}
	}
	c.closed <- true
}

// Close stops background read task
func (c *BSConnection) Close(wait bool) {
	c.closingFlag = true
	c.closing <- true
	if wait {
		<-c.closed
	}
}

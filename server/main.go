package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type (
	Message struct {
		from    *Client
		payload []byte
	}
	Client struct {
		con  net.Conn
		addr string
	}
	Server struct {
		addr    string
		msgch   chan Message
		clients map[string]*Client
		closech chan os.Signal
		ln      net.Listener
	}
)

func NewServer(addr string) *Server {
	return &Server{
		addr:    addr,
		msgch:   make(chan Message),
		clients: make(map[string]*Client),
		closech: make(chan os.Signal),
	}
}

const MAXBYTE = 1024

var (
	ERR_START = errors.New("could not start the server")
	ERR_SHUT  = errors.New("could not stop the server")
)

func (s *Server) Start() error {
	var err error
	s.ln, err = net.Listen("tcp", s.addr)
	if err != nil {
		return ERR_START
	}

	// TODO:
	// accept and handle messages
	go s.accept()
	go s.handleMessages()

	slog.Info("Starting the server on", "ADDR", s.addr)

	signal.Notify(s.closech, os.Interrupt, syscall.SIGABRT, syscall.SIGTERM)
	<-s.closech
	// TODO: exponential backoff?
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	if err = s.Shutdown(ctx); err != nil {
		return ERR_SHUT
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("Shuttind down...")
	temp := make(chan struct{})
	go func() {
		s.ln.Close()
		close(temp)
	}()

	select {
	case <-temp:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) accept() {
	for {
		con, err := s.ln.Accept()
		if err != nil {
			break
		}
		go s.handleConnection(con)
	}
}

func (s *Server) handleConnection(con net.Conn) {
	// TODO:
	// register user
	// read from the connection
	// send to the messages channel
	client := &Client{
		con:  con,
		addr: con.RemoteAddr().String(),
	}
	s.clients[client.addr] = client
	defer func() {
		delete(s.clients, client.addr)
		if client.con != nil {
			client.con.Close()
			slog.Info("Client disconnected", "addr", client.addr)
			for _, c := range s.clients {
				fmt.Fprintf(c.con, "Room>>%s has disconnected.\n", client.addr)
			}
		}
	}()

	buf := make([]byte, MAXBYTE)

	for {
		n, err := con.Read(buf)
		if err != nil {
			break
		}
		s.msgch <- Message{
			payload: buf[:n],
			from:    client,
		}
	}
}

func (s *Server) handleMessages() {
	// TODO:
	// recieve the messages and redirect the new message to the other users in the same
	// server
	for msg := range s.msgch {
		formatted := fmt.Sprintf("%s", string(msg.payload))
		fmt.Print(formatted)
		for _, client := range s.clients {
			if client.addr != msg.from.addr {
				fmt.Fprint(client.con, formatted)
			}
		}
	}
}

func main() {
	if err := NewServer("127.0.0.1:3000").Start(); err != nil {
		log.Fatal(err)
	}
}

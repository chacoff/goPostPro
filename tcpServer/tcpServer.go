/*
 * File:    tcpServer.go
 * Date:    May 10, 2024
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   TCP server pattern in order to better handle different TCP server operations in within the same software
 *
 */

package tcpServer

import (
	"log"
	"net"

	"goPostPro/global"
)

type Message struct {
	From    string
	Payload []byte
	Conn    net.Conn
}

type Server struct {
	listenAddr string        // address to listen at
	ln         net.Listener  // go listener
	quitch     chan struct{} // empty struct channel for data exchange
	Msgch      chan Message  // main channel for incoming messages
	name       string        // name of the client
}

func NewServer(listenAddr string, name string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		Msgch:      make(chan Message, 10),
		name:       name,
	}
}

func (s *Server) Start() error {

	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln
	log.Printf("[%s] listening on: %s\n", s.name, s.listenAddr)

	go s.acceptLoop()

	<-s.quitch
	close(s.Msgch)

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue // otherwise will not pass more connections
		}

		log.Printf("[%s] new connection from: %s\n", s.name, conn.RemoteAddr())

		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, global.AppParams.MaxBufferSize)
	
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("read error:", err)
			continue
		}

		s.Msgch <- Message{
			From:    conn.RemoteAddr().String(),
			Payload: buf[:n],
			Conn:    conn,
		}

		// conn.Write([]byte("Message ECHO\n"))
	}
}

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
	"encoding/hex"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
	"strings"

	"goPostPro/global"
)

var (
	recordingOn        bool      = false
	recordingOnlyMES bool = true
	timestampRecording time.Time = time.Now()
	recordingFile      *os.File
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

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Println("Timeout read error, retrying:", err)
				time.Sleep(2 * time.Second) // delay to avoid tight loop in temporary errors
				continue
			}

			if err == io.EOF {
				log.Println("Connection closed by the Client")
				break
			}

			log.Println("Read error, closing connection:", err) // non temporary error, closes the connection
			break
		}

		s.Msgch <- Message{
			From:    conn.RemoteAddr().String(),
			Payload: buf[:n],
			Conn:    conn,
		}

		// conn.Write([]byte("Message ECHO\n"))
	}
}

// Start the communication recording, it will only record MES for the moment
func StartRecording() {
	if recordingOn {
		return
	}
	file, err := os.OpenFile(global.Graphics.Savingfolder+"/recording.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	recordingFile = file
	if err != nil {
		log.Println("[RERUN RECORD] : Error ", err)
		return
	}
	recordingOn = true
	recordingOnlyMES = true
	timestampRecording = time.Now()
}

// Change the boolean to record all the communication (DIAS + MES)
func RecordAll(){
	recordingOnlyMES=false
}

// Local function, stop the recording file and save it
func stopRecording(resultFolder string, name string) {
	if !recordingOn {
		return
	}
	err := recordingFile.Close()
	if err != nil {
		log.Println("[RERUN RECORD] : Error ", err)
		return
	}
	recordingOn = false

	saveRecording(resultFolder, name)

}

// Change the recording file
func ChangeRecording(resultFolder string, name string){
	stopRecording(resultFolder, name)
	StartRecording()
}


// Save the recording file in the given folder (type of beam)
func saveRecording(resultFolder string, name string) {

	sourceFile := global.Graphics.Savingfolder + "/recording.txt"
	targetDir := global.Graphics.Savingfolder + "/" + resultFolder
	targetFile := filepath.Join(targetDir, timestampRecording.Format("02-01-06___15h04m05s"))+ "___"+name+ ".txt"

	// Check if folder exists else create it
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		err := os.MkdirAll(targetDir, 0755)
		if err != nil {
			log.Println("[RERUN RECORD] : Error ", err)
			return
		}
		log.Println("[RERUN RECORD] ", targetDir)
	}

	// Move the file in the given folder
	err := os.Rename(sourceFile, targetFile)
	if err != nil {
		log.Println("[RERUN RECORD] : Error ", err)
		return
	}

}

// Function to test if the recording is longer than 5 minutes and stop it if it's the case
func CheckToStopRecording(resultFolder string) {
	diff := time.Since(timestampRecording)

	if diff > 5*time.Minute {
		ChangeRecording(resultFolder, "limit")
	}
}

// Write the payload in the communication recording file using the name of sender given
func WritePayload(payload []byte, sender string) {
	if !recordingOn {
		return
	}
	if recordingOnlyMES && !strings.Contains(sender, "MES") {
		return
	}
	_, err := recordingFile.WriteString("\n" + time.Now().Format("2006-01-02 15:04:05,999") + " | " + sender + " | " + hex.EncodeToString(payload))
	if err != nil {
		log.Println("[RERUN RECORD] : Error ", err)
		return
	}
}

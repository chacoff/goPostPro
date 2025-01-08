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
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"goPostPro/global"
)

var (
	recordingOn        bool      = false
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

func StartRecording() {
	if recordingOn {
		return
	}
	file, err := os.OpenFile(global.Graphics.Savingfolder+"/recording.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	recordingFile = file
	if err != nil {
		log.Println("Error opening file :", err)
		return
	}
	recordingOn = true
	timestampRecording = time.Now()
}

func StopRecording(resultFolder string) {
	if !recordingOn {
		return
	}
	err := recordingFile.Close()
	if err != nil {
		log.Println("Error closing file :", err)
		return
	}
	recordingOn = false

	saveRecording(resultFolder)

}

func saveRecording(resultFolder string) {
	// Chemin du fichier à déplacer
	sourceFile := global.Graphics.Savingfolder + "/recording.txt"
	// Dossier cible
	targetDir := global.Graphics.Savingfolder + "/" + resultFolder

	targetFile := filepath.Join(targetDir, timestampRecording.Format("2006_01_02-15_04_05")) + ".txt" // "a/fichier.txt"
	log.Println(targetFile)

	// Vérifier si le dossier existe, sinon le créer
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		err := os.Mkdir(targetDir, 0755) // Permissions standard pour un dossier
		if err != nil {
			log.Println("Erreur lors de la création du dossier :", err)
			return
		}
		log.Println("Dossier créé :", targetDir)
	}

	// Déplacer le fichier
	err := os.Rename(sourceFile, targetFile)
	if err != nil {
		log.Println("Erreur lors du déplacement du fichier :", err)
		return
	}

	log.Println("Fichier déplacé avec succès dans :", targetFile)
}

func CheckToStopRecording(resultFolder string) {
	diff := time.Since(timestampRecording)

	// Comparer la différence avec 5 minutes
	if diff > 5*time.Minute {
		StopRecording(resultFolder)
	}
}

func WritePayload(payload []byte) {
	if !recordingOn {
		return
	}
	_, err := recordingFile.Write(payload)
	if err != nil {
		log.Println("Erreur lors de l'écriture dans le fichier :", err)
		return
	}
}

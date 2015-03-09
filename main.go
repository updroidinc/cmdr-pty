package main

import "io"
import "log"
import "net/http"
import "os"
import "os/exec"

import "github.com/gorilla/websocket"
import "github.com/kr/pty"

func start() (*exec.Cmd, *os.File) {
	var err error

	cmdString := "/bin/bash"
	cmd := exec.Command(cmdString)
	file, err := pty.Start(cmd)
	if err != nil {
		log.Fatalf("Failed to start command: %s\n", err)
	}

	return cmd, file
}

func stop(pty *os.File, cmd *exec.Cmd) {
	pty.Close()
	cmd.Wait()
}

func ptyHandler(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1,
		WriteBufferSize: 1,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("Websocket upgrade failed: %s\n", err)
	}
	defer conn.Close()

	cmd, file := start()

	// Copy everything from the pty master to the websocket.
	go func() {
		buf := make([]byte, 128)
		// TODO: more graceful exit on socket close / process exit
		for {
			n, err := file.Read(buf)
			if err != nil {
				log.Printf("Failed to read from pty master: %s", err)
				return
			}

			err = conn.WriteMessage(websocket.BinaryMessage, buf[0:n])

			if err != nil {
				log.Printf("Failed to send %d bytes on websocket: %s", n, err)
				return
			}
		}
	}()

	// Read from the websocket, copying to the pty master.
	for {
		mt, payload, err := conn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				log.Printf("conn.ReadMessage failed: %s\n", err)
				return
			}
		}

		switch mt {
		case websocket.BinaryMessage:
			file.Write(payload)
		default:
			log.Printf("Invalid message type %d\n", mt)
			return
		}
	}

	stop(file, cmd)
}

func main() {
	http.HandleFunc("/pty", ptyHandler)

	addr := ":12061"
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("net.http could not listen on address '%s': %s\n", addr, err)
	}
}

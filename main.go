package main

import "io"
import "fmt"
import "net/http"
import "os"
import "os/exec"
import "flag"

import "github.com/gorilla/websocket"
import "github.com/kr/pty"

func start() (*exec.Cmd, *os.File) {
	var err error

	cmdString := "/bin/bash"
	cmd := exec.Command(cmdString)
	file, err := pty.Start(cmd)
	if err != nil {
		fmt.Println("Failed to start command: %s", err)
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
		fmt.Println("Websocket upgrade failed: %s", err)
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
				fmt.Println("Failed to read from pty master: %s", err)
				return
			}

			err = conn.WriteMessage(websocket.BinaryMessage, buf[0:n])

			if err != nil {
				fmt.Println("Failed to send %d bytes on websocket: %s", n, err)
				return
			}
		}
	}()

	// Read from the websocket, copying to the pty master.
	for {
		mt, payload, err := conn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				fmt.Println("conn.ReadMessage failed: %s", err)
				return
			}
		}

		switch mt {
		case websocket.BinaryMessage:
			file.Write(payload)
		default:
			fmt.Println("Invalid message type %d", mt)
			return
		}
	}

	stop(file, cmd)
}

func main() {
	var addrFlag string
	flag.StringVar(&addrFlag, "addr", ":12061", "IP:PORT or :PORT address to listen on")

	flag.Parse()

	http.HandleFunc("/pty", ptyHandler)

	err := http.ListenAndServe(addrFlag, nil)
	if err != nil {
		fmt.Println("net.http could not listen on address '%s': %s", addrFlag, err)
	}
}

package main

import "io"
import "fmt"
import "net/http"
import "os"
import "unicode/utf8"

import "github.com/gorilla/websocket"

// Copy everything from the pty master to the websocket.
func handleOutputWs(ptym *os.File, conn *websocket.Conn) {
	buf := make([]byte, 512)
	var payload, overflow []byte
	// TODO: more graceful exit on socket close / process exit.
	for {
		n, err := ptym.Read(buf)
		if err != nil {
			fmt.Println("failed to read from pty master: ", err)
			return
		}

		// Empty the overflow from the last read into the payload first.
		payload = append(payload[0:], overflow...)
		overflow = nil
		// Then empty the new buf read into the payload.
		payload = append(payload, buf[:n]...)

		// Strip out any incomplete utf-8 from current payload into overflow.
		for !utf8.Valid(payload) {
			overflow = append(overflow[:0], append(payload[len(payload)-1:], overflow[0:]...)...)
			payload = payload[:len(payload)-1]
		}

		// Send out the finished payload as long as it's not empty.
		if len(payload) >= 1 {
			err = conn.WriteMessage(websocket.BinaryMessage, payload[:len(payload)])
			if err != nil {
				fmt.Println("failed to send bytes on websocket: ", err)
				return
			}
		}

		// Empty the payload.
		payload = nil
	}
}

// Read from the websocket, copying to the pty master.
func handleInputWs(ptym *os.File, conn *websocket.Conn) {
	for {
		mt, payload, err := conn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				fmt.Println("conn.ReadMessage failed: ", err)
				return
			}
		}

		// The client has likely disconnected.
		if mt == -1 {
			return
		}

		if mt == websocket.BinaryMessage {
			ptym.Write(payload)
		} else {
			fmt.Println("invalid message type: ", mt)
			return
		}
	}
}

func ptyHandlerWs(w http.ResponseWriter, r *http.Request, sizeFlag string) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1,
		WriteBufferSize: 1,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("websocket upgrade failed: ", err)
	}

	ptySetupWs(conn, sizeFlag)
}

func ptySetupWs(ws *websocket.Conn, sizeFlag string) {
	ptym, cmd := start()
	setPtySize(ptym, sizeFlag)

	go func() {
		handleOutputWs(ptym, ws)
	}()

	go func() {
		handleInputWs(ptym, ws)
	}()

	// Listen for a new winsize on stdin.
	for {
		var newSize string
		_, scanErr := fmt.Scanln(&newSize)
		if scanErr != nil {
			fmt.Println("scan failed: ", scanErr)
		}

		setPtySize(ptym, newSize)
		fmt.Println("new size: ", newSize)
	}

	stop(ptym, cmd)
	ws.Close()
}
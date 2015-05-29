package main

import "io"
import "fmt"
import "net"
import "net/http"
import "log"
import "os"
import "os/exec"
import "flag"
import "strings"
import "strconv"
import "unicode/utf8"

import "github.com/gorilla/websocket"
import "github.com/kr/pty"
import "github.com/creack/goterm/win"

func start() (*os.File, *exec.Cmd) {
	var err error

	cmdString := "/bin/bash"
	cmd := exec.Command(cmdString)
	ptym, err := pty.Start(cmd)
	if err != nil {
		fmt.Println("Failed to start command: ", err)
	}

	return ptym, cmd
}

func stop(ptym *os.File, cmd *exec.Cmd, conn *websocket.Conn) {
	ptym.Close()
	conn.Close()
	cmd.Wait()
}

// Read from the websocket, copying to the pty master.
func handleInput(ptym *os.File, conn *websocket.Conn) {
	for {
		mt, payload, err := conn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				fmt.Println("conn.ReadMessage failed: ", err)
				return
			}
		}

		if mt == -1 {
			// The client has likely disconnected.
			return
		}

		switch mt {
		case websocket.BinaryMessage:
			ptym.Write(payload)
		default:
			fmt.Println("Invalid message type: ", mt)
			return
		}
	}
}

// Copy everything from the pty master to the websocket.
func handleOutput(ptym *os.File, conn *websocket.Conn) {
	buf := make([]byte, 512)
	var payload, overflow []byte
	// TODO: more graceful exit on socket close / process exit.
	for {
		n, err := ptym.Read(buf)
		if err != nil {
			fmt.Println("Failed to read from pty master: ", err)
			return
		}

		// Empty the overflow from the last read into the payload first.
		payload = append(payload[0:], overflow...)
		overflow = nil
		// Then empty the new buf read into the payload.
		payload = append(payload, buf[:n]...)

		// Strip out any incomplete utf-8 from current payload into overflow.
		for !utf8.Valid(payload) {
			overflow = append(overflow, payload[len(payload)-1:len(payload)]...)
			payload = payload[:len(payload)-1]
		}

		// Send out the finished payload.
		err = conn.WriteMessage(websocket.BinaryMessage, payload[:len(payload)])
		if err != nil {
			fmt.Println("Failed to send bytes on websocket: ", err)
			return
		}

		// Empty the payload.
		payload = nil
	}
}

func setPtySize(ptym *os.File, size string) {
	sizeArr := strings.Split(size, "x")
	lines, _ := strconv.Atoi(sizeArr[0])
	cols, _ := strconv.Atoi(sizeArr[1])
	if err := win.SetWinsize(ptym.Fd(), &win.Winsize{Height: uint16(lines), Width: uint16(cols)}); err != nil {
		panic(err)
	}
}

func ptyHandler(w http.ResponseWriter, r *http.Request, sizeFlag string) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1,
		WriteBufferSize: 1,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Websocket upgrade failed: ", err)
	}

	ptym, cmd := start()
	setPtySize(ptym, sizeFlag)

	go func() {
		handleOutput(ptym, conn)
	}()

	go func() {
		handleInput(ptym, conn)
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

	stop(ptym, cmd, conn)
}

func main() {
	addrFlag := flag.String("addr", ":0", "IP:PORT or :PORT address to listen on")
	sizeFlag := flag.String("size", "24x80", "initial size for the tty")

	flag.Parse()

	http.HandleFunc("/pty", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("client connected.")
		ptyHandler(w, r, *sizeFlag)
	})

	listener, err := net.Listen("tcp", *addrFlag)
	if err != nil {
		log.Fatal("listen: %s", err)
	}

	fmt.Println("now listening on: " + listener.Addr().String())

	err = http.Serve(listener, nil)
	if err != nil {
		fmt.Println("net.http could not listen on address '%s': %s", addrFlag, err)
	}
}

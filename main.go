package main

import "io"
import "fmt"
import "net/http"
import "os"
import "os/exec"
import "flag"
import "strings"
import "strconv"

import "github.com/gorilla/websocket"
import "github.com/kr/pty"
import "github.com/creack/goterm/win"

func start() (*exec.Cmd, *os.File) {
	var err error

	cmdString := "/bin/bash"
	cmd := exec.Command(cmdString)
	f, err := pty.Start(cmd)
	if err != nil {
		fmt.Println("Failed to start command: %s", err)
	}

	return cmd, f
}

func stop(pty *os.File, cmd *exec.Cmd) {
	pty.Close()
	cmd.Wait()
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
		fmt.Println("Websocket upgrade failed: %s", err)
	}
	defer conn.Close()

	cmd, file := start()

	size := strings.Split(sizeFlag, "x")
	x, _ := strconv.Atoi(size[0])
	y, _ := strconv.Atoi(size[1])
	if err := win.SetWinsize(file.Fd(), &win.Winsize{Height: uint16(x), Width: uint16(y)}); err != nil {
        panic(err)
    }

	// Copy everything from the pty master to the websocket.
	go func() {
		buf := make([]byte, 256)
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
	addrFlag := flag.String("addr", ":12061", "IP:PORT or :PORT address to listen on")
	sizeFlag := flag.String("size", "80x24", "initial size for the tty")

	flag.Parse()

	http.HandleFunc("/pty", func(w http.ResponseWriter, r *http.Request) {
              ptyHandler(w, r, *sizeFlag)
       })

	err := http.ListenAndServe(*addrFlag, nil)
	if err != nil {
		fmt.Println("net.http could not listen on address '%s': %s", addrFlag, err)
	}
}
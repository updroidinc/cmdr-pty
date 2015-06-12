package main

import "flag"
import "fmt"
import "net"
import "net/http"
import "os"
import "os/exec"
import "strings"
import "strconv"

import "github.com/kr/pty"
import "github.com/creack/goterm/win"

func setPtySize(ptym *os.File, size string) {
	sizeArr := strings.Split(size, "x")
	lines, _ := strconv.Atoi(sizeArr[0])
	cols, _ := strconv.Atoi(sizeArr[1])
	if err := win.SetWinsize(ptym.Fd(), &win.Winsize{Height: uint16(lines), Width: uint16(cols)}); err != nil {
		panic(err)
	}
}

func start() (*os.File, *exec.Cmd) {
	var err error

	cmdString := "/bin/bash"
	cmd := exec.Command(cmdString)
	ptym, err := pty.Start(cmd)
	if err != nil {
		fmt.Println("failed to start command: ", err)
	}

	return ptym, cmd
}

func stop(ptym *os.File, cmd *exec.Cmd) {
	ptym.Close()
	cmd.Wait()
}


func main() {
	protocolFlag := flag.String("protocol", "websocket", "specify websocket or tcp")
	addrFlag := flag.String("addr", ":0", "IP:PORT or :PORT address to listen on")
	sizeFlag := flag.String("size", "24x80", "initial size for the tty")

	flag.Parse()

	if *protocolFlag == "websocket" {
		http.HandleFunc("/pty", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("client connected.")
			ptyHandlerWs(w, r, *sizeFlag)
		})

		listener, err := net.Listen("tcp", *addrFlag)
		if err != nil {
			fmt.Println("listen error: ", err)
		}

		fmt.Println("now listening on: ", listener.Addr().String())

		err = http.Serve(listener, nil)
		if err != nil {
			fmt.Printf("net.http could not listen on address '%s': %s\n", addrFlag, err)
		}
	} else {
		addr, err := net.ResolveTCPAddr("tcp", *addrFlag)
		if err != nil {
	        fmt.Println("resolve error", err)
	        return
	    }

		listener, err := net.ListenTCP("tcp", addr)
	    if err != nil {
	        fmt.Println("listen error", err)
	        return
	    }

	    fmt.Println("now listening on: ", listener.Addr().String())

	    for {
	        conn, err := listener.AcceptTCP()
	        if err != nil {
	            fmt.Println("accept error", err)
	            return
	        }

	        fmt.Println("client connected.")

	        go ptySetupSock(conn, *sizeFlag)
	    }
	}
}

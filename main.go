package main

import "fmt"
import "net"
import "net/http"
import "os"
import "os/exec"
import "strings"
import "strconv"

import "github.com/kr/pty"
import "github.com/creack/goterm/win"
import "gopkg.in/alecthomas/kingpin.v2"

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
	protocolFlag := kingpin.Flag("protocol", "specify websocket or tcp").Short('p').Default("websocket").String()
	addrFlag := kingpin.Flag("addr", "IP:PORT or :PORT address to listen on").Short('a').Default(":0").String()
	sizeFlag := kingpin.Flag("size", "initial size for the tty").Short('s').Default("25x80").String()

	kingpin.Parse()

	if *protocolFlag == "websocket" {
		http.HandleFunc("/pty", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("client connected.")
			ptyHandlerWs(w, r, *sizeFlag)
		})

		listener, err := net.Listen("tcp", *addrFlag)
		if err != nil {
			fmt.Println("listen error: ", err)
		}

		_, port, _ := net.SplitHostPort(listener.Addr().String())
		fmt.Println("listening on port:", port)

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

		_, port, _ := net.SplitHostPort(listener.Addr().String())
		fmt.Println("listening on port:", port)

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

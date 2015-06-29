package main

import "fmt"
import "net"
import "os"

import "unicode/utf8"

// Copy everything from the pty master to the socket.
func handleOutputSock(ptym *os.File, conn *net.TCPConn) {
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
			_, err = conn.Write(payload)
	        if err != nil {
	            fmt.Println("Write: ", err)
	        }
		}

		// Empty the payload.
		payload = nil
	}
}

// Read from the websocket, copying to the pty master.
func handleInputSock(ptym *os.File, conn *net.TCPConn) {
	for {
        buf := make([]byte, 512)
        bytes, err := conn.Read(buf)
        if err != nil {
            return
        }

        buf = buf[:bytes]
		ptym.Write(buf)
	}
}

func ptySetupSock(conn *net.TCPConn, sizeFlag string) {
	ptym, cmd := start()
	setPtySize(ptym, sizeFlag)

	go func() {
		handleOutputSock(ptym, conn)
	}()

	go func() {
		handleInputSock(ptym, conn)
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
	conn.Close()
}
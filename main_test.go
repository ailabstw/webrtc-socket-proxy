package main

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestSmoke(t *testing.T) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	go newEchoServer(":9064")

	signalServerAddr := "ws://localhost:8000/connection/websocket"

	go NewAs("smoke", "secret", ":9064", signalServerAddr)

	time.Sleep(1 * time.Second)

	to := NewTo("smoke", "secret", signalServerAddr, ":9063")
	go to.ListenAndServe()

	time.Sleep(1 * time.Second)

	receiveCount := 0

	go func() {
		addr, err := net.ResolveTCPAddr("tcp", ":9063")
		if err != nil {
			panic(err)
		}
		conn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			panic(err)
		}

		log.Print("connected")

		// read from socket
		go func() {
			for {
				data := make([]byte, 1024)
				n, err := conn.Read(data)
				if err != nil {
					panic(err)
				}
				resp := string(data[:n])
				log.Printf("echo received: %#v", resp)

				if resp == "smoke test\n" {
					receiveCount++
				} else {
					t.Errorf("incorrect echo data %#v", data)
				}

				if receiveCount == 3 {
					os.Exit(0)
				}
			}
		}()
		for i := 0; i < 3; i++ {
			if _, err := conn.Write([]byte("smoke test\n")); err != nil {
				panic(err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	for {
	}
}

func newEchoServer(port string) {
	l, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}
	log.Print("Listening to connections on :9064")
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	log.Print("Accepted new connection.")
	defer conn.Close()
	defer log.Print("Closed connection.")

	for {
		buf := make([]byte, 1024)
		size, err := conn.Read(buf)
		if err != nil {
			return
		}
		data := buf[:size]
		log.Printf("Read new data from connection %#v", string(data))
		conn.Write(data)
	}
}

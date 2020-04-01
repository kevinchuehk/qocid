package main

import (
	"net"
	"io"
	"log"
	"os/signal"
	"io/ioutil"
	"os"
	"fmt"
)

var closeFlag = false

func handleConnection(conn net.Conn) {
	// Create unix domain socket connection
	log.Println("conn in...")
	sockConn, err := net.Dial("unix", sock)
	if err != nil {
		log.Println(err)
		return 
	}
	defer sockConn.Close()
	defer conn.Close()
	
	// r, w := io.Pipe()
	b, err := ioutil.ReadAll(conn)
	fmt.Printf("%s",b)

	io.Copy(sockConn, conn)
	io.Copy(conn, sockConn)
}

func proxyShutdown(ch <-chan os.Signal, ln net.Listener) {
	for {
		select {
		case <-ch:
			closeFlag = true
			ln.Close()
			break
		default:
			continue
		}
	}
}

func proxyServe()  {
	ln, err := net.Listen("tcp", ":2375")
	if err != nil {
		log.Println(err)
		return 
	}
	log.Println("server up...")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	go proxyShutdown(ch, ln)

	for {
		conn, err:= ln.Accept()
		if err != nil {
			if closeFlag {
				break
			}

			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}

	log.Println("server down...")
}


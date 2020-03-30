package main

import (
	"net"
	"io"
	"log"
	"os/signal"
	"os"
)

var closeFlag = false

func handleConnection(conn net.Conn) {
	// Create unix domain socket connection
	sockConn, err := net.Dial("unix", sock)
	if err != nil {
		log.Println(err)
		return 
	}
	defer sockConn.Close()

	go io.Copy(conn, sockConn)
	go io.Copy(sockConn, conn)
}

func proxyShutdown(ch <-chan os.Signal, ln net.Listener) {
	<-ch
	closeFlag = true
	ln.Close()
}

func proxyServe()  {
	ln, err := net.Listen("tcp", ":2375")
	if err != nil {
		log.Println(err)
		return 
	}

	ch := make(chan os.Signal)
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
}


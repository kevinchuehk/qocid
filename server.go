package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"log"
	"os"
	"os/signal"
	"fmt"
)

func proxyShutdown(ch <-chan os.Signal, srv http.Server) {
	for {
		select {
		case <-ch:
			err := srv.Shutdown(context.Background())
			if err != nil {
				log.Printf("server shutdown: %v", err)
			}
			break
		default:
			continue
		}
	}
}

func proxyServe()  {
	trueAddr := fmt.Sprint("unix://", sock)
	url, err := url.Parse(trueAddr)
	if err != nil {
		log.Println(err)
		return
	}

	proxyHandler := httputil.NewSingleHostReverseProxy(url)
	srv := http.Server {
		Addr: ":2375",
		Handler: proxyHandler,
	}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal("HTTP server ListenAndServer: %v", err)
	}
	log.Println("server up...")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	go proxyShutdown(ch, srv)

	log.Println("server down...")
}


package main

import (
	"os"
	"os/exec"
	"os/signal"
	"fmt"
	"log"
)

const (
	tmp = "/var/tmp"
	lib = "/var/lib/containers"
	run = "/var/run/containers"
	sock = "/var/run/containers/docker.sock"
)

func config() {
	err := os.MkdirAll(tmp, os.ModeDir)
	if err != nil {
		log.Println(err)
	}

	err = os.MkdirAll(lib, os.ModeDir)
	if err != nil {
		log.Println(err)
	}

	err = os.MkdirAll(run, os.ModeDir)
	if err != nil {
		log.Println(err)
	}
}

func listenForShutdown(ch <-chan os.Signal, cmd exec.Cmd) {
	for {
		switch sig := <-ch {
		case os.Interrupt:
		case os.Kill:
			cmd.Process.Kill()
			log.Println("container runtime closed...")
			break
		default:
			continue
		}
	}
	
	defer os.Remove(sock)
}

func main() {
	// Remove sock file
	os.Remove(sock)
	config()
	
	args := []string {
		// flags
		fmt.Sprint("--root ", lib),
		fmt.Sprint("--runroot ", run),
		fmt.Sprint("--tmpdir ", tmp),
		fmt.Sprint("--runtime ", "/bin/runc"),
		
		// subcommand
		fmt.Sprint("system service ", "-t 0 ", "unix://",sock),
	}

	cmd := exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Println("container runtime started...")
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	go listenForShutdown(ch, *cmd)
	proxyServe()
}
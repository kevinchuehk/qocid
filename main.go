package main

import (
	"os"
	"os/exec"
	"os/signal"
	"fmt"
	"log"
)

const (
	tmp = "/var/run/libpod"
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
		select {
		case <-ch:
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

	runCmd := fmt.Sprint(
		"podman system service",
        " -t 0 ", "unix://", sock,
		" --root ", lib,
		" --runtime ", "/bin/runc ",
	)
	cmd := exec.Command("sh", "-c", runCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Println("container runtime started...")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	listenForShutdown(ch, *cmd)
	
	log.Println("container runtime shutdown...")
}

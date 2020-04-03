package main

import (
	"os"
	"os/exec"
	"os/signal"
	"fmt"
	"log"
	"io/ioutil"
)

const (
	tmp = "/var/run/libpod"
	lib = "/var/lib/containers"
	run = "/var/run/containers"
	sock = "/var/run/containers/docker.sock"
	registries = "/etc/containers/register.conf"
	policy = "/etc/containers/policy.json"
	libpod = "/etc/containers/libpod.conf"
	cniConf = "/etc/cni/net.d/87-podman-bridge.conf"
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

	snapEnv := os.Getenv("SNAP")
	if snapEnv == "" { return }

	if _, err := os.Stat(registries); os.IsNotExist(err) {
		data, _ := ioutil.ReadFile(
			fmt.Sprint(snapEnv, registries),
		)

		if err := ioutil.WriteFile(registries, data, 0644); err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(policy); os.IsNotExist(err) {
		data, _ := ioutil.ReadFile(
			fmt.Sprint(snapEnv, policy),
		)

		if err := ioutil.WriteFile(policy, data, 0644); err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(libpod); os.IsNotExist(err) {
		data, _ := ioutil.ReadFile(
			fmt.Sprint(snapEnv, libpod),
		)

		if err := ioutil.WriteFile(libpod, data, 0644); err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(cniConf); os.IsNotExist(err) {
		data, _ := ioutil.ReadFile(
			fmt.Sprint(snapEnv, cniConf),
		)

		if err := ioutil.WriteFile(cniConf, data, 0644); err != nil {
			log.Fatal(err)
		}
	}
}

func listenForShutdown(ch <-chan os.Signal, cmd exec.Cmd) {
	for {
		select {
		case <-ch:
			cmd.Process.Kill()
			log.Println("container runtime closed...")
			os.Exit(0)
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
	snapEnv := os.Getenv("SNAP")

	runCmd := fmt.Sprint(
		"podman system service",
        " -t 0 ", "unix://", sock,
		" --root ", lib,
		" --runtime ", "/bin/runc ",
		" --conmon ", fmt.Sprint(snapEnv, "/libexec/podman/conmon"),
	)

	cmd := exec.Command("sh", "-c", runCmd)
	env := fmt.Sprint(
		"PATH=",
		fmt.Sprint(snapEnv, "/usr/sbin"),
		fmt.Sprint(":", snapEnv, "/usr/bin"),
		fmt.Sprint(":", snapEnv, "/sbin"),
		fmt.Sprint(":", snapEnv, "/bin"),
	)
	
	LDEnv := fmt.Sprint(
		"LD_LIBRARY_PATH=",
		fmt.Sprint(snapEnv, "/lib"),
		fmt.Sprint(":", snapEnv, "/usr/lib"),
	) 

	cmd.Env = append( os.Environ(), env, LDEnv)
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

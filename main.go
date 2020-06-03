package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	SYS "syscall"

	DEATH "github.com/vrecan/death"
)

const (
	tmp        = "/var/run/libpod"
	lib        = "/var/lib/containers"
	run        = "/var/run/containers"
	sock       = "/var/run/containers/docker.sock"
	registries = "/etc/containers/register.conf"
	policy     = "/etc/containers/policy.json"
	cniConf    = "/etc/cni/net.d/87-podman-bridge.conf"
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

	err = os.MkdirAll("/etc/containers", os.ModeDir)
	if err != nil {
		log.Println(err)
	}

	err = os.MkdirAll("/etc/cni/net.d", os.ModeDir)
	if err != nil {
		log.Println(err)
	}

	snapEnv := os.Getenv("SNAP")
	if snapEnv == "" {
		return
	}

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

	if _, err := os.Stat(cniConf); os.IsNotExist(err) {
		data, _ := ioutil.ReadFile(
			fmt.Sprint(snapEnv, cniConf),
		)

		if err := ioutil.WriteFile(cniConf, data, 0644); err != nil {
			log.Fatal(err)
		}
	}
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
		os.Getenv("PATH"),
		fmt.Sprint(":", snapEnv, "/usr/sbin"),
		fmt.Sprint(":", snapEnv, "/usr/bin"),
		fmt.Sprint(":", snapEnv, "/sbin"),
		fmt.Sprint(":", snapEnv, "/bin"),
		fmt.Sprint(":", snapEnv, "/opt/cni/bin"),
	)

	LDEnv := fmt.Sprint(
		"LD_LIBRARY_PATH=",
		fmt.Sprint(snapEnv, "/lib"),
		fmt.Sprint(":", snapEnv, "/usr/lib"),
	)

	cmd.Env = append(os.Environ(), env, LDEnv)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Println("container runtime started...")
	death := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM)
	death.WaitForDeathWithFunc(func() {
		cmd.Process.Kill()
		log.Println("container runtime closed...")
		os.Remove(sock)
		os.Exit(0)
	})

	log.Println("container runtime shutdown...")
}

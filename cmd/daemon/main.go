package main

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"ops-ctrl/pkg/manager"
)

var mgr = manager.NewManager()

func verifyAction(err error, message string) map[string]string {
	if err != nil {
		return map[string]string{"status": "error", "message": err.Error()}
	} else {
		return map[string]string{"status": "success", "message": message}
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	var request map[string]string
	decoder := json.NewDecoder(conn)
	err := decoder.Decode(&request)
	if err != nil {
		log.Println("Failed to decode request:", err)
		return
	}

	action := request["action"]
	name := request["name"]
	command := request["command"]

	workingDir := "/"
	serviceMode := "binary_argument"

	var response map[string]string

	switch action {
	case "start":
		args := []string{""}
		envs := []string{""}
		mgr.AddService(name, command, args, envs, workingDir, serviceMode)
		err := mgr.StartService(name)
		response = verifyAction(err, "Service "+name+" started")
	case "stop":
		err := mgr.StopService(name)
		response = verifyAction(err, "Service "+name+" stopped")
	case "status":
		status := mgr.ServiceStatus(name)
		response = map[string]string{"status": "success", "message": status}
	case "firefox":
		args := []string{"github.com/viirret/ops-ctrl"}
		envs := []string{"DISPLAY=:0"}
		mgr.AddService(name, command, args, envs, workingDir, serviceMode)
		err := mgr.StartService(name)
		response = verifyAction(err, "Firefox started")
	default:
		response = map[string]string{"status": "error", "message": "Unknown action"}
	}

	encoder := json.NewEncoder(conn)
	err = encoder.Encode(response)
	if err != nil {
		log.Println("Failed to encode response:", err)
	}
}

func main() {
	listener, err := net.Listen("unix", "/tmp/ops-ctrl-daemon.sock")
	if err != nil {
		log.Fatal("Failed to listen on socket:", err)
	}
	defer listener.Close()
	log.Println("Service manager daemon started")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Println("Shutting down service manager daemon")
		listener.Close()
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

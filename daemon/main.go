package main

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"ops-ctrl/pkg/manager"
)

var aliases = map[string]string{
	"firefox":  "/usr/bin/firefox",
	"chromium": "/usr/bin/chromium",
}

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

	// The first argument, "start", "stop", etc.
	action := request["action"]

	// Own identifier
	id := request["id"]

	// Arguments to the program binary
	arg0 := request["arg0"]
	//arg1 := request["arg1"]
	//arg2 := request["arg2"]

	workingDir := "/"
	var response map[string]string

	switch action {
	case "start":
		args := []string{arg0}
		envs := []string{"DISPLAY=:0"}

		binary, binaryExists := request["binary"]

		if binaryExists {
			log.Println("Binary argument exists!")
			mgr.AddService(id, binary, args, envs, workingDir)
		}

		alias, aliasExists := request["alias"]

		if aliasExists {
			if familiarAlias, familiarAliasesExist := aliases[alias]; familiarAliasesExist {
				log.Println("Found defined alias:->", familiarAlias)
			} else {
				log.Println("Aliases not found for:", alias)
			}
			log.Println("Alias argument exists!")
			mgr.AddService(id, alias, args, envs, workingDir)
		}

		err := mgr.StartService(id)
		pid := strconv.Itoa(mgr.GetPID(id))
		response = verifyAction(err, "Service "+id+" started with pid: "+pid)
	case "stop":
		err := mgr.StopService(id)
		response = verifyAction(err, "Service "+id+" stopped")
	case "status":
		status := mgr.ServiceStatus(id)
		response = map[string]string{"status": "success", "message": status}
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

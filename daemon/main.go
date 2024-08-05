package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"ops-ctrl/pkg/config"
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

	// The first argument, "start", "stop", etc.
	action := request["action"]

	// Arguments to the program binary
	arg0 := request["arg0"]
	arg1 := request["arg1"]
	args := []string{arg0, arg1}

	// Print each argument using a loop
	for index, arg := range args {
		fmt.Printf("args[%d]: %s\n", index, arg)
	}

	workingDir := "/"
	dir, dirExists := request["working_dir"]
	if dirExists {
		log.Println("Workingdir argument exists: ", dir)
		workingDir = dir
	} else {
		log.Println("Running default working dir: ", workingDir)
	}

	var response map[string]string

	switch action {
	case "start":
		id := request["id"]
		envs := []string{"DISPLAY=:0"}

		binary, binaryExists := request["binary"]

		if binaryExists {
			log.Println("Binary argument exists!")
			mgr.AddService(id, binary, args, envs, workingDir)
		}

		alias, aliasExists := request["alias"]

		if aliasExists {
			tomlFile := "config.toml"
			aliases, err := config.LoadAliases(tomlFile)

			if err != nil {
				log.Fatal("Error loading aliases: ", err)
			}

			if familiarAlias, familiarAliasesExist := aliases[alias]; familiarAliasesExist {
				log.Println("Found defined alias:->", familiarAlias)
				mgr.AddService(id, alias, args, envs, workingDir)
			} else {
				log.Fatal("Aliases not found for:", alias)
				return
			}
		}
		// Start service
		err := mgr.StartService(id)
		pid := strconv.Itoa(mgr.GetPID(id))
		response = verifyAction(err, "Service "+id+" started with pid: "+pid)

	// Stop service
	case "stop":
		pid, pidExists := request["pid"]
		if pidExists {
			log.Println("PID argument exits: ", pid)
			pidValue, pidErr := strconv.Atoi(pid)
			if pidErr != nil {
				log.Println("Error with int conversion: ", pidErr)
				return
			}
			err := mgr.StopServiceWithPID(pidValue)
			response = verifyAction(err, "Service pid "+pid+" stopped")
			break
		}

		id, idExists := request["id"]
		if idExists {
			log.Println("ID argument exists: ", id)
			err := mgr.StopServiceWithID(id)
			response = verifyAction(err, "Service "+id+" stopped")
			break
		}
		log.Fatal("No method for finding program found")
		return

	// Check status of the service
	case "status":
		id, idExists := request["id"]
		if idExists {
			log.Println("ID argument exists: ", id)
			status := mgr.ServiceStatusByID(id)
			response = map[string]string{"status": "success", "message": status}
			break
		}
		pid, pidExists := request["pid"]
		if pidExists {
			log.Println("PID argument exists: ", pid)
			pidValue, pidErr := strconv.Atoi(pid)
			if pidErr != nil {
				fmt.Println("Error with string to int conversion: ", pidErr)
				return
			}
			status := mgr.ServiceStatusByPID(pidValue)
			response = map[string]string{"status": "success", "message": status}
			break
		}
		log.Fatal("No method for finding program found")
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

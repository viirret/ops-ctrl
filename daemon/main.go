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

func verifyAction(err error, message string) map[string]interface{} {
	if err != nil {
		return map[string]interface{}{"status": "error", "message": err.Error()}
	} else {
		return map[string]interface{}{"status": "success", "message": message}
	}
}

func argumentArrayValue[T any](request map[string]interface{}, argumentType string) []T {
	item, itemOk := request[argumentType].([]interface{})
	itemValues := []T{}

	if itemOk {
		newItemValues := make([]T, len(item))

		// Convert each interface{} to string
		for index, value := range item {
			if envStr, ok := value.(T); ok {
				newItemValues[index] = envStr
			} else {
				log.Fatalln("Warning: Unexpected type in env slice: ", value)
			}
		}
		itemValues = newItemValues

		for i, e := range itemValues {
			fmt.Printf("%s [%d]: %v\n", argumentType, i, e)
		}
	}
	return itemValues
}

func argumentValue[T any](request map[string]interface{}, argumentType string, defaultValue T) T {
	if item, itemOk := request[argumentType].(T); itemOk {
		if str, isString := any(item).(string); isString && str == "" {
			fmt.Printf("Found empty string for: %s, using default value\n", argumentType)
			return defaultValue
		}
		fmt.Printf("Found argument for: %s\n", argumentType)
		return item
	}
	fmt.Printf("Using default value for: %s\n", argumentType)
	return defaultValue
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	var request map[string]interface{}
	decoder := json.NewDecoder(conn)
	err := decoder.Decode(&request)
	if err != nil {
		log.Fatalln("Failed to decode request:", err)
		return
	}
	log.Println(request)

	// The first argument, "start", "stop", etc.
	action := request["action"].(string)

	// Environment variables
	envStrings := argumentArrayValue[string](request, "env")

	// Arguments for the program binary
	argStrings := argumentArrayValue[string](request, "program_argument")

	workingDir := argumentValue(request, "working_dir", "/")
	var response = make(map[string]interface{})

	switch action {
	case "start":
		id := argumentValue(request, "id", mgr.RandomID(10))

		binary := argumentValue(request, "binary", "")
		if binary != "" {
			mgr.AddService(id, binary, argStrings, envStrings, workingDir)
		}

		alias, aliasExists := request["alias"].(string)

		if aliasExists {
			cfg := config.GetConfig()

			if familiarAlias, familiarAliasesExist := cfg.Aliases[alias]; familiarAliasesExist {
				log.Println("Found defined alias:->", familiarAlias)
				mgr.AddService(id, alias, argStrings, envStrings, workingDir)
			} else {
				log.Fatal("Aliases not found for:", alias)
				return
			}
		}

		if binary == "" && !aliasExists {
			log.Fatal("Program binary undefined!")
			return
		}

		// Start service
		err := mgr.StartService(id)
		pid := strconv.Itoa(mgr.GetPID(id))
		response = verifyAction(err, "Service "+id+" started with pid: "+pid)

	// Stop service
	case "stop":
		pid, pidExists := request["pid"].(string)
		if pidExists {
			log.Println("PID argument exits: ", pid)
			pidValue, pidErr := strconv.Atoi(pid)

			if pidErr != nil {
				log.Println("Error with int conversion: ", pidErr)
				return
			}
			err := mgr.StopServiceWithPID(pidValue)
			message := "Service with PID:" + pid + "and ID:" + mgr.GetID(pidValue) + " stopped"
			response = verifyAction(err, message)
			break
		}

		id, idExists := request["id"].(string)
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
		id, idExists := request["id"].(string)
		if idExists {
			log.Println("ID argument exists: ", id)
			status := mgr.ServiceStatusByID(id)
			response = map[string]interface{}{"status": "success", "message": status}
			break
		}
		pid, pidExists := request["pid"].(string)
		if pidExists {
			log.Println("PID argument exists: ", pid)
			pidValue, pidErr := strconv.Atoi(pid)
			if pidErr != nil {
				fmt.Println("Error with string to int conversion: ", pidErr)
				return
			}
			status := mgr.ServiceStatusByPID(pidValue)
			response = map[string]interface{}{"status": "success", "message": status}
			break
		}
		log.Fatal("No method for finding program found")
	default:
		response = map[string]interface{}{"status": "error", "message": "Unknown action"}
	}

	encoder := json.NewEncoder(conn)
	err = encoder.Encode(response)
	if err != nil {
		log.Println("Failed to encode response:", err)
	}
}

func main() {
	tomlFile := "config.toml"
	config.LoadConfig(tomlFile)

	listener, err := net.Listen("unix", "/tmp/ops-ctrl-daemon.sock")
	if err != nil {
		log.Fatal("Failed to listen on socket:", err)
	}
	defer listener.Close()
	log.Println("Service manager daemon started")

	mgr.RunAutostart()

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

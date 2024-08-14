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
	"ops-ctrl/pkg/service"
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
				log.Fatalf("Warning: Unexpected type in env slice: %s", value)
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
	fmt.Println(request)

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
				fmt.Printf("Found defined alias:->%s", familiarAlias)
				mgr.AddService(id, alias, argStrings, envStrings, workingDir)
			} else {
				log.Fatalln("Aliases not found for:", alias)
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
		response = verifyAction(err, "Service "+id+" started with pid: "+pid+"\n")

	// Stop service
	case "signal":
		signalType := argumentValue(request, "signalType", "")
		if signalType == "" {
			log.Fatal("Missing signal type")
			return
		}

		signal, err := service.GetSignal(signalType)
		if err != nil {
			log.Fatal(err)
			return
		}

		pidFloat, pidFloatExists := request["pid"].(float64)
		if pidFloatExists {
			intVal := int(pidFloat)
			err := mgr.SignalServiceWithPID(intVal, signal)
			message := "Service with PID:" + string(int(pidFloat)) + "and ID:" + mgr.GetID(intVal) + " stopped"
			response = verifyAction(err, message)
			break
		}

		id, idExists := request["id"].(string)
		if idExists {
			log.Println("ID argument exists: ", id)
			err := mgr.SignalServiceWithID(id, signal)
			response = verifyAction(err, "Service "+id+" stopped")
			break
		}
		log.Fatal("No method for finding program found")
		return

	// Check status of the service
	case "status":
		id, idExists := request["id"].(string)
		if idExists {
			fmt.Printf("ID argument exists: %s", id)
			status := mgr.ServiceStatusByID(id)
			response = map[string]interface{}{"status": "success", "message": status}
			break
		}
		pid, pidExists := request["pid"].(float64)
		if pidExists {
			status := mgr.ServiceStatusByPID(int(pid))
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
		log.Fatal("Failed to encode response: ", err)
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
	fmt.Print("Service manager daemon started\n")

	mgr.RunAutostart()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		fmt.Print("Shutting down service manager daemon")
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

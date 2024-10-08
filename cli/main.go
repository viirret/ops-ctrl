package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"ops-ctrl/pkg/service"
)

func sendRequest(request map[string]interface{}) {
	conn, err := net.Dial("unix", "/tmp/ops-ctrl-daemon.sock")
	if err != nil {
		log.Fatal("Failed to connect to daemon:", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	err = encoder.Encode(request)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	var response map[string]interface{}
	decoder := json.NewDecoder(conn)
	err = decoder.Decode(&response)
	if err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	fmt.Printf("Response:%s\n", response["message"])
}

func addArguments(arguments map[service.Argument]interface{}, request map[string]interface{}) {
	for key, value := range arguments {
		if key.SupportsArrays() {
			// TODO add multiple arguments with other values than "string".
			itemValues := strings.Split(value.(string), ",")
			for i, item := range itemValues {
				fmt.Printf("%s [%d] %s\n", key, i, item)
			}
			request[string(key)] = itemValues
			continue
		}

		switch key {
		case service.PID:
			strValue, ok := value.(string)
			if !ok {
				fmt.Println("Value is not a string:", value)
			} else if intValue, err := strconv.Atoi(strValue); err == nil {
				request[string(key)] = intValue
			} else {
				fmt.Println("Int conversion failed:", value)
			}
		default:
			request[string(key)] = value
		}
	}
}

func main() {
	first_argument := os.Args[1]
	argumentsAfterAction := os.Args[2:]

	switch first_argument {
	case "start":
		request := map[string]interface{}{
			"action":           "start",
			"id":               "",
			"env":              []string{},
			"program_argument": []string{},
		}
		validArgs := service.CheckArguments(argumentsAfterAction)

		addArguments(validArgs, request)
		sendRequest(request)
	case "signal":
		// Reserve first argument to the signal type
		signalString := argumentsAfterAction[0]
		validArgs := service.CheckArguments(argumentsAfterAction[1:])
		request := map[string]interface{}{"action": "signal", "signalType": signalString}

		addArguments(validArgs, request)
		sendRequest(request)
	case "status":
		validArgs := service.CheckArguments(argumentsAfterAction)
		request := map[string]interface{}{"action": "status"}

		addArguments(validArgs, request)
		sendRequest(request)
	case "poweroff":
		err := exec.Command("poweroff").Run()
		if err != nil {
			fmt.Printf("Error powering off: %v\n", err)
			return
		}
	case "help":
		fmt.Print(`Usage: <action> <param paramValue...>

Action: Start process
(Depends: -b or -a)
start -b /usr/bin/chromium
start -i uniqueName -b /usr/bin/firefox -arg google.com
start -a firefox

Action: Send signal
(Depends: signal and -p or -i)
signal SIGTERM -p 123150
signal SIGKILL -i uniqueName

Action: Check process status
(Depends: -p or -i)
status -p 321312
status -i uniqueName

Autostart and aliases ("-a", "--alias") are found "config.toml"
`)
		os.Exit(0)
	}
}

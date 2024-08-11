package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
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
		log.Fatal("Failed to send request:", err)
	}

	var response map[string]interface{}
	decoder := json.NewDecoder(conn)
	err = decoder.Decode(&response)
	if err != nil {
		log.Fatal("Failed to decode response:", err)
	}

	fmt.Printf("Response:%s\n", response["message"])
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

		for key, value := range validArgs {
			if key.SupportsArrays() {
				itemValues := strings.Split(value.(string), ",")

				for i, item := range itemValues {
					fmt.Printf("%s [%d] %s\n", key, i, item)
				}
				request[string(key)] = itemValues
				continue
			}
			request[string(key)] = value
		}

		sendRequest(request)

	case "stop":
		validArgs := service.CheckArguments(argumentsAfterAction)
		request := map[string]interface{}{"action": "stop", "pid": 100}

		for key, value := range validArgs {
			if key == service.PID {
				strValue, ok := value.(string)
				if !ok {
					fmt.Println("Value is not a string:", value)
				} else if intValue, err := strconv.Atoi(strValue); err == nil {
					fmt.Println("Added PID argument:", intValue)
					request[string(key)] = intValue
				} else {
					fmt.Println("Int conversion failed:", value)
				}
			} else {
				request[string(key)] = value
			}
		}
		sendRequest(request)
	case "status":
		validArgs := service.CheckArguments(argumentsAfterAction)
		request := map[string]interface{}{"action": "status"}

		for key, value := range validArgs {
			switch key {
			case service.PID:
				fmt.Println("PID argument exists for status: ", value)
				request["pid"] = value
			case service.ID:
				fmt.Println("ID argument exists for status: ", value)
				request["id"] = value
			}
		}
		sendRequest(request)
	case "help":
		fmt.Print(`Usage: <action> <param paramValue...>

Action: start
(Depends: -b or -a)
start -b /usr/bin/chromium
start -i uniqueName -b /usr/bin/firefox -arg google.com
start -a firefox

Action: stop
(Depends: -p or -i)
stop -p 123150
stop -i uniqueName

Action: status
(Depends: -p or -i)
status -p 321312
status -i uniqueName

Autostart and aliases ("-a", "--alias") are found "config.toml"
`)
		os.Exit(0)
	}
}

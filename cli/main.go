package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

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

	var response map[string]string
	decoder := json.NewDecoder(conn)
	err = decoder.Decode(&response)
	if err != nil {
		log.Fatal("Failed to decode response:", err)
	}

	fmt.Println("Response:", response["message"])
}

func main() {
	first_argument := os.Args[1]
	args := os.Args[2:]

	switch first_argument {
	case "start":
		request := map[string]interface{}{
			"action": "start",
			"id":     "",
			"env":    "",
		}

		counter := 0
		validArgs := service.CheckArguments(args)

		for key, value := range validArgs {
			fmt.Println("Argument exists: ", key, " value:", value)
			request[string(key)] = value
			counter += 2
		}

		// Start reading through arguments once "command", which is the program, is found.
		for i, arg := range args[counter:] {
			key := fmt.Sprintf("arg%d", i)
			log.Println("KEY: ", key, "VALUE: ", arg)
			request[key] = arg
		}

		sendRequest(request)

	case "stop":
		validArgs := service.CheckArguments(args)
		request := map[string]interface{}{"action": "stop"}

		for key, value := range validArgs {
			fmt.Println("Argument exists: ", key)
			request[string(key)] = value
		}
		sendRequest(request)
	case "status":
		validArgs := service.CheckArguments(args)
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
start -e DISPLAY:=0 -b /usr/bin/chromium
start -e DISPLAY:=0 -i uniqueName -b /usr/bin/firefox
start -e DISPLAY:=0 -a firefox

Action: stop
stop -p 123150
stop -i uniqueName

Action: status
status -p 321312
status -i uniqueName

Autostart and aliases "-a", "--alias"
are found "config.toml"
`)
		os.Exit(0)
	}
}

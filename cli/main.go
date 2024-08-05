package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"ops-ctrl/pkg/randomidgen"
	"ops-ctrl/pkg/service"
)

func sendRequest(request map[string]string) {
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
		// Initializing with random id, overwritten if id argument found
		randomId := randomidgen.RandomID(10)
		request := map[string]string{"action": "start", "id": randomId}

		counter := 0
		validArgs := service.CheckArguments(args)

		for key, value := range validArgs {
			switch key {
			// Only one argument allowed, rest of the arguments are
			// omitted to the binary.
			case service.Binary:
				counter += 2
				fmt.Println("Binary argument exists: ", value)
				request["binary"] = value
			case service.ID:
				counter += 2
				fmt.Println("ID argument exists: ", value)
				request["id"] = value
			case service.Alias:
				counter += 2
				fmt.Println("Alias argument exists: ", value)
				request["alias"] = value
			case service.PID:
				counter += 2
				fmt.Println("PID argument exists: ", value)
				request["pid"] = value
			case service.WorkingDir:
				counter += 2
				fmt.Print("Workingdir argument exists: ", value)
				request["working_dir"] = value
			}
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
		request := map[string]string{"action": "stop"}

		for key, value := range validArgs {
			switch key {
			case service.PID:
				fmt.Println("PID argument exists for stop: ", value)
				request["pid"] = value
			case service.ID:
				fmt.Println("ID argument exists for stop: ", value)
				request["id"] = value
			}
		}
		sendRequest(request)
	case "status":
		validArgs := service.CheckArguments(args)
		request := map[string]string{"action": "status"}

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
		fmt.Println("Usage: <action> <param paramValue...>")
		fmt.Println("Example: start -n uniqueName -b /usr/bin/firefox")
		fmt.Println("Example: start -a firefox")
		fmt.Println("Exaple: stop -p 123150")
		os.Exit(0)
	}
}

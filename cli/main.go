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

	switch first_argument {
	case "start":
		id := randomidgen.RandomID(10)
		request := map[string]string{"action": "start", "id": id}
		args := os.Args[2:]
		counter := 0

		validArgs := service.CheckArguments(args)

		if binaryValue, exists := validArgs[service.Binary]; exists {
			counter += 2
			fmt.Println("Binary exists")
			request["binary"] = binaryValue
		}

		if aliasValue, exists := validArgs[service.Alias]; exists {
			counter += 2
			fmt.Println("Alias exists")
			request["alias"] = aliasValue
		}

		if pidValue, exists := validArgs[service.PID]; exists {
			counter += 2
			fmt.Println("PID exists")
			request["pid"] = pidValue
		}

		for i, arg := range args[counter:] {
			key := fmt.Sprintf("arg%d", i)
			log.Println("KEY: ", key, "VALUE: ", arg)
			request[key] = arg
		}

		sendRequest(request)

	//case "stop":
	//case "status":
	case "help":
		fmt.Println("Usage: <action> <service_name> <param paramValue>")
		fmt.Println("Example: start -n uniqueName -b /usr/bin/firefox")
		fmt.Println("Example: start -a firefox")
		fmt.Println("Exaple: stop -p 123150")
		os.Exit(0)
	}
}

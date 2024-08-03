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
		request := map[string]string{"action": "start", "id": id, "command": os.Args[2]}
		args := os.Args[3:]

		validArgs := service.CheckArguments(args)

		// Check if there is a binary argument and add it to the request
		if binaryValue, exists := validArgs[service.BinaryArgument]; exists {
			request["binary"] = binaryValue
		}

		for i, arg := range args {
			key := fmt.Sprintf("arg%d", i)
			request[key] = arg
		}

		sendRequest(request)

	//case "stop":
	//case "status":
	case "firefox":
		id := randomidgen.RandomID(10)
		request := map[string]string{"action": "firefox", "id": id, "command": "/usr/bin/firefox"}
		args := os.Args[2:]
		for i, arg := range args {
			key := fmt.Sprintf("arg%d", i)
			request[key] = arg
		}
		sendRequest(request)
	case "help":
		fmt.Println("Usage: <action> <service_name> <param paramValue>")
		fmt.Println("Example: start -n uniqueName -b /usr/bin/firefox")
		fmt.Println("Example: start -a firefox")
		fmt.Println("Exaple: stop -p 123150")
		os.Exit(0)
	}
}

package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net"
    "os"
)

func sendRequest(action, name, command string) {
    conn, err := net.Dial("unix", "/tmp/ops-ctrl-daemon.sock")
    if err != nil {
        log.Fatal("Failed to connect to daemon:", err)
    }
    defer conn.Close()

    request := map[string]string{"action": action, "name": name, "command": command}
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
    if len(os.Args) < 3 {
        fmt.Println("Usage: <start|stop|status> <service_name> [command]")
        os.Exit(1)
    }

    action := os.Args[1]
    name := os.Args[2]
    var command string

    switch action {
		case "start":
			if len(os.Args) < 4 {
				fmt.Println("Usage: start <service_name> <command>")
				os.Exit(1)
			}
			command = os.Args[3]

		// Stop a service.
		case "stop":
		command = ""

		// Check service status.
		case "status":
		command = ""
    }

    sendRequest(action, name, command)
}

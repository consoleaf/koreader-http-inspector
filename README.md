# KOReader HTTP Inspector Client

[![Go Reference](https://pkg.go.dev/badge/github.com/Consoleaf/koreader-http-inspector.svg)](https://pkg.go.dev/github.com/Consoleaf/koreader-http-inspector)
[![Go Report Card](https://goreportcard.com/badge/github.com/Consoleaf/koreader-http-inspector)](https://goreportcard.com/report/github.com/Consoleaf/koreader-http-inspector)

A Go client library for interacting with the KOReader's HTTP Inspector. This allows you to programmatically control and query your KOReader device.

## 📦 Installation

```bash
go get github.com/Consoleaf/koreader-http-inspector
```

## 💡 Usage

```go
package main

import (
	"fmt"
	"log"
	koreaderinspector "github.com/Consoleaf/koreader-http-inspector"
)

func main() {
	koreaderURL := "http://192.168.15.244:8080" // Default IP for Usbnetlite. You can also use `http://localhost:8080` if running KOReader on PC

	client, err := koreaderinspector.New(koreaderURL)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Example: Get KOReader's Lua version
	luaVersion, err := client.GetLuaVersion()
	if err != nil {
		log.Printf("Error getting Lua version: %v", err)
	} else {
		fmt.Printf("KOReader Lua Version: %s\n", luaVersion)
	}

	// Example: Start SSH on the device
	sshPort, err := client.SSHStart()
	if err != nil {
		log.Printf("Error starting SSH: %v", err)
	} else {
		fmt.Printf("SSH started on port: %d\n", sshPort)
	}

	// Check the source code for more available methods!
}
```

---

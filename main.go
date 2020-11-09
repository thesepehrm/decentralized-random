package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
)

func main() {
	defer os.Exit(0)

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main <node_id>")
		runtime.Goexit()
	}

	if nodeID, err := strconv.Atoi(os.Args[1]); err == nil {
		Start(nodeID)
	}
}

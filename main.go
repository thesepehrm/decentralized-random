package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"gitlab.com/thesepehrm/random-miner/selector"
)

func main() {
	defer os.Exit(0)

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main <node_id>")
		runtime.Goexit()
	}

	if nodeID, err := strconv.Atoi(os.Args[1]); err == nil {
		selector.Start(nodeID)

	}

}

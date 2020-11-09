package main

import (
	"fmt"
	"log"
)

const commandLength = 12

func Handle(e error) {
	if e != nil {
		log.Panic(e)
	}
}

func BytesToCmd(data []byte) string {
	var cmd []byte

	for _, b := range data {
		if b != 0x00 {
			cmd = append(cmd, b)
		}
	}

	return fmt.Sprintf("%s", cmd)
}

func CmdToBytes(command string) []byte {
	var cmd [commandLength]byte

	for i, char := range command {
		cmd[i] = byte(char)
	}

	return cmd[:]
}

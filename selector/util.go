package selector

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/sha3"
)

const commandLength = 12

func Handle(e error) {
	if e != nil {
		log.Panic(e)
	}
}

func gobEncode(data interface{}) []byte {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	Handle(err)

	return buffer.Bytes()
}

func BuildData(command string, data interface{}) []byte {
	payload := gobEncode(data)
	req := append(CmdToBytes(command), payload...)
	return req
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

func broadcast(ownAddress string, data []byte) {
	for _, nodeAddress := range KnownNodes {
		if nodeAddress != ownAddress {
			dial(nodeAddress, data)
		}
	}
}

func dial(nodeAddress string, data []byte) {
	conn, err := net.Dial("tcp", nodeAddress)
	if err != nil {
		fmt.Printf("%s is not available\n", nodeAddress)
		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	Handle(err)
}

func HashVote(vote int) []byte {
	hash := sha3.Sum256(gobEncode(vote))
	return hash[:]
}

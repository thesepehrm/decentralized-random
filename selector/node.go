package selector

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"time"
)

const (
	VOTE_ROUNDS_INTERVAL = 10 // seconds
	VOTERS               = 5
)

var KnownNodes []string = []string{
	"localhost:3000",
	"localhost:3001",
	"localhost:3002",
	"localhost:3003",
	"localhost:3004",
}

type Node struct {
	ID int

	randomVote int

	ActiveRound         bool
	receivedReady       map[int]bool
	receivedHashedVotes map[int][]byte
	receivedVotes       map[int]int
	receivedResults     map[int]RoundResult
}

type NodeReady struct {
	From int
}

type RandomHashedVote struct {
	From        int
	HashedValue []byte
}

type RandomVote struct {
	From  int
	Value int
}

type RoundResult struct {
	From         int
	Valid        bool
	GlobalRandom int
}

func (node *Node) GetAddress() string {
	return fmt.Sprintf("localhost:%d", node.ID)
}

func (node *Node) generateRandomVote() {
	rand.Seed(10*int64(node.ID) + time.Now().Unix()) // better randomness
	node.randomVote = rand.Int() % 64
}

func (node *Node) StartPolling() {

	for {
		if time.Now().Second()%VOTE_ROUNDS_INTERVAL == 0 {

			if node.ActiveRound {
				fmt.Println("Last round is not over yet")
			} else {
				fmt.Println("Pinging network...")
				node.SendReady()
			}
		}
		time.Sleep(time.Second)
	}

}

func (node *Node) SendReady() {
	broadcast(node.GetAddress(), BuildData("sendready", NodeReady{node.ID}))
}

func (node *Node) SendHashedVote() {
	node.generateRandomVote()
	hashedVote := HashVote(node.randomVote)
	broadcast(node.GetAddress(), BuildData("sendhash", RandomHashedVote{From: node.ID, HashedValue: hashedVote}))
	fmt.Println("Broadcasted hashed vote!")
}

func (node *Node) SendVote() {
	fmt.Printf("sent %d\n", node.randomVote)
	broadcast(node.GetAddress(), BuildData("sendvote", RandomVote{node.ID, node.randomVote}))
	fmt.Println("Broadcasted vote!")

}

func (node *Node) ValidateVotes() bool {
	for nodeID, Receivedhash := range node.receivedHashedVotes {
		hashedVote := HashVote(node.receivedVotes[nodeID])
		if !bytes.Equal(hashedVote, Receivedhash) {
			fmt.Printf("vote: %d\n\thash:\t\t%x\n\treceiverhash:\t\t%x\n", node.receivedVotes[nodeID], hashedVote, Receivedhash)
			return false
		}

	}
	return true
}

func (node *Node) generateTotalRandom() int {
	total := node.randomVote
	for _, vote := range node.receivedVotes {
		total ^= vote
	}

	return total % VOTERS
}

func (node *Node) SendRoundResult() {
	valid := node.ValidateVotes()
	totalRandom := node.generateTotalRandom()

	broadcast(node.GetAddress(), BuildData("sendresult", RoundResult{From: node.ID, Valid: valid, GlobalRandom: totalRandom}))
	fmt.Println("Broadcasted results!")
}

func (node *Node) FinalizeResults() {

	valid := true
	lastRandom := -1
	for _, result := range node.receivedResults {
		if !result.Valid {
			valid = false
			break
		}
		lastRandom = result.GlobalRandom
	}
	if valid {
		fmt.Printf("Round was successful. Node #%d was selected! \n", lastRandom)
	} else {
		fmt.Printf("There were some cheaters thus the round was failed!\n")
	}
	node.ResetRound()
}

func (node *Node) ResetRound() {
	node.randomVote = 0
	node.receivedReady = make(map[int]bool)
	node.receivedHashedVotes = make(map[int][]byte)
	node.receivedVotes = make(map[int]int)
	node.receivedResults = make(map[int]RoundResult)

	node.ActiveRound = false
	fmt.Println("Ready!")
}

func (node *Node) HandleReady(req []byte) {

	var buffer bytes.Buffer
	var payload NodeReady
	buffer.Write(req[commandLength:])

	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&payload)
	Handle(err)

	if _, ok := node.receivedReady[payload.From]; !ok {

		node.receivedReady[payload.From] = true
		dial(fmt.Sprintf("localhost:%d", payload.From), BuildData("sendready", NodeReady{node.ID}))
		if len(node.receivedReady) == VOTERS-1 {
			node.ActiveRound = true
			fmt.Println("All participents are ready! Starting the selection...")
			node.SendHashedVote()
		}
	}
}

func (node *Node) HandleHashedVote(req []byte) {

	var buffer bytes.Buffer
	var payload RandomHashedVote

	buffer.Write(req[commandLength:])

	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&payload)
	Handle(err)

	if _, ok := node.receivedHashedVotes[payload.From]; !ok {

		node.receivedHashedVotes[payload.From] = payload.HashedValue

		if len(node.receivedHashedVotes) == VOTERS-1 {
			node.SendVote()
		}

	}

}

func (node *Node) HandleVote(req []byte) {
	var buffer bytes.Buffer
	var payload RandomVote

	buffer.Write(req[commandLength:])

	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&payload)
	Handle(err)

	if _, ok := node.receivedVotes[payload.From]; !ok {

		node.receivedVotes[payload.From] = payload.Value

		if len(node.receivedVotes) == VOTERS-1 {
			node.SendRoundResult()
		}

	}

}

func (node *Node) HandleRoundResult(req []byte) {
	var buffer bytes.Buffer
	var payload RoundResult

	buffer.Write(req[commandLength:])

	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&payload)
	Handle(err)

	if _, ok := node.receivedResults[payload.From]; !ok {

		node.receivedResults[payload.From] = payload

		if len(node.receivedResults) == VOTERS-1 {
			node.FinalizeResults()
		}

	}

}

func Start(id int) {
	node := Node{ID: id}
	node.ResetRound()

	listener, err := net.Listen("tcp", node.GetAddress())
	Handle(err)

	go node.StartPolling()

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		Handle(err)
		go HandleConnection(conn, &node)
	}

}

func HandleConnection(conn net.Conn, node *Node) {
	req, err := ioutil.ReadAll(conn)
	Handle(err)

	command := BytesToCmd(req[:commandLength])
	switch command {
	case "sendvote":
		node.HandleVote(req)
	case "sendhash":
		node.HandleHashedVote(req)
	case "sendresult":
		node.HandleRoundResult(req)
	case "sendready":
		node.HandleReady(req)
	default:
	}
}

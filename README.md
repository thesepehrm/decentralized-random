# Random Miner

Decentralized random selection of a group of nodes.

## How to Run

Open five shells (configurable in the code) and assign ports from 3000 to 3004 to them.

`go run main.go <port_number>`


Known Bug(s):

- Crashes after multiple rounds, there's some data racing going on.

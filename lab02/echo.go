package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

const (
	TERMINATE = "#TERMINATE#"
)

type message struct {
	NodeId  int
	Host    string
	Port    string
	Message string
}
type node struct {
	NodeId        int
	Host          string
	Port          string
	IsInitiator   bool
	HaveSent      bool
	HasReplied    bool
	ParentMessage message
}

var neighbours []node
var self node
var selfMutex sync.Mutex

func main() {
	// Setup and parse CLI flags.
	configFile := flag.String("config", "this is not a path", "Path to config file.")
	flag.Parse()

	// Check if a config file has been passed as a flag.
	if *configFile == "this is not a path" {
		panic("Invalid config file, please pass a valid config file.")
	} else {
		log.Println("Reading configuration from: " + *configFile)
	}

	// Read data from config file.
	configBytes, err := ioutil.ReadFile(*configFile)
	if err != nil {
		panic("Error reading config file.")
	}

	// Parse config file.
	configData := string(configBytes)
	scanner := bufio.NewScanner(strings.NewReader(configData))

	addresses := make([]node, 0)

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ":")
		host := line[0]
		port := line[1]

		addr := node{
			Host: host,
			Port: port,
		}

		if len(line) == 4 || len(line) == 3 {
			nodeId, _ := strconv.Atoi(line[2])
			addr.NodeId = nodeId

			if len(line) == 4 {
				// Node is an initiator.
				log.Printf("Initiator node %d (%s:%s)\n", addr.NodeId, addr.Host, addr.Port)
				addr.IsInitiator = true
				addresses = append(addresses, addr)
			} else {
				// Node is not an initiator.
				log.Printf("Non-initiator node %d (%s:%s)\n", addr.NodeId, addr.Host, addr.Port)
				addresses = append(addresses, addr)
			}

		} else if len(line) == 2 {
			addresses = append(addresses, addr)
		}
	}

	self = addresses[0]
	neighbours = addresses[1:]

	hasInitiated := false

	go listener()

	// Main event loop.
	if allNeighboursUp() {
		for {
			// Initiate communication from initiator node.
			if self.IsInitiator && !hasInitiated {
				msg := message{
					NodeId:  self.NodeId,
					Host:    self.Host,
					Port:    self.Port,
					Message: "ping",
				}

				for idx, recvAddr := range neighbours {
					sendMessage(recvAddr, msg)
					neighbours[idx].HaveSent = true
				}
				hasInitiated = true
			} else {
				time.Sleep(1 * time.Second)

				// Check if all neighbours have replied.
				if allNeighboursReplied() {
					// Send message to terminate.
					if self.IsInitiator {
						terminateNeighbours()
					} else {
						parent := node{
							Host: self.ParentMessage.Host,
							Port: self.ParentMessage.Port,
						}

						msg := message{
							NodeId:  self.NodeId,
							Host:    self.Host,
							Port:    self.Port,
							Message: "pong",
						}

						sendMessage(parent, msg)
					}
				}
			}
		}
	}
}

func listener() {
	l, _ := net.Listen("tcp", self.Host+":"+self.Port)
	defer l.Close()

	log.Printf("Listening on %s:%s\n", self.Host, self.Port)

	for {
		conn, _ := l.Accept()

		decoder := gob.NewDecoder(conn)
		var payloadData message
		if err := decoder.Decode(&payloadData); err != nil {
			// If there is an error, skip the message.
			continue
		}

		var id int
		for nid, n := range neighbours {
			if n.Port == payloadData.Port {
				id = nid
				break
			}
		}

		// Invalid message.
		if payloadData.NodeId == 0 {
			continue
		}

		log.Printf("Received %s from node %d.\n", payloadData.Message, payloadData.NodeId)

		if payloadData.Message == TERMINATE {
			terminateNeighbours()
		}

		if self.IsInitiator {
			if neighbours[id].HaveSent {
				neighbours[id].HasReplied = true
			}
		} else {
			// If node has no parent. Make node that sent this message the parent.
			selfMutex.Lock()
			if (message{} == self.ParentMessage) {
				self.ParentMessage = payloadData
				neighbours[id].HasReplied = true

				bold := color.New(color.Bold).SprintFunc()
				log.Println(bold("Parent of node"), bold(self.NodeId), bold("is node "), bold(payloadData.NodeId))

				// Send message to all other neighbours.
				for idx, receivingNode := range neighbours {
					if receivingNode.Port != self.ParentMessage.Port {
						msg := message{
							self.NodeId,
							self.Host,
							self.Port,
							"ping"}
						sendMessage(receivingNode, msg)
					}
					neighbours[idx].HaveSent = true
				}
			} else {
				if neighbours[id].HaveSent {
					neighbours[id].HasReplied = true
				}
			}
			selfMutex.Unlock()
		}
	}
}

func terminateNeighbours() {
	for _, n := range neighbours {
		selfMutex.Lock()
		if self.ParentMessage.Port != n.Port {
			msg := message{
				self.NodeId,
				self.Host,
				self.Port,
				TERMINATE}
			sendMessage(n, msg)
		}
		selfMutex.Unlock()
	}

	os.Exit(0)
}

func allNeighboursUp() bool {
	for _, addr := range neighbours {
		for {
			log.Printf("Trying to dial %s:%s\n", addr.Host, addr.Port)
			conn, err := net.Dial("tcp", addr.Host+":"+addr.Port)

			if err == nil {
				conn.Close()
				log.Printf("Successfully dialled %s:%s\n", addr.Host, addr.Port)
				break
			}

			time.Sleep(1 * time.Second)
		}
	}

	return true
}

func allNeighboursReplied() bool {
	allReplied := true
	for _, addr := range neighbours {
		if self.ParentMessage.Port != addr.Port && !addr.HasReplied {
			allReplied = false
		}
	}

	return allReplied
}

func sendMessage(recvAddr node, msg message) {
	log.Printf("Sending %s to %s:%s.\n", msg.Message, recvAddr.Host, recvAddr.Port)
	for {
		conn, err := net.Dial("tcp", recvAddr.Host+":"+recvAddr.Port)

		if err == nil && msg.NodeId != 0 {
			encoder := gob.NewEncoder(conn)
			if err := encoder.Encode(msg); err != nil {
				log.Fatal(err)
			}

			conn.Close()
			break
		}
	}
	log.Println("Sent " + msg.Message + " to " + recvAddr.Host + ":" + recvAddr.Port + ".")
}

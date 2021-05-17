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
)

// TERMINATE is a special message that is sent to terminate a child node.
const (
	TERMINATE = "#TERMINATE#"
)

// message is used to send payloads from one node to another.
// messages carry ID of the sender node, Host (IP) of the sender, Port of the
// sender and a string Message.
type message struct {
	NodeId  int
	Host    string
	Port    string
	Message string
	Leader  int
}

// node represents details pertaining to different nodes in the network graph.
// NodeId is the ID of a node, Host is the IP address, Port is the port used,
// IsInitiator indicates if a node is an initiator, HaveSent and HasReplied are
// used to track whether nodes have sent and replied to messages, ParentMessage
// is used to keep track of the parent of a node.
type node struct {
	NodeId        int
	Host          string
	Port          string
	IsInitiator   bool
	HaveSent      bool
	HasReplied    bool
	Leader        int
	ParentMessage message
}

var neighbours []node    // Neighbours of the current node.
var self node            // Current node (self)
var selfMutex sync.Mutex // Mutex to manage access to member of self.
var hasInitiated bool    // Track if initiation message has been sent.

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

	self = addresses[0]        // Populate current node.
	self.Leader = self.NodeId  // Initialize current node to leader.
	neighbours = addresses[1:] // Populate neighbours.
	hasInitiated := false

	// Current leader for each node is set to the ID of self.
	log.Println("Current leader is", self.Leader)

	go listener() // Run goroutine to listen for messages.

	// Main event loop.
	if allNeighboursUp() {
		// If all neighbours are up then proceed with main event loop.
		for {
			// Initiate communication from initiator node.
			if self.IsInitiator && !hasInitiated {
				// Send ping message to neighbours.
				msg := message{
					NodeId:  self.NodeId,
					Host:    self.Host,
					Port:    self.Port,
					Message: "ping",
					Leader:  self.NodeId,
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
					if self.IsInitiator {
						// Send message to terminate.
						terminateNeighbours()
					} else {
						// Send pong message to parent.
						parent := node{
							Host: self.ParentMessage.Host,
							Port: self.ParentMessage.Port,
						}

						msg := message{
							NodeId:  self.NodeId,
							Host:    self.Host,
							Port:    self.Port,
							Message: "pong",
							Leader:  self.Leader,
						}

						sendMessage(parent, msg)
					}
				}
			}
		}
	}
}

// listener listens for incoming connections.
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

		// Message to terminate received from parent.
		if payloadData.Message == TERMINATE && payloadData.NodeId == self.ParentMessage.NodeId {
			terminateNeighbours()
		}

		leaderChanged := false

		// Change leader if a node with higher ID is received.
		selfMutex.Lock()
		if payloadData.Leader > self.Leader || self.Leader == 0 {
			log.Printf("Changing leader to node %d.\n", payloadData.Leader)
			self.Leader = payloadData.Leader
			leaderChanged = true
		}
		selfMutex.Unlock()

		// If leader has changed for current node then broadcast to all nodes
		// about it.
		if leaderChanged {
			for nid, n := range neighbours {
				neighbours[nid].HaveSent = true    // Mark message as sent.
				neighbours[nid].HasReplied = false // Mark reply as false.
				hasInitiated = false               // Initiator needs to reinitiate.

				msg := message{
					NodeId:  self.NodeId,
					Host:    self.Host,
					Port:    self.Port,
					Message: "leader changed",
					Leader:  self.Leader,
				}

				sendMessage(n, msg)
			}
		}

		if self.IsInitiator {
			if neighbours[id].HaveSent {
				// Reply received from a node that was previously contacted.
				neighbours[id].HasReplied = true
			}
		} else {
			// If node has no parent or wave tagged with higher id hits node.
			// Make node that sent this message the parent.
			// Mutex used to restrict access to members of self struct.
			selfMutex.Lock()

			if (message{} == self.ParentMessage) {
				self.ParentMessage = payloadData
				neighbours[id].HasReplied = true

				log.Printf("Parent of node %d is node %d.\n", self.NodeId, payloadData.NodeId)

				// Send ping message to neighbours.
				for idx, receivingNode := range neighbours {
					if receivingNode.Port != self.ParentMessage.Port {
						msg := message{
							NodeId:  self.NodeId,
							Host:    self.Host,
							Port:    self.Port,
							Message: "ping",
							Leader:  self.Leader,
						}
						sendMessage(receivingNode, msg)
					}
					// Mark message as being sent to node.
					neighbours[idx].HaveSent = true
				}
			} else {
				if neighbours[id].HaveSent {
					// Reply received from a node that was previously contacted.
					neighbours[id].HasReplied = true
				}
			}
			selfMutex.Unlock()
		}
	}
}

// terminateNeighbours send messages to non parent neighbours of node to
// terminate.
func terminateNeighbours() {
	for _, n := range neighbours {
		selfMutex.Lock()
		if self.ParentMessage.Port != n.Port {
			msg := message{
				NodeId:  self.NodeId,
				Host:    self.Host,
				Port:    self.Port,
				Message: TERMINATE,
				Leader:  self.Leader,
			}
			sendMessage(n, msg)
		}
		selfMutex.Unlock()
	}

	log.Println("Leader is:", self.Leader)
	log.Println("Sleeping for 3 seconds before terminating.")
	time.Sleep(3 * time.Second)

	os.Exit(0)
}

// allNeighboursUp checks if all neighbours can be reached if not, it will
// block till all neighbours are up.
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

// allNeighboursReplied checks if all neighbours have replied to a node.
func allNeighboursReplied() bool {
	allReplied := true
	for _, addr := range neighbours {
		if self.ParentMessage.Port != addr.Port && !addr.HasReplied {
			allReplied = false
		}
	}

	return allReplied
}

// sendMessage sends msg of type message to node recvAddr. Retries sending
// indefinitely.
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

		time.Sleep(1 * time.Second)
	}
	log.Printf("Sent %s to %s:%s.\n", msg.Message, recvAddr.Host, recvAddr.Port)
}

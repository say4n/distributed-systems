package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
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
	Host    string
	Port    string
	Message string
	Leader  int
	Round   int
	Size    int
}

// node represents details pertaining to different nodes in the network graph.
// NodeId is the ID of a node, Host is the IP address, Port is the port used,
// IsInitiator indicates if a node is an initiator, HaveSent and HasReplied are
// used to track whether nodes have sent and replied to messages, ParentMessage
// is used to keep track of the parent of a node.
type node struct {
	Size          int
	Host          string
	Port          string
	IsInitiator   bool
	HaveSent      bool
	HasReplied    bool
	ParentMessage message
}

var neighbours []node    // Neighbours of the current node.
var self node            // Current node (self)
var selfMutex sync.Mutex // Mutex to manage access to member of self.
var hasInitiated bool    // Track if initiation message has been sent.
var roundNumber int      // Track round numbers.
var leader int           // Leader.
var randomId int         // Random ID.
var status bool          // Current status of node.
var numNodes int         // Size of the network.

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

	addresses := parseConfig(*configFile)

	self = addresses[0]        // Populate current node.
	neighbours = addresses[1:] // Populate neighbours.
	hasInitiated := false

	go listener() // Run goroutine to listen for messages.

	// Main event loop.
	if allNeighboursUp() {
		// If all neighbours are up then proceed with main event loop.
		shouldExit := false

		for {
			// Initiate communication from initiator node.
			if self.IsInitiator && !hasInitiated {
				// Send ping message to neighbours.
				roundNumber = 0

				msg := message{
					Host:    self.Host,
					Port:    self.Port,
					Message: "ping",
					Leader:  leader,
					Round:   roundNumber,
					Size:    0,
				}

				sendMessageToAllNeighbours(msg)
				hasInitiated = true
			} else {
				time.Sleep(3 * time.Second)

				size := computeSize()

				if status {
					if size == numNodes {
						// Done!
						log.Println("Done.")

						shouldExit = true
					} else {
						if allNeighboursReplied() {
							roundNumber = roundNumber + 1
							randomId = getRandomId()
							leader = randomId

							log.Println("New ID is: ", leader)

							msg := message{
								Host:    self.Host,
								Port:    self.Port,
								Message: "ping",
								Leader:  leader,
								Round:   roundNumber,
								Size:    0,
							}

							resetActivities()
							sendMessageToAllNeighbours(msg)
						}
					}
				} else {
					if allNeighboursReplied() {
						msg := message{
							Host:    self.Host,
							Port:    self.Port,
							Message: "pong",
							Leader:  leader,
							Round:   roundNumber,
							Size:    size,
						}

						parent := node{
							Host: self.ParentMessage.Host,
							Port: self.ParentMessage.Port,
						}

						log.Printf("Network size: %d, Detected size: %d.\n", numNodes, size)
						sendMessage(parent, msg)
						resetActivities()
					}
				}
			}

			if shouldExit {
				log.Println("I was elected leader.")
				log.Printf("Detected network size is: %d, should be: %d.\n", computeSize(), numNodes)

				terminateNeighbours()
			}
		}
	}
}

func resetActivities() {
	for nid := 0; nid < len(neighbours); nid++ {
		neighbours[nid].HaveSent = false
		neighbours[nid].HasReplied = false
		neighbours[nid].Size = 0
	}

	hasInitiated = false
}

func getRandomId() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(numNodes) + 1
}

func computeSize() int {
	size := 0

	for _, node := range neighbours {
		if self.ParentMessage.Port != node.Port {
			size += node.Size
		}
	}

	return size + 1
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
		err := decoder.Decode(&payloadData)
		if err != nil {
			// If there is an error, skip the message.
			continue
		}

		log.Printf("Received message from %s:%s.\n", payloadData.Host, payloadData.Port)
		log.Printf("Round is %d, payload round is %d.", roundNumber, payloadData.Round)
		log.Printf("Leader is %d, payload leader is %d.", leader, payloadData.Leader)

		// Message to terminate received from parent.
		if payloadData.Message == TERMINATE {
			terminateNeighbours()
		}

		// Invalid message.
		if payloadData.Host == "" || payloadData.Port == "" || payloadData.Leader == 0 {
			log.Printf("Message from %s:%s was INVALID.\n", payloadData.Host, payloadData.Port)
			continue
		}

		var id int
		for nid, n := range neighbours {
			if n.Port == payloadData.Port {
				id = nid
				break
			}
		}

		if payloadData.Message == "pong" && neighbours[id].HaveSent {
			log.Printf("Received reply from %s:%s.\n", payloadData.Host, payloadData.Port)
			neighbours[id].HasReplied = true
			neighbours[id].Size = payloadData.Size
		} else {
			if payloadData.Round > roundNumber || (payloadData.Round == roundNumber && payloadData.Leader > leader) {
				log.Printf("Selecting %d as leader.", payloadData.Leader)
				status = false
				resetActivities()

				neighbours[id].HasReplied = true
				leader = payloadData.Leader
				roundNumber = payloadData.Round

				self.ParentMessage = payloadData

				msg := message{
					Host:    self.Host,
					Port:    self.Port,
					Message: "ping",
					Leader:  leader,
					Round:   roundNumber,
					Size:    0,
				}

				sendMessageToAllNeighbours(msg)
			}

			if payloadData.Round < roundNumber || (payloadData.Round == roundNumber && payloadData.Leader < leader) {
				log.Printf("Ignoring message from %s:%s. Current leader is: %d.\n", payloadData.Host, payloadData.Port, leader)
			}

			if payloadData.Round == roundNumber && payloadData.Leader == leader {
				log.Printf("Leader %d remains unchanged.\n", leader)
				neighbours[id].HasReplied = true
				neighbours[id].Size = payloadData.Size
			}
		}
	}
}

// terminateNeighbours send messages to non parent neighbours of node to
// terminate.
func terminateNeighbours() {
	for _, n := range neighbours {
		msg := message{
			Host:    self.Host,
			Port:    self.Port,
			Message: TERMINATE,
			Leader:  leader,
		}
		sendMessage(n, msg)
	}

	log.Printf("I am : %d, leader is %d.\n", randomId, leader)
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
		selfMutex.Lock()
		if self.ParentMessage.Port != addr.Port && !addr.HasReplied {
			allReplied = false
		}
		selfMutex.Unlock()
	}

	return allReplied
}

func sendMessageToAllNeighbours(msg message) {
	for idx, recvAddr := range neighbours {
		sendMessage(recvAddr, msg)
		neighbours[idx].HaveSent = true
	}
}

// sendMessage sends msg of type message to node recvAddr. Retries sending
// indefinitely.
func sendMessage(recvAddr node, msg message) {
	log.Printf("Sending %s to %s:%s.\n", msg.Message, recvAddr.Host, recvAddr.Port)
	for {
		conn, err := net.Dial("tcp", recvAddr.Host+":"+recvAddr.Port)

		if err == nil {
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

// parseConfig parses config file.
func parseConfig(configFile string) []node {
	// Read data from config file.
	configBytes, err := ioutil.ReadFile(configFile)
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
			numNodes, _ = strconv.Atoi(line[2])

			if len(line) == 4 {
				// Node is an initiator.
				log.Printf("Initiator node (%s:%s)\n", addr.Host, addr.Port)
				addr.IsInitiator = true
				status = true
				addresses = append(addresses, addr)
			} else {
				// Node is not an initiator.
				log.Printf("Non-initiator node (%s:%s)\n", addr.Host, addr.Port)
				addresses = append(addresses, addr)
				status = false
			}

		} else if len(line) == 2 {
			addresses = append(addresses, addr)
		}
	}

	randomId = getRandomId()
	leader = randomId
	log.Println("Random ID is: ", randomId)

	return addresses
}

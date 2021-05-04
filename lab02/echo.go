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
)

type address struct {
	NodeId      int
	Host        string
	Port        string
	IsInitiator bool
}

type message struct {
	NodeId  int
	Host    string
	Port    string
	Message string
	IsReply bool
}

type graph struct {
	Parent     address
	Neighbours []address
}

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

	addresses := make([]address, 0)

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ":")
		addr := line[0]
		port := line[1]

		if len(line) == 4 || len(line) == 3 {
			nodeId, _ := strconv.Atoi(line[2])

			if len(line) == 4 {
				// Node is an initiator.
				log.Println("Initiator node: " + addr + ":" + port)
				addresses = append(addresses, address{nodeId, addr, port, true})
			} else {
				// Node is not an initiator.
				log.Println("Non-initiator node: " + addr + ":" + port)
				addresses = append(addresses, address{nodeId, addr, port, false})
			}

		} else if len(line) == 2 {
			addresses = append(addresses, address{-1, addr, port, false})
		} else {

		}
	}

	selfAddr := addresses[0]
	addresses = addresses[1:]
	networkGraph := graph{}
	var ngMutex sync.Mutex

	l, err := net.Listen("tcp", selfAddr.Host+":"+selfAddr.Port)
	if err != nil && err.Error() != "EOF" {
		log.Fatal(err)
	}
	defer l.Close()

	// Send message to all children if node is an initiator.
	if selfAddr.IsInitiator {
		for _, node := range addresses {
			sendMessage(selfAddr, node, "ping", false)
		}
	}

	// Check if message received from all neighbours (except its parent).
	go func() {
		goroutineShouldTerminate := false
		for {
			ngMutex.Lock()

			if len(addresses) == len(networkGraph.Neighbours) {
				parentAddr := address{
					Host: networkGraph.Parent.Host,
					Port: networkGraph.Parent.Port,
				}

				log.Printf("Node %d received message from all neighbours.\n", selfAddr.NodeId)
				if selfAddr.IsInitiator {
					for _, addr := range addresses {
						sendMessage(selfAddr, addr, "TERMINATE", false)
					}

					log.Println("Terminating")
					os.Exit(0)
				} else {
					// Send message to parent.
					sendMessage(selfAddr, parentAddr, "pong", true)
					goroutineShouldTerminate = true
				}
			}
			ngMutex.Unlock()

			if goroutineShouldTerminate {
				break
			}
		}
	}()

	terminate := false
	var tMutex sync.Mutex

	for {
		tMutex.Lock()
		if terminate {
			log.Println("Terminating.")
			tMutex.Unlock()

			break
		}
		tMutex.Unlock()

		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// This goroutine enables handling a new connection in a concurrent
		// way.
		go func(c net.Conn) {
			decoder := gob.NewDecoder(c)
			var payloadData message
			if err := decoder.Decode(&payloadData); err != nil {
				log.Println(err.Error())
			}

			defer c.Close()

			log.Printf("Node %d received message %s from node %d.\n", selfAddr.NodeId, payloadData.Message, payloadData.NodeId)

			addr := address{
				Host:   payloadData.Host,
				Port:   payloadData.Port,
				NodeId: payloadData.NodeId,
			}

			ngMutex.Lock()
			if networkGraph.Parent == (address{}) {
				// First communication, make sender parent of current node.
				if !selfAddr.IsInitiator {
					log.Println("Parent has ID :", payloadData.NodeId)

					networkGraph.Parent = addr

					// Send messages to all other neighbouring nodes.
					for _, node := range addresses {
						if node.Port != networkGraph.Parent.Port {
							sendMessage(selfAddr, node, payloadData.Message, false)
						}
					}
				}
			} else {
				// Subsequent communications.
				found := false
				for _, tn := range networkGraph.Neighbours {
					if tn.Port == payloadData.Port && tn.Host == payloadData.Host {
						found = true
						break
					}
				}
				// Only add the node to the network graph if it is not a reply to
				// a previously sent message.
				if !found && payloadData.IsReply {
					log.Printf("Received reply from node %d at %d\n", payloadData.NodeId, selfAddr.NodeId)
					networkGraph.Neighbours = append(networkGraph.Neighbours, addr)
				}
			}
			ngMutex.Unlock()

			if payloadData.Message == "TERMINATE" {
				tMutex.Lock()
				terminate = true
				tMutex.Unlock()
			}
		}(conn)
	}
}

func sendMessage(selfAddr, recvAddr address, msg string, isReply bool) {
	log.Println("Sending " + msg + " to " + recvAddr.Host + ":" + recvAddr.Port + ".")
	for {
		conn, err := net.Dial("tcp", recvAddr.Host+":"+recvAddr.Port)

		if err == nil {
			encoder := gob.NewEncoder(conn)
			if err := encoder.Encode(message{selfAddr.NodeId, selfAddr.Host, selfAddr.Port, msg, isReply}); err != nil {
				log.Fatal(err)
			}

			conn.Close()

			break
		}

	}
	log.Println("Sent " + msg + " to " + recvAddr.Host + ":" + recvAddr.Port + ".")
}

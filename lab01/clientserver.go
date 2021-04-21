package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

// address stores the IP address as well as nature of a node.
// In addition to this, it also store the port, the ID of the node if it is
// the listening address for a node and a flag that tells if a node if the
// address is to be used for listening or sending messages to.
type address struct {
	id         int
	host       string
	port       string
	willListen bool
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

		// The first line of the config file has 3 ':' separated values, the IP,
		// the port and the ID of the node. This address is the one that the
		// node is configured to listen on for incoming messages.
		// All other lines only contain the IP and the port that this node will
		// send messages to.
		if len(line) == 3 {
			log.Println("Will listen for messages on: " + addr + ":" + port)
			id, _ := strconv.Atoi(line[2])
			addresses = append(addresses, address{id, addr, port, true})

		} else if len(line) == 2 {
			log.Println("Will broadcast messages to: " + addr + ":" + port)
			addresses = append(addresses, address{-1, addr, port, false})
		}
	}

	// Channel to receive input from the stdin.
	ch := make(chan string)

	// Non-blocking read from standard input. This goroutine checks for text
	// text input from the stdin and passes it to the previously defined channel
	// in case any input is provided by the user and then the return key is
	// pressed.
	go func(ch chan string) {
		fmt.Println("Type messages in the console and hit return to send.")
		reader := bufio.NewReader(os.Stdin)

		for {
			// Read stdin till a new line is encountered.
			s, err := reader.ReadString('\n')
			if err != nil {
				close(ch)
				log.Fatal(err.Error())
				return
			}

			// The text input is passed to the channel after trimming spaces.
			ch <- strings.TrimSpace(s)
		}
	}(ch)

	// Instantiate TCP listener. Since the first line of the file is the address
	// at which the node will listen for incoming messages, it is indexed
	// directly from the array here.
	l, err := net.Listen("tcp", addresses[0].host+":"+addresses[0].port)
	if err != nil && err.Error() != "EOF" {
		log.Fatal(err)
	}
	defer l.Close()

	// This goroutine is checks for incoming connections to the previously
	// defined listener. If a message is received, it logs it to the stdout
	// alongwith information about the remote address it received the
	// message from.
	go func() {
		for {
			// Wait for a connection.
			conn, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}

			// This goroutine enables handling a new connection in a concurrent
			// way.
			go func(c net.Conn) {
				// Read and parse data from the connection to a string till a
				// newline is encountered.
				netData, err := bufio.NewReader(c).ReadString('\n')
				if err != nil {
					log.Println(err.Error())
				}

				defer c.Close()

				// Log received message to stdin.
				log.Print("Message from " + c.RemoteAddr().String() + " > " + string(netData))
			}(conn)
		}
	}()

	// eventloop label is used to check for messages in the previously defined
	// channel for stdin. If a message is received, all the addresses that were
	// initially registered as peers of the node are dialled to with a TCP
	// connection and the corresponding message from the channel is forwarded to
	// them.
	// If there is any issue with the channel, then the control exits from the
	// loop and the program exits.
eventloop:
	for {
		select {
		case stdin, ok := <-ch:
			if !ok {
				break eventloop
			} else {
				data := stdin

				for _, addrItem := range addresses {
					// If this address is a peer node, then send the message.
					if !addrItem.willListen {

						// This goroutine dials to a given address (addr) of a
						// peer node and send it the message from stdin (data).
						go func(addr address, data string) {
							conn, err := net.Dial("tcp", addr.host+":"+addr.port)
							if err != nil {
								log.Fatalf("Failed to dial: %v", err)
							}

							defer conn.Close()

							log.Println("Sending `" + data + "` to " + addr.host + ":" + addr.port + ".")

							// Data is sent to the peer node here as a byte stream.
							if _, err := conn.Write([]byte(data + "\n")); err != nil {
								log.Fatal(err)
							}
						}(addrItem, data)
					}
				}
			}
		}
	}
}

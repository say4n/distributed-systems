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

		if len(line) == 3 {
			log.Println("Will listen for messages on: " + addr + ":" + port)
			id, _ := strconv.Atoi(line[2])
			addresses = append(addresses, address{id, addr, port, true})

		} else if len(line) == 2 {
			log.Println("Will broadcast messages to: " + addr + ":" + port)
			addresses = append(addresses, address{-1, addr, port, false})
		}
	}

	ch := make(chan string)

	// Non-blocking read from standard input.
	go func(ch chan string) {
		fmt.Println("Type messages in the console to send to other nodes.")
		reader := bufio.NewReader(os.Stdin)
		for {
			s, err := reader.ReadString('\n')
			if err != nil {
				close(ch)
				log.Fatal(err.Error())
				return
			}
			ch <- strings.TrimSpace(s)
		}
	}(ch)

	// Instantiate listener.
	l, err := net.Listen("tcp", addresses[0].host+":"+addresses[0].port)
	if err != nil && err.Error() != "EOF" {
		log.Fatal(err)
	}
	defer l.Close()

	go func() {
		for {
			// Wait for a connection.
			conn, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}

			go func(c net.Conn) {
				netData, err := bufio.NewReader(c).ReadString('\n')
				if err != nil {
					log.Println(err.Error())
				}

				log.Println("Message from " + c.RemoteAddr().String() + " > " + string(netData))

				c.Close()
			}(conn)
		}
	}()

eventloop:
	for {
		select {
		case stdin, ok := <-ch:
			if !ok {
				break eventloop
			} else {
				data := stdin

				for _, addrItem := range addresses {
					if !addrItem.willListen {

						go func(addr address, data string) {
							conn, err := net.Dial("tcp", addr.host+":"+addr.port)
							if err != nil {
								log.Fatalf("Failed to dial: %v", err)
							}

							fmt.Println("Sending `" + data + "` to " + addr.host + ":" + addr.port + ".")

							if _, err := conn.Write([]byte(data + "\n")); err != nil {
								log.Fatal(err)
							}

							conn.Close()
						}(addrItem, data)
					}
				}
			}
		}
	}
}

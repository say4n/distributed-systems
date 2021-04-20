package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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

	fmt.Println(addresses)
}

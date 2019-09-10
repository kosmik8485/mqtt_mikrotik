package main

import (
	"flag"
	"log"
	"strings"

	"gopkg.in/routeros.v2"
)

var (
	command  = flag.String("command", "/interface/wireless/registration-table/print", "ROS command")
	address  = flag.String("address", "192.168.88.1:8728", "ROS address")
	username = flag.String("username", "admin", "ROS Username")
	password = flag.String("password", "JashOtEag6", "ROS password")
	async	 = flag.Bool("async", false, "Use async code")
	useTLS   = flag.Bool("tls", false, "Use TLS")
)

func dial() (*routeros.Client, error) {
	if *useTLS {
		return routeros.DialTLS(*address, *username, *password, nil)
	}
	return routeros.Dial(*address, *username, *password)
}

func main() {
	flag.Parse()

	c, err := dial()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	if *async {
		c.Async()
	}

	r, err := c.RunArgs(strings.Split(*command, " "))
	if err != nil {
		log.Fatal(err)
	}

	log.Println(r)
}

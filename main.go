package main

import (
	"flag"
	"log"
	"strings"
	"time"
	"os"
	"fmt"
	"os/signal"
	"syscall"
	"encoding/json"
	
	"gopkg.in/routeros.v2"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	command  = flag.String("command", "/interface/wireless/registration-table/print", "ROS command")
	address  = flag.String("address", "192.168.88.1:8728", "ROS address")
	username = flag.String("username", os.Getenv("ros_username"), "ROS Username")
	password = flag.String("password", os.Getenv("ros_password"), "ROS password")
	async	 = flag.Bool("async", false, "Use async code")
	useTLS   = flag.Bool("tls", false, "Use TLS")
	logLevel = flag.Int("loglevel",1,"log level (0 - disable, 1 - info, 2 - debug, 3 - error only)")
	
	mqtt_user  = flag.String("mqtt_user", os.Getenv("mqtt_user"), "MQTT username")
	mqtt_pass  = flag.String("mqtt_pass", os.Getenv("mqtt_pass"), "MQTT password")
	mqtt_topic = flag.String("mqtt_topic", "router/home", "MQTT topic")
	mqtt_addr  = flag.String("mqtt_addr", "srv.rpi:1883", "MQTT address")
	mqtt_upd   = flag.Int("mqtt_upd", 10, "MQTT update timeout in secs")
)

func main() {
	flag.Parse()

	lg("Init and dial","info")
	
	interrupt := make(chan os.Signal, 1)
	signal.Notify( interrupt, os.Interrupt, syscall.SIGTERM)
	
	c, err := dial()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	if *async {
		lg("Start in Async mode","info")
		c.Async()
	}

	client := connect("pub")
	
	go func() {
		for {
			time.Sleep(*mqtt_upd * time.Second)

			r, err := c.RunArgs(strings.Split(*command, " "))
			if err != nil {
				log.Fatal(err)
			}				
			data, _ := json.Marshal(r)			
			client.Publish(*mqtt_topic, 0, false, data)
		}
	}()
	
	killSig := <-interrupt
	switch killSig {
		case os.Interrupt:
			lg("Got SIGINT...", "error")
		case syscall.SIGTERM:
			lg("Got SIGTERM...", "error")
	}
	
	lg("Service is shutdown...", "info")
}

func connect(clientId string) mqtt.Client {
	opts := createClientOptions(clientId)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func createClientOptions(clientId string) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", *mqtt_addr))
	opts.SetUsername(*mqtt_user)	
	opts.SetPassword(*mqtt_pass)
	opts.SetClientID(clientId)
	return opts
}

func lg(msg, level string) {
	msgLevel := 0
	switch level {
		case "info":
			msgLevel = 1
		case "debug":
			msgLevel = 2
		case "error":
			msgLevel = 3
		default:
			msgLevel = 0			
	}
	if msgLevel >= *logLevel {
		log.Printf("[%s] %s\n", uc(level), msg)		
	}
}

func dial() (*routeros.Client, error) {
	if *useTLS {
		lg("Use TLS","info")
		return routeros.DialTLS(*address, *username, *password, nil)
	}
	return routeros.Dial(*address, *username, *password)
}

package main

import (
	"flag"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"os"
	cfg "solis-go/config"
	"solis-go/inverter/solis"
)

var config *cfg.Config

func init() {
	configfile := flag.String("config", "config.development.json", "config file to use")
	flag.Parse()
	config = cfg.LoadConfiguration(configfile)
}

func main() {
	fmt.Println("Starting Solis Inverter Connector")

	mqo := mqtt.NewClientOptions()
	mqo.AddBroker(config.MQTTBroker)

	mq := mqtt.NewClient(mqo)

	if !mq.IsConnected() {
		fmt.Println("[New]: Connecting to mqtt broker")
		t := mq.Connect()
		if t.Wait() && t.Error() != nil {
			fmt.Printf("[New]: Unable to connect to mqtt broker: %s", t.Error().Error())
			os.Exit(1)
		}
	}

	// TODO: configurable inverter types
	inv := solis.New(mq, config)
	inv.Run()

}


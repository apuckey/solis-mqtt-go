package cfg

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Config struct {
	Hostname string `json:"hostname"`
	Inverter struct {
		Serial         uint64 `json:"serial"`
		Host           string `json:"host"`
		ConnectTimeout int    `json:"connecttimeout"`
		ReadTimeout    int    `json:"readtimeout"`
		ReadInterval   int    `json:"readinterval"`
	} `json:"inverter"`
	MQTTBroker string `json:"mqttbroker"`
}

func LoadConfiguration(filename *string) *Config {
	config := &Config{}

	configFile, err := os.Open(*filename)
	if err != nil {
		log.Fatal(fmt.Sprintf("Configuration file not found: %s, %s", *filename, err))
	}

	defer func() {
		_ = configFile.Close()
	}()

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		fmt.Printf("Error decoding json config file: %s", err.Error())
		os.Exit(1)
	}

	hostname, err := os.Hostname()
	if err != nil {
		config.Hostname = "unknown"
	} else {
		config.Hostname = hostname
	}

	return config
}

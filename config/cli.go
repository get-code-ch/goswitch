package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func NewCliConfig(configFile string) *ConfCli {

	// New config creation
	c := new(ConfCli)

	// If no config file is provided we use "hardcoded" default filepath
	if configFile == "" {
		configFile = defaultCliConfigFile
	}

	// Testing if config file exist if not, return a fatal error
	_, err := os.Stat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Panic(fmt.Sprintf("Config file %s not exist\n", configFile))
		} else {
			log.Panic(fmt.Sprintf("Something wrong with config file %s -> %v\n", configFile, err))
		}
	}
	buffer, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf(fmt.Sprintf("Error reading config file --> %v", err))
		return nil
	}

	// Parsing config file
	json.Unmarshal(buffer, c)

	return c
}

// TODO Create Stringer interface to return human readable config content
func (c *ConfCli) String() string {
	return fmt.Sprintf("Contoller %v", c.Controller)
}

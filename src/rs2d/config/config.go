package config

import (
	"os"
	"fmt"
	"encoding/json"
	"crypto/rsa"
)

var loaded = false

// Set default values here
var Config = configuration {
	Port: 43594,
	JaggrabPort: 43595,
	ReadTimeout: 30,
	UseRSAEncryption: false,
	CachePath: "cache",
}

type configuration struct {
	Port             int
	JaggrabPort      int
	ReadTimeout      int
	PrivateKey       rsa.PrivateKey
	UseRSAEncryption bool
	CachePath 		 string
}

func Save() {
	file, err := os.Create("config.json")
	defer file.Close()

	if err != nil {
		fmt.Printf("Failed to open config.json for saving, err: %v\n", err)
		return
	}

	data, err := json.MarshalIndent(Config, "", "  ")

	if err != nil {
		fmt.Printf("Error saving config.json: %v\n", err)
		return
	}
	for len(data) > 0 {
		n, err := file.Write(data)
		if err != nil {
			fmt.Printf("Error saving config.json: %v\n", err)
			return
		}
		data = data[n:]
	}
}

func Load() {
	loaded = true

	file, err := os.Open("config.json")
	defer file.Close()
	if err == nil {
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&Config)

		if err != nil {
			fmt.Printf("Error loading config.json: %v\n", err)
			os.Exit(-4)
		}
	}
	Save()
}
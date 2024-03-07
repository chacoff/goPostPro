package main

import (
	"encoding/xml"
	"fmt"
	"os"
)

type Config struct {
	XMLName       xml.Name `xml:"config"`
	Cage          string   `xml:"cage"`
	NetType       string   `xml:"netType"`
	Address       string   `xml:"address"`
	MaxBufferSize int      `xml:"maxBufferSize"`
	HeaderSize    int      `xml:"headerSize"`
	Verbose       bool     `xml:"verbose"`
}

func loadConfig() Config {
	file, err := os.Open("./config.xml")
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	var _config Config
	if err := xml.NewDecoder(file).Decode(&_config); err != nil {
		fmt.Println("Error decoding XML:", err)
		os.Exit(1)
	}

	fmt.Printf("Parameters Ok.\n")

	return _config
}

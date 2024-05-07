package config

import (
	"encoding/xml"
	"log"
	"os"
)

type Config struct {
	XMLName       xml.Name `xml:"config"`
	Cage          string   `xml:"cage"`
	NetType       string   `xml:"netType"`
	Address       string   `xml:"address"`
	AddressDias   string   `xml:"addressDias"`
	MaxBufferSize int      `xml:"maxBufferSize"`
	HeaderSize    int      `xml:"headerSize"`
	Verbose       bool     `xml:"verbose"`
	DataFolders   string   `xml:"jsonFolders"`
	TickWatcher   int      `xml:"tickWatcher"`
}

func LoadConfig() Config {
	file, err := os.Open("./config.xml")
	if err != nil {
		log.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	var _config Config
	if err := xml.NewDecoder(file).Decode(&_config); err != nil {
		log.Println("Error decoding XML:", err)
		os.Exit(1)
	}

	log.Println("[Parameters] OK")

	return _config
}

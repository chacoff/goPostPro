package config

import (
	"encoding/xml"
	"log"
	"os"
)

type Parameters struct {
	XMLName  xml.Name  `xml:"parameters"`
	Config   Config    `xml:"config"`
	PostPro  PostPro   `xml:"postpro"`
	DataBase DataBase  `xml:"database"`
	Logs     LogParams `xml:"logger"`
}

type Config struct {
	Cage          string `xml:"cage"`
	NetType       string `xml:"netType"`
	Address       string `xml:"address"`
	AddressDias   string `xml:"addressDias"`
	MaxBufferSize int    `xml:"maxBufferSize"`
	HeaderSize    int    `xml:"headerSize"`
	Verbose       bool   `xml:"verbose"`
}

type LogParams struct {
	FileName   string `xml:"fileName"`
	MaxSize    int    `xml:"maxSize"`
	MaxBackups int    `xml:"maxBackups"`
	MaxAge     int    `xml:"maxAge"`
	Compress   bool   `xml:"compress"`
}

type PostPro struct {
	TimeFormat              string  `xml:"timeFormat"`
	FirstMeasuresRemoved    int     `xml:"firstRemoved"`
	AdaptativeFactor        float64 `xml:"adaptativeFactor"`
	MinTemperatureThreshold float64 `xml:"minTemperatureThreshold"`
	GradientFactor          float64 `xml:"gradientFactor"`
	MinWidth                int64   `xml:"minWidth"`
}

type DataBase struct {
	Path              string `xml:"path"`
	TimeFormatRequest string `xml:"timeFormatRequest"`
}

func LoadConfig() Parameters {
	file, err := os.Open("./config.xml")
	if err != nil {
		log.Fatalf("Error opening file: %s\n", err)
	}
	defer file.Close()

	var params Parameters
	if err := xml.NewDecoder(file).Decode(&params); err != nil {
		log.Fatalf("Error decoding XML: %s\n", err)
	}

	return params
}

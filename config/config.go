/*
 * File:    config.go
 * Date:    May 11, 2024
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   handles config.xml in different struct types to make the parameters available across the whole project
 *
 */

package config

import (
	"encoding/xml"
	"log"
	"os"
)

type Parameters struct {
	XMLName  xml.Name  `xml:"parameters"`
	Build    Build     `xml:"build"`
	Config   Config    `xml:"config"`
	PostPro  PostPro   `xml:"postpro"`
	DataBase DataBase  `xml:"database"`
	Logs     LogParams `xml:"logger"`
	Graphics Graphics  `xml:"graphics"`
	Profiles Profiles  `xml:"profiles"`
}

type Build struct {
	Version string `xml:"version"`
	Type    string `xml:"type"`
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
	Cage12Split             bool    `xml:"cage12split"`
	LtcOffset				int		`xml:"ltcOffset"`
}

type DataBase struct {
	Path              string `xml:"path"`
	TimeFormatRequest string `xml:"timeFormatRequest"`
	CleaningHoursKept int    `xml:"cleaningHoursKept"`
	CleaningPeriod    int    `xml:"cleaningPeriod"`
}

type Graphics struct {
	ImageHeight       int 	 `xml:"imageHeight"`
	ImageWidth        int 	 `xml:"imageWidth"`
	ThermalScaleStart int 	 `xml:"thermalScaleStart"`
	ThermalScaleEnd   int 	 `xml:"thermalScaleEnd"`
	Savingfolder	  string `xml:"savingFolder"`
}

type Profiles struct {
	Default string	`xml:"default"`
	AS500 	string	`xml:"AS500"`
	E870NAZ	string	`xml:"3870NAZ"`	// 3870NAZ
}

func LoadConfig() (Parameters, error) {
	file, err := os.Open("./config.xml")
	if err != nil {
		log.Fatalf("Error opening file: %s\n", err)
	}
	defer file.Close()

	var params Parameters
	if err := xml.NewDecoder(file).Decode(&params); err != nil {
		log.Fatalf("Error decoding XML: %s\n", err)
	}

	// log.Println("[Parameters] OK")
	return params, nil
}

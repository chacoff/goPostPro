package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var (
	lastMESMessageType      string        = ""
	lastMESMessageTimestamp time.Time     = time.Now()
	lastMESMessageBody      []interface{} = []interface{}{0, "0", "0", 0, 0}
)

// Objects that needs to match those defined in the API

type Api_Beam_Info struct {
	BeamID                        uint32 `gorm:"primaryKey"`
	Start_LTC_message_timestamp   time.Time
	End_Postpro_message_timestamp time.Time
	RollProfile                   string
	RollNumber                    string
}

type Api_Beam_PostPro_result struct {
	BeamID                   string `gorm:"primaryKey"`
	PassNumber               int    `gorm:"primaryKey"`
	PostProStartTimestamp    time.Time
	PostProEndTimestamp      time.Time
	PostProTr1Max            int
	PostProTr1Mean           int
	PostProWebMean           int
	PostProWebMin            int
	PostProTr3Max            int
	PostProTr3Mean           int
	PostProWidth             int
	PostProThreshold         int
	LongLtcStartTimestamp    time.Time
	LongLtcEndTimestamp      time.Time
	LongLtcTr1Max            int
	LongLtcTr1Mean           int
	LongLtcWebMean           int
	LongLtcWebMin            int
	LongLtcTr3Max            int
	LongLtcTr3Mean           int
	InstantLtcStartTimestamp time.Time
	InstantLtcEndTimestamp   time.Time
	InstantLtcTr1Max         int
	InstantLtcTr1Mean        int
	InstantLtcWebMean        int
	InstantLtcWebMin         int
	InstantLtcTr3Max         int
	InstantLtcTr3Mean        int
}

type Api_DIAS_Communication struct {
	Timestamp time.Time `json:"Timestamp" gorm:"primaryKey"`
	Message   string    `json:"Message:"`
}

type Api_Line_PostPro_result struct {
	Timestamp            time.Time `json:"Timestamp" gorm:"primaryKey"`
	BeamIndexLeftBorder  int       `json:"BeamIndexLeftBorder"`
	BeamIndexRightBorder int       `json:"BeamIndexRightBorder"`
	BeamIndexLeftWeb     int       `json:"BeamIndexLeftWeb"`
	BeamIndexRightWeb    int       `json:"BeamIndexRightWeb"`
}

type Api_MES_Communication struct {
	Timestamp time.Time `gorm:"primaryKey"`
	Message   string
}

type Api_Auto_Beam_Start struct {
	Start_timestamp time.Time `gorm:"primaryKey"`
}

// Write in the switch the api destination depending on the api object sent

func SendToApi(apiObject any) {
	api_destination := "http://localhost:8765"
	switch v := apiObject.(type) {
	case Api_DIAS_Communication:
		api_destination += "/diasCommunication"
	case Api_MES_Communication:
		api_destination += "/mesCommunication"
	case Api_Beam_PostPro_result:
		api_destination += "/beamPostProResult"
	case Api_Beam_Info:
		api_destination += "/beamInfo"
	case Api_Line_PostPro_result:
		api_destination += "/linePostProResult"
	case Api_Auto_Beam_Start:
		api_destination += "/autoBeamStart"
	default:
		log.Println("[API] Unknown api object : ", v)
		return
	}
	jsonData, _ := json.Marshal(apiObject)
	_, err := http.Post(api_destination, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		//
		// log.Println("[API] Error while trying to send an object to the API : ", err)
	}

}

func HandleLTCMessageReceived(BodyValues []interface{}) {
	if len(BodyValues) < 3 {
		return
	}

	switch lastMESMessageType {
	case "ltc": // 2 consecutive LTC message = mistake to handle
		SendToApi(Api_Beam_Info{BeamID: BodyValues[0].(uint32), Start_LTC_message_timestamp: lastMESMessageTimestamp, End_Postpro_message_timestamp: time.Now(), RollProfile: BodyValues[1].(string), RollNumber: BodyValues[2].(string)})
	default:
	}
	lastMESMessageType = "ltc"
	lastMESMessageTimestamp = time.Now()
	lastMESMessageBody = BodyValues
}

func HandlePostproMessageReceived(BodyValues []interface{}) {
	switch lastMESMessageType {
	case "ltc":
		if BodyValues[0].(uint32) == lastMESMessageBody[0].(uint32) { // Case if everything went good
			SendToApi(Api_Beam_Info{BeamID: BodyValues[0].(uint32), Start_LTC_message_timestamp: lastMESMessageTimestamp, End_Postpro_message_timestamp: time.Now(), RollProfile: BodyValues[1].(string), RollNumber: BodyValues[2].(string)})
		} else { // LTC message and postpro have different beams ID -> We insert the 2 beams in the database in case
			SendToApi(Api_Beam_Info{BeamID: BodyValues[0].(uint32), Start_LTC_message_timestamp: lastMESMessageTimestamp, End_Postpro_message_timestamp: time.Now(), RollProfile: BodyValues[1].(string), RollNumber: BodyValues[2].(string)})
			SendToApi(Api_Beam_Info{BeamID: lastMESMessageBody[0].(uint32), Start_LTC_message_timestamp: lastMESMessageTimestamp, End_Postpro_message_timestamp: time.Now(), RollProfile: lastMESMessageBody[1].(string), RollNumber: lastMESMessageBody[2].(string)})
		}
	case "postpro": // 2 consecutive postpro message = mistake to handle
		SendToApi(Api_Beam_Info{BeamID: BodyValues[0].(uint32), Start_LTC_message_timestamp: lastMESMessageTimestamp, End_Postpro_message_timestamp: time.Now(), RollProfile: BodyValues[1].(string), RollNumber: BodyValues[2].(string)})
	default:
	}
	lastMESMessageType = "postpro"
	lastMESMessageTimestamp = time.Now()
	lastMESMessageBody = BodyValues
}

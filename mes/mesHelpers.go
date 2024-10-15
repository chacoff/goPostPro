/*
 * File:    mesHelpers.go
 * Date:    May 10, 2024
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   Handle decoding incoming messages from MES
 *
 */

package mes

import (
	"encoding/hex"
	"goPostPro/global"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// HandleMesData ensures the header is at least 40bytes before decoding it. It returns the body in bytes
func HandleMesData(payload []byte) ([]uint32, []byte) {
	var headerValues []uint32
	var FullLength int
	var hexBytesBody []byte

	if len(payload) >= global.AppParams.HeaderSize {
		hexBytesHeader := payload[:global.AppParams.HeaderSize] // Extract first 40 bytes, only header
		headerValues, _ = DecodeHeaderUint32(hexBytesHeader)    // decode little-endian uint32 values
		FullLength = int(headerValues[0])
		log.Println("[MES Header] >> Decoded Header values:", headerValues)
	}

	if len(payload) >= FullLength && FullLength != 0 { // TODO attention with the '>='
		hexBytesBody = payload[global.AppParams.HeaderSize:FullLength] // Extract the rest of the bytes
	}

	return headerValues, hexBytesBody
}

// HandleAnswerToMes dispatches the next process according the messageType
func HandleAnswerToMes(_headerValues []uint32, _hexBytesBody []byte) (bool, []byte, []uint16, int, uint32) {
	var echo = false
	var response []byte
	var dataLTC []uint16

	messageType := int(_headerValues[1]) // message type on the header
	messageCounter := _headerValues[2]   // already in uint32
	messageTypeAns := uint32(messageType - 100)
	lastTimestamp := getLastTimeStamp(_headerValues) // gets last timestamp for passes based on the message timestamp
	// log.Println(lastTimestamp)

	switch messageType {
	case 4701, 4711, 4721: // watchdog: only header
		log.Printf("[MES Watchdog] >> Watchdog message %d - type %d", messageCounter, messageTypeAns)
		response = encodeUint32(headerType(40, messageTypeAns, messageCounter))
		echo = true

	case 4702, 4712, 4722: // process message: header + body >> WHEN WE DO THE POST PROCESSING
		bodyValuesStatic, bodyValueDynamic := decodeBody(_hexBytesBody, messageType)
		log.Println("[MES Process] >> Decoded Body values:", bodyValuesStatic, bodyValueDynamic)

		_bodyAns := encodeProcess(processType(bodyValuesStatic, bodyValueDynamic, lastTimestamp)) // processType actually does the processing
		_length := uint32(40 + len(_bodyAns))
		_headerAns := encodeUint32(headerType(_length, messageTypeAns, messageCounter))

		_response := make([]byte, 0, len(_headerAns)+len(_bodyAns))
		_response = append(_response, _headerAns...)
		_response = append(_response, _bodyAns...)

		response = _response

		echo = true

	case 4704, 4714: // process message: header + LTC - Cage3 and Cage4 only
		// Cage 3: pass1 is AI_01 and pass3 is AI_02
		// Cage 4: pass1 is AI_01 and pass3 is AI_02

		var val1 uint16
		var val2 uint16
		var pas int

		bodyValuesStatic, _ := decodeBody(_hexBytesBody, messageType)
		log.Println("[MES LTC]  LTC received:", bodyValuesStatic)

		processID, _ := bodyValuesStatic[0].(uint32)
		LTCpass, _ := bodyValuesStatic[8].(int)

		global.ProcessID = processID
		global.LTCpass = LTCpass

		log.Printf("[PROCESS] Current Process ID is: %d", global.ProcessID)

		val := reflectToUint16(bodyValuesStatic[7]) // LTC temperature
		// pas := reflectToUint16(bodyValuesStatic[8]) // LTC pass

		if len(bodyValuesStatic) > 8 { // ugly patch STUPID PR FUCK YOU AZURE!
			pas = int(reflectToUint16(bodyValuesStatic[8])) // LTC pass
		} else {
			pas = 2 // default pass if for whatever the reason there is no pass number in the MES message
		}

		log.Printf("[MES LTC] LTC reflected to uint16 %d for pass %d\n", val, pas)

		switch int(pas) {
		case 1:
			val1 = val
			val2 = 8996
		case 3:
			val1 = 8996
			val2 = val
		default:
			val1 = 8997
			val2 = 8998
		}

		dataLTC = []uint16{500, val1, 500, val2, 44, 55, 66, 77} // DIAS Analog inputs: AI_00, AI_01, AI_02, AI_03, AI_04, AI_05, AI_06, AI_07,
		echo = false

	case 4703, 4713, 4723: // acknowledge data message
		log.Printf("[MES ACK] MES received process data properly for process ID %d", global.ProcessID)
		echo = false

	default:
		log.Println("[MES Unknown] Unknown message", messageType, messageCounter)
		echo = false
	}

	return echo, response, dataLTC, messageType, messageCounter
}

// getLastTimeStamp provides the timestamp of the message to use it as a limit for the last pass postprocessing
func getLastTimeStamp(values []uint32) string {
	// LastTimestamp is the timestamp of the message.
	// we know the sheet-pile is out of the rolling mill at this stage
	//
	// Year					_headerValues[3]
	// Month				_headerValues[4]
	// Day					_headerValues[5]
	// Hour					_headerValues[6]
	// Minute				_headerValues[7]
	// Second				_headerValues[8]
	// Hundred-of-Seconds	_headerValues[9]
	//

	datS := strings.Join([]string{strconv.FormatUint(uint64(values[3]), 10), strconv.FormatUint(uint64(values[4]), 10), strconv.FormatUint(uint64(values[5]), 10)}, "-")
	timS := strings.Join([]string{strconv.FormatUint(uint64(values[6]), 10), strconv.FormatUint(uint64(values[7]), 10), strconv.FormatUint(uint64(values[8]), 10)}, ":")

	input := strings.Join([]string{datS, timS}, " ")

	t, err := time.Parse("2006-1-2 15:4:5,99", input)
	if err != nil {
		log.Println("Error parsing input:", err)
		return "Error parsing input >>"
	}

	return t.Format(global.DBParams.TimeFormatRequest) // ISO MES
}

// reflectToUint16 reflects interface values to uint16 before sending these to DIAS-IO
func reflectToUint16(val interface{}) uint16 {

	var value uint16
	valType := reflect.TypeOf(val)

	if valType.ConvertibleTo(reflect.TypeOf(uint16(0))) {
		value = reflect.ValueOf(val).Convert(reflect.TypeOf(uint16(0))).Interface().(uint16)
		// log.Println("[LTC] LTC reflected to uint16", val)
	} else {
		value = 1432
		// log.Println("[LTC] Type not convertible", valType)
	}

	return value
}

// DataScope is used only for printing byte arrays while debugging
func DataScope(buffer []byte) (string, int) {
	return hex.EncodeToString(buffer), len(buffer)
}

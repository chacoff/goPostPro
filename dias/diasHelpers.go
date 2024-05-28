/*
 * File:    diasHelpers.go
 * Date:    May 10, 2024
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   Handle different processes to decode/encode from/to DIAS-Pyrosoft software
 *
 */

package dias

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"goPostPro/global"
	"goPostPro/postpro"
	"log"
)

// ProcessDiasData gets the payload, decode the data and process live to input the data in the DB
func ProcessDiasData(payload []byte) {

	array, digitalOutput := DecodeDiasData(payload)
	processing_list := make([][]int16, 0)

	if global.PostProParams.Cage12Split && len(array) < 513 {
		log.Println("error : not enough measures to split cage 1 / cage 2")
	}

	if global.PostProParams.Cage12Split {
		processing_list = append(processing_list, array[:500])
		processing_list = append(processing_list, array[502:])
	} else {
		processing_list = append(processing_list, array)
	}

	for _, measures := range processing_list {
		processError := postpro.Process_live_line(measures, digitalOutput)

		if errors.Is(processError, postpro.No_beam_error) {
			continue
		}
		if processError != nil {
			log.Printf("[PROCESSING] error: %s\n", processError)
		}
	}

	if global.AppParams.Verbose {
		log.Printf("[PROCESSING] completed: %d\n", array)
	}
}

// DecodeDiasData decodes the incoming data of DIAS-Pyrosoft: a block length 767 analog outputs and 4 digital outputs
func DecodeDiasData(payload []byte) ([]int16, int16) {
	message := payload
	int16_message := make([]int16, 0)

	for index := 2; index < len(message); index += 2 {
		int16_message = append(int16_message, int16(binary.LittleEndian.Uint16(message[index:index+2])))
	}

	measurementArray := int16_message[:len(int16_message)-1]
	digitalOutput := int16_message[len(int16_message)-1]

	return measurementArray, digitalOutput
}

// EncodeToDias currently DIAS-Pyrosoft is supporting 8 Analog Inputs, i.e., LTCValues is a slice of 8 elements
func EncodeToDias(LTCValues []uint16) []byte {
	answer := make([]byte, 0)

	for _, val := range LTCValues {
		binaryValue := make([]byte, 2)
		binary.LittleEndian.PutUint16(binaryValue, val)
		answer = append(answer, binaryValue...)
	}

	return answer
}

// DataScope is used only for printing byte arrays while debugging
func DataScope(buffer []byte) (string, int) {
	return hex.EncodeToString(buffer), len(buffer)
}

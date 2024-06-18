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

	array, digitalOutput, errD := DecodeDiasData(payload)

	if errD != nil{
		log.Println("[DIAS DECODER] error decoding dias data:", errD)
		return
	}

	processing_list := make([][]int16, 0)

	if global.PostProParams.Cage12Split && len(array) < 513 {
		log.Println("error : not enough measures to split cage 1 / cage 2")
	}

	if global.PostProParams.Cage12Split {
		cage12 := append(array[:500], array[502:]...)
		processing_list = append(processing_list, cage12)
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

// DecodeDiasData decodes the incoming data of DIAS-Pyrosoft: a block length 767 analog outputs and 8 digital outputs
func DecodeDiasData(payload []byte) ([]int16, int16, error) {
	
	var measurementArray []int16
	var digitalOutput int16

	if len(payload) == 0 {
		log.Printf("[DIAS DECODER] empty payload: %d\n", payload)
		return measurementArray, digitalOutput, errors.New("payload is empty")
	}
	
	if global.AppParams.Verbose{
		log.Println(payload)
	}
	
	message := payload
	int16_message := make([]int16, 0)

	for index := 2; index < len(message); index += 2 {
		val := int16(binary.LittleEndian.Uint16(message[index:index+2]))
		if global.AppParams.Verbose{
			log.Println(val)
		}
		int16_message = append(int16_message, val)
	}

	end := len(int16_message)-1

	if global.AppParams.Verbose{
		log.Printf("[DIAS DECODER] end of Array: %d\n", end)
	}

	measurementArray = int16_message[:end]
	digitalOutput = int16_message[end]

	return measurementArray, digitalOutput, nil
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

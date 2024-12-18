/*
 * File:    diasInterface.go
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
	"fmt"
	"goPostPro/global"
	"goPostPro/postpro"
	"log"

	"strconv"
	"strings"
)

// Digital outputs of DIAS
type DigitalOutputs struct {
	Pressence bool // Pressence
	Pass1     bool // Pass 1 presence
	Pass2     bool // Pass 2 presence
	Pass3     bool // Pass 3 presence
	Free1     bool // go LTC pass 1
	Free2     bool // go LTC pass 3
	Free3     bool // moving Presence
	Free4     bool
	CamLine   bool // Line status
	CamTemp   bool // Temperature status
	CamStatus bool // Camera Status
	CamConn   bool // Camera connection
}

// Average temperatures measure to compare with LTC
type AnalogsVOIs struct {
	AO_768 uint32
	AO_769 uint32
}

var Outputs DigitalOutputs
var Analogs AnalogsVOIs
var previousPassname string

// ProcessDiasData gets the payload, decode the data and process live to input the data in the DB
func ProcessDiasData(payload []byte) {

	var array []int16
	var isMoving int

	full_array, digitalOutput, errD := DecodeDiasData(payload)
	if errD != nil {
		log.Println("[DIAS DECODER] error decoding dias data:", errD)
		return
	}

	Outputs.decodeDiasDigitalOutput(digitalOutput)
	passname, _ := determine_passname()

	if !(passname == previousPassname) {
		log.Printf("[DIAS] %s - Process id %d", passname, global.ProcessID)
		previousPassname = passname
	}

	processing_list := make([][]int16, 0)

	if global.PostProParams.Cage12Split {
		array = full_array
		cage12 := append(array[:500], array[502:]...)
		processing_list = append(processing_list, cage12)
	} else { // else is cage 3 or cage4
		array = full_array[0:767]                      // 0:767 are the measurements array block
		Analogs.updateAnalogsVOIs(full_array[767:769]) // last 2 elements are VOIs
		processing_list = append(processing_list, array)
	}

	if global.PostProParams.Cage12Split && len(array) < 513 {
		log.Println("error : not enough measures to split cage 1 / cage 2")
	}

	for _, measures := range processing_list {

		if global.AppParams.Verbose {
			log.Printf("[PROCESSING] pass number: %s", passname)
		}

		isMoving = btoi(Outputs.Free3) // Moving Presence flag from DIAS

		processError := postpro.Process_live_line(measures, passname, isMoving)

		if errors.Is(processError, postpro.NoBeamError) {
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

// DecodeDiasData decodes the incoming data of DIAS-Pyrosoft: a block length, for cage3 and cage4 of 770 (AO_0 to AO_769) analog outputs and 12 digital outputs
func DecodeDiasData(payload []byte) ([]int16, int16, error) {
	// Length of DIAS = (digital outputs - 1)/8 + 1 + amount of analog outputs * 2
	// AO_00 is skipped because contains the length of the payload (index = 2)
	// AO_767 is the last element of the measurements array
	// AO_768 and AO_769 are free VOIs

	var measurementArray []int16
	var digitalOutput int16

	if len(payload) == 0 {
		log.Printf("[DIAS DECODER] empty payload: %d\n", payload)
		return measurementArray, digitalOutput, errors.New("payload is empty")
	}

	message := payload
	int16_message := make([]int16, 0)

	// index = 2 instead of instead of 0 because we skip the first element because is the length of the array
	for index := 2; index < len(message); index += 2 {
		val := int16(binary.LittleEndian.Uint16(message[index : index+2]))
		int16_message = append(int16_message, val)
	}

	end := len(int16_message) - 1

	measurementArray = int16_message[:end]
	digitalOutput = int16_message[end]

	if global.AppParams.Verbose {
		ui := hex.EncodeToString(payload)
		log.Println(ui)
		fmt.Println(len(measurementArray))
		fmt.Println(measurementArray[0])   // element 0 is AO_01
		fmt.Println(measurementArray[768]) // element 768 is AO_769
		fmt.Println(digitalOutput)
	}

	return measurementArray, digitalOutput, nil
}

// EncodeToDias currently DIAS-Pyrosoft is supporting 12 Analog Inputs, i.e., LTCValues is a slice of 8 elements
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

// btoi Boolean to integer helper function. Returns 1 if the IO boolean is active and 0 if inactive
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// determine_passname
func determine_passname() (string, error) {

	global.SaveImage = false

	if Outputs.Pass3 && !Outputs.Pass2 && !Outputs.Pass1 {
		global.PreviousPassNumber = 3
		global.CurrentPass = "Pass 3"
		return "Pass 3", nil
	}

	if Outputs.Pass2 && !Outputs.Pass1 && !Outputs.Pass3 {
		global.PreviousPassNumber = 2
		global.CurrentPass = "Pass 2"
		return "Pass 2", nil
	}

	if Outputs.Pass1 && !Outputs.Pass2 && !Outputs.Pass3 {
		if global.PreviousPassNumber == 3 {
			global.SaveImage = true
		}
		global.PreviousPassNumber = 1
		global.CurrentPass = "Pass 1"
		return "Pass 1", nil
	}

	// log.Println("[DIAS] Attention: pass number couldn't be define, using global.CurrentPass")
	return global.CurrentPass, errors.New("using global.CurrentPass. Most likely an IO error")
}

// decodeDigitalOutput gets the decimal value sent from DIAS and convert it to its binary representation to fill DigitalOutputs Struct
func (d *DigitalOutputs) decodeDiasDigitalOutput(digits int16) {

	var n int64
	var nbin string
	var nbinSlice []string

	n = int64(digits)
	nbin = strconv.FormatInt(n, 2)

	nbinSlice = strings.Split(nbin, "")

	if global.AppParams.Verbose {
		log.Printf("[DIAS] Digital outputs: %s\n", nbin)
		log.Printf("[DIAS] Digital outputs: %s\n", nbinSlice)
	}

	if len(nbinSlice) < 12 {
		nbinSlice = []string{"0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0"}
	}

	d.Pressence, _ = strconv.ParseBool(nbinSlice[11])
	d.Pass1, _ = strconv.ParseBool(nbinSlice[10])
	d.Pass2, _ = strconv.ParseBool(nbinSlice[9])
	d.Pass3, _ = strconv.ParseBool(nbinSlice[8])
	d.Free1, _ = strconv.ParseBool(nbinSlice[7])
	d.Free2, _ = strconv.ParseBool(nbinSlice[6])
	d.Free3, _ = strconv.ParseBool(nbinSlice[5])
	d.Free4, _ = strconv.ParseBool(nbinSlice[4])
	d.CamLine, _ = strconv.ParseBool(nbinSlice[3])
	d.CamTemp, _ = strconv.ParseBool(nbinSlice[2])
	d.CamStatus, _ = strconv.ParseBool(nbinSlice[1])
	d.CamConn, _ = strconv.ParseBool(nbinSlice[0])
}

// updateAnalogsVOIs helps to keep up to the date the struct with analog VOIs (in our applicationsa are the LTCs from DIAS)
func (a *AnalogsVOIs) updateAnalogsVOIs(analogs []int16) {
	a.AO_768 = uint32(analogs[0])
	a.AO_769 = uint32(analogs[1])
}

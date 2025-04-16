/*
 * File:    processMessages.go
 * Date:    October 25, 2024
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 *
 * Description:
 *		Process message including 2 methods to calculate postprocessing and LTC.
 *		Either using MES/PC-Couple timestamps and calculating the timestamps with the Dias flag: << moving >>
 *
 *	PROTOCOL:
 *
 *		n is number of passes
 *		Position of each element is very relevant in encodeProcess()
 *
 *		j	DESCRIPTION						TYPE	RELATIVE POSITION
 *		0	Unique product ID				UINT32	0
 *		1	Rolling campaign profile		STRING	1
 *		2	Rolling campaign number			STRING	2
 *		3	Roll stand number				UINT32	3
 *		4	Pass counter					UINT32	4
 *		5	Pass number n					UINT32 	5 LOOP STARTS HERE
 * 		6	Pass date n						STRING 	6 31 56 81 106 131 156 181 206 231 256 281 306
 *		7	Dummy							STRING 	7 32 57 82 107 132 157 182 207 232 257 282 307

	POST PROCESSING (ALL DATA OF THE PASS)
 *		8	Max Temp mill3 pass n			UINT32	8
 *		9	Avg Temp mill3 pass n			UINT32	9
 *		10	Max Temp mill1 pass n			UINT32	10
 *		11	Avg Temp mill1 pass n			UINT32	11
 *		12	Min Temp web pass n				UINT32	12
 *		13	Avg Temp web pass n				UINT32	13
 *		14	Avg STD pass n					UINT32	14
 *		15	Pix width pass n				UINT32	15

	LTC POST PROCESSING (POST PROCESSING FROM MOVING FLAG, all the bar after waiting (+40))
 *		16	Max Temp mill3 pass n LTC		UINT32	16
 *		17	Avg Temp mill3 pass n LTC		UINT32	17
 *		18	Max Temp mill1 pass n LTC		UINT32	18
 *		19	Avg Temp mill1 pass n LTC		UINT32	19
 *		20	Min Temp web pass n LTC			UINT32	20
 *		21	Avg Temp web pass n LTC			UINT32	21

	LTC REALIZED WHICH IS MAX TEMP MILL3 LTC (position 23 is the same as position 16)
 *		22	LTC Pass number pass n			UINT32	22
 *		23	LTC Realized pass n				UINT32	23

	FIRST LTC ROWS (2 measures)
 *      24 Max Temp mill3 pass n FirstLTC   UINT32	24
 *      25 Avg Temp mill3 pass n FirstLTC   UINT32	25
 *      26 Max Temp mill1 pass n FirstLTC   UINT32	26
 *      27 Avg Temp mill1 pass n FirstLTC   UINT32	27
 *      28 Min Temp web pass n FirstLTC     UINT32	28
 *      29 Avg Temp web pass n FirstLTC     UINT32	29
*/

package mes

import (
	"fmt"
	"goPostPro/api"
	"goPostPro/global"
	"goPostPro/klusters"
	"goPostPro/postpro"
	"log"
)

// processType return the real values to answer process messages according the number of passes
func processType(_bodyStatic []interface{}, _bodyDynamic []interface{}, lastTimeStamp string) []interface{} {

	var _bodyAns []interface{}
	var newData postpro.PostProData
	var err error
	var ltcTimestamp string
	var beginStamp string
	var endStamp string

	beamId := _bodyStatic[0].(uint32)      // Beam ID
	passCounter := _bodyStatic[4].(uint32) // Pass counter

	listOfStamps := parseTimeStamps(passCounter, _bodyDynamic, lastTimeStamp) // passDates are available in positions 1, 4, 7, 10, 13, 16, 19 ... = pass+(i*2)
	log.Printf("[PostPro] BeamID %d Process timestamps %s", beamId, listOfStamps)

	_bodyAns = append(_bodyAns, _bodyStatic[0]) // unique product ID
	_bodyAns = append(_bodyAns, _bodyStatic[1]) // rolling campaign profile
	_bodyAns = append(_bodyAns, _bodyStatic[2]) // rolling campaign number
	_bodyAns = append(_bodyAns, _bodyStatic[3]) // roll stand number
	_bodyAns = append(_bodyAns, _bodyStatic[4]) // pass counter

	for i := 0; i < int(passCounter); i++ {

		if global.PostProParams.Cage12Split {
			beginStamp = listOfStamps[i]
			endStamp = listOfStamps[i+1]
		} else {
			beginStamp = global.PreviousLastTimeStamp
			endStamp = lastTimeStamp

			fmt.Printf("- Calling a new reCluster between %s and %s for Beam %d\n", beginStamp, endStamp, beamId)
			klusters.ReClusterPasses(beginStamp, endStamp, beamId)
		}

		// Standard post processing data
		log.Printf("[PostPro] BeamID %d Pass: %d/%d between timestamps %s - %s", beamId, i+1, passCounter, beginStamp, endStamp)
		newData, err = postpro.DATABASE.QueryDatabase(beginStamp, endStamp, i, false) // TODO : Add rerun offset

		if err != nil {
			log.Println("ERROR : ", err)
		}

		newData.PassNumber = uint32(i + 1)
		newData.PassDate = listOfStamps[i] // time.Now().Format("20060102150405")
		newData.Dummy = "du"

		_bodyAns = append(_bodyAns, newData.PassNumber)
		_bodyAns = append(_bodyAns, newData.PassDate)
		_bodyAns = append(_bodyAns, newData.Dummy)
		_bodyAns = append(_bodyAns, newData.MaxTempMill3)
		_bodyAns = append(_bodyAns, uint32(newData.AvgTempMill3))
		_bodyAns = append(_bodyAns, newData.MaxTempMill1)
		_bodyAns = append(_bodyAns, uint32(newData.AvgTempMill1))
		_bodyAns = append(_bodyAns, newData.MinTempWeb)
		_bodyAns = append(_bodyAns, uint32(newData.AvgTempWeb))
		_bodyAns = append(_bodyAns, uint32(newData.AvgStdTemp))
		_bodyAns = append(_bodyAns, uint32(newData.PixWidth))

		log.Printf("[PostPro] BeamID %d Pass: %d/%d partial PostPro answer: %v", beamId, i+1, passCounter, _bodyAns)

		log.Println("[PostPro LTC Cage3-4] Calling FindLTCRow with:", beginStamp, endStamp)
		ltcTimestamp = postpro.DATABASE.FindLTCRow(beginStamp, endStamp, i)

		if ltcTimestamp == "" {
			ltcTimestamp = listOfStamps[i] // if no LTC is found, it uses the original method of MES timestamp
			log.Println("[PostPro LTC] LTC timestamp is empty, using MES/PC-Couple timestamp:")
		}

		ltcTimestamp_begin := addOffsetToTimestamp(ltcTimestamp, min(0, global.PostProParams.LtcOffset))
		ltcTimestamp_end := addOffsetToTimestamp(ltcTimestamp, max(0, global.PostProParams.LtcOffset))

		log.Printf("[PostPro LTC] BeamID %d Pass: %d/%d between timestamps %s - %s", beamId, i+1, passCounter, ltcTimestamp_begin, ltcTimestamp_end)

		ltcData, errLtc := postpro.DATABASE.QueryDatabase(ltcTimestamp_begin, ltcTimestamp_end, i, false)

		if errLtc != nil {
			log.Println("ERROR : ", errLtc)
		}

		_bodyAns = append(_bodyAns, ltcData.MaxTempMill3)
		_bodyAns = append(_bodyAns, uint32(ltcData.AvgTempMill3))
		_bodyAns = append(_bodyAns, ltcData.MaxTempMill1)
		_bodyAns = append(_bodyAns, uint32(ltcData.AvgTempMill1))
		_bodyAns = append(_bodyAns, ltcData.MinTempWeb)
		_bodyAns = append(_bodyAns, uint32(ltcData.AvgTempWeb))

		_bodyAns = append(_bodyAns, newData.PassNumber)   // LTC Realized
		_bodyAns = append(_bodyAns, ltcData.MaxTempMill3) // LTC Realized

		log.Printf("[PostPro LTC] BeamID %d Pass: %d/%d partial PostPro answer with LTC: %v", beamId, i+1, passCounter, _bodyAns)

		ltcTimestampFirst := addOffsetToTimestamp(ltcTimestamp, 0)
		ltcTimestampFirstOffset := addOffsetToTimestamp(ltcTimestamp, 1)
		log.Printf("[First LTC row] BeamID %d Pass: %d/%d - First LTC between: %s - %s", beamId, i+1, passCounter, ltcTimestampFirst, ltcTimestampFirstOffset)
		firstLtc, errFirst := postpro.DATABASE.QueryDatabase(ltcTimestampFirst, ltcTimestampFirstOffset, i, true)

		if errFirst != nil {
			log.Println("ERROR : ", errFirst)
		}

		_bodyAns = append(_bodyAns, firstLtc.MaxTempMill3)
		_bodyAns = append(_bodyAns, uint32(firstLtc.AvgTempMill3))
		_bodyAns = append(_bodyAns, firstLtc.MaxTempMill1)
		_bodyAns = append(_bodyAns, uint32(firstLtc.AvgTempMill1))
		_bodyAns = append(_bodyAns, firstLtc.MinTempWeb)
		_bodyAns = append(_bodyAns, uint32(firstLtc.AvgTempWeb))

		log.Printf("[PostPro LTC] BeamID %d Pass: %d/%d partial PostPro answer with LTC and First LTC: %v", beamId, i+1, passCounter, _bodyAns)

		// Send the postpro result to the API
		api_object_postpro_result := api.Api_Beam_PostPro_result{
			BeamID: fmt.Sprint(beamId), PassNumber: i + 1,
			PostProStartTimestamp: newData.FirstTimestampDatabaseTimeObject, PostProEndTimestamp: newData.LastTimeStampDatabaseTimeObject, PostProTr1Max: int(newData.MaxTempMill1), PostProTr1Mean: int(newData.AvgTempMill1), PostProWebMean: int(newData.AvgTempWeb), PostProWebMin: int(newData.MinTempWeb), PostProTr3Max: int(newData.MaxTempMill3), PostProTr3Mean: int(newData.AvgTempMill3), PostProWidth: int(newData.PixWidth), PostProThreshold: int(newData.MeanThreshold),
			LongLtcStartTimestamp: ltcData.FirstTimestampDatabaseTimeObject, LongLtcEndTimestamp: ltcData.LastTimeStampDatabaseTimeObject, LongLtcTr1Max: int(ltcData.MaxTempMill1), LongLtcTr1Mean: int(ltcData.AvgTempMill1), LongLtcWebMean: int(ltcData.AvgTempWeb), LongLtcWebMin: int(ltcData.MinTempWeb), LongLtcTr3Max: int(ltcData.MaxTempMill3), LongLtcTr3Mean: int(ltcData.AvgTempMill3),
			InstantLtcStartTimestamp: firstLtc.FirstTimestampDatabaseTimeObject, InstantLtcEndTimestamp: firstLtc.LastTimeStampDatabaseTimeObject, InstantLtcTr1Max: int(firstLtc.MaxTempMill1), InstantLtcTr1Mean: int(firstLtc.AvgTempMill1), InstantLtcWebMean: int(firstLtc.AvgTempWeb), InstantLtcWebMin: int(firstLtc.MinTempWeb), InstantLtcTr3Max: int(firstLtc.MaxTempMill3), InstantLtcTr3Mean: int(firstLtc.AvgTempMill3),
		}
		api.SendToApi(api_object_postpro_result)

	}

	// @jaime: TODO, marked as Treated all rows between first and last timestamp
	// _, _ = postpro.DATABASE.UpdateTreated(listOfStamps[i], listOfStamps[i+1])

	log.Printf("[PostPro] BeamID %d Final PostPro answer: %v", beamId, _bodyAns)
	global.PreviousLastTimeStamp = lastTimeStamp
	return _bodyAns
}

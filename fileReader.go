/*
* go get -u github.com/go-gota/gota/...
 */

package main

import (
	"bufio"
	"fmt"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	_file, err := os.Open("D:\\00_Dev\\RD_AM\\PostProTrain2\\testing_data\\DUO01-02_0891_half.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer _file.Close()

	scanner := bufio.NewScanner(_file)
	scanner.Scan() // skip first line

	df := createDataFrame(scanner)
	fmt.Println(df)
}

func createDataFrame(scanner *bufio.Scanner) dataframe.DataFrame {
	var stamps []string
	var temps [][]float64
	var _stamps string
	var _dataLine []string
	var _dataLineConcat string

	for scanner.Scan() {
		_line := strings.ReplaceAll(scanner.Text(), ",", ".") // Replace commas with periods, sometimes!
		line := strings.Fields(_line)                         // like str.join.(' ')

		_stamps = line[0] + " " + line[1]
		_dataLine = line[7 : len(line)-4] // extract temperature data, elements from 7 to (total-4)

		_dataLineConcat = strings.Join(_dataLine, " ")
		_tempsArray := dataFormatter(_dataLineConcat)
		stamps = append(stamps, _stamps)
		temps = append(temps, _tempsArray)
	}

	tempsStrings := arrayFloatToString(temps)

	df := dataframe.New(
		series.New(stamps, series.String, "TimeStamps"),
		series.New(tempsStrings, series.String, "Temperatures"),
	)

	return df
}

func arrayFloatToString(temps [][]float64) []string {
	var tempsStrings []string
	for _, t := range temps {
		tempsStrings = append(tempsStrings, arrayToString(t))
	}
	return tempsStrings
}

func arrayToString(arr []float64) string {
	strArr := make([]string, len(arr))
	for i, v := range arr {
		strArr[i] = fmt.Sprintf("%f", v)
	}
	return strings.Join(strArr, " ")
}

func timeFormatter(s string) time.Time {
	// convert string to timestamp

	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		panic(err)
	}
	return t
}

func dataFormatter(dataAsString string) []float64 {
	// convert data read from the TXT to a slice of floats64

	splitString := strings.Fields(dataAsString)

	floatSlice := make([]float64, len(splitString))

	for i, str := range splitString {
		f, err := strconv.ParseFloat(str, 64)
		if err != nil {
			fmt.Printf("Error parsing float: %v\n", err)
		}

		floatSlice[i] = f
	}

	return floatSlice
}

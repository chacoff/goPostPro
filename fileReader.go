/*
* go get -u github.com/go-gota/gota/...
 */

package main

import (
	"bufio"
	"fmt"
	"github.com/go-gota/gota/dataframe"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type dataFrame struct {
	TimeStamp string // pay attention to define dataFrame struct with exported fields, capital letter first!!
	DataTemps string
}

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
	var datas []dataFrame

	for scanner.Scan() {
		_line := strings.ReplaceAll(scanner.Text(), ",", ".") // Replace commas with periods, sometimes!
		line := strings.Fields(_line)                         // like str.join.(' ')
		_dataLine := line[7:10]                               // extract temperature data, elements from 7 to (total-4)
		_dataLineConcat := strings.Join(_dataLine, " ")       // because dataframe.LoadStructs doesn't support []string []float64

		_df := dataFrame{
			line[0] + " " + line[1], // extract timestamp, elements 0 and 1
			_dataLineConcat,
		}
		datas = append(datas, _df)
	}

	df := dataframe.LoadStructs(datas)

	return df
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

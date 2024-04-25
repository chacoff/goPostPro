package postpro

import (
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"goPostPro/global"
)

var DATABASE CalculationsDatabase = CalculationsDatabase{}

type LineProcessing struct {
	// Reduce sizes for efficiency ?
	filename                     string
	timestamp                    time.Time
	processed_temperatures_array []float64
	gradient_temperatures_array  []float64
	max_Tr1                      float64
	mean_Tr1                     float64
	mean_Web                     float64
	min_Web                      float64
	max_Tr3                      float64
	mean_Tr3                     float64
	width                        int64
	threshold                    float64
	gradient_limit               float64
}

// Take the string received and parse its datas
func (line_processing *LineProcessing) parse_string_received(string_received string) error {
	splited_string_array := strings.Split(string_received, "\t")
	if len(splited_string_array) < 516 {
		log.Println("There is an error : a line isnt long enough")
	}
	// Parse timestamp
	timestamp, parsing_error := time.Parse(global.TIME_FORMAT, splited_string_array[0])
	line_processing.timestamp = timestamp
	if parsing_error != nil {
		return parsing_error
	}
	// Removed the unwanted measures
	measures_string_array := append(
		splited_string_array[1+global.NUMBER_FIRST_MEASURES_REMOVED:500],
		splited_string_array[510:len(splited_string_array)-4]...,
	)
	// Parse temperature measures and by the same time find the max and min
	var max_temperature float64
	var min_temperature float64
	line_processing.processed_temperatures_array = make([]float64, len(measures_string_array))
	for index, temperature_string := range measures_string_array {
		temperature_string = strings.ReplaceAll(temperature_string, ",", ".")
		temperature_float, parsing_error := strconv.ParseFloat(temperature_string, 64)
		if parsing_error != nil {
			return parsing_error
		}
		if index == 0 {
			max_temperature = temperature_float
			min_temperature = temperature_float
		}
		max_temperature = math.Max(max_temperature, temperature_float)
		min_temperature = math.Min(min_temperature, temperature_float)
		line_processing.processed_temperatures_array[index] = temperature_float
	}
	// Calcul the threshold that will be used
	line_processing.threshold = math.Max(
		min_temperature*(1-global.TEMPERATURE_THRESHOLD_FACTOR)+max_temperature*global.TEMPERATURE_THRESHOLD_FACTOR,
		global.TEMPERATURE_THRESHOLD_MINIMUM,
	)
	return nil
}

// Threshold the temperatures and compute the gradient array
func (line_processing *LineProcessing) threshold_compute_gradient() error {
	line_processing.gradient_temperatures_array = make([]float64, len(line_processing.processed_temperatures_array))
	line_processing.gradient_temperatures_array[0] = 0
	max_gradient := float64(0)
	for index, temperature_float := range line_processing.processed_temperatures_array {
		// Threshold the temperature
		temperature_thresholded := math.Max(line_processing.threshold, temperature_float)
		line_processing.processed_temperatures_array[index] = temperature_thresholded
		// Compute the gradient
		if index > 0 {
			gradient_temperature := math.Abs(temperature_thresholded - line_processing.processed_temperatures_array[index-1])
			line_processing.gradient_temperatures_array[index] = gradient_temperature
			max_gradient = math.Max(max_gradient, gradient_temperature)
		}
	}
	line_processing.gradient_limit = max_gradient / global.GRADIENT_LIMIT_FACTOR
	return nil
}

// Find the lower and higher index where the gradient is above the limit and crop all arrays to keep the values in between
func (line_processing *LineProcessing) gradient_cropping() error {
	lower_index_crop := int(0)
	higher_index_crop := int(0)
	for index, gradient_temperature := range line_processing.gradient_temperatures_array {
		if gradient_temperature > line_processing.gradient_limit {
			if lower_index_crop == 0 {
				lower_index_crop = index
			}
			higher_index_crop = index
		}
	}
	line_processing.processed_temperatures_array = line_processing.processed_temperatures_array[lower_index_crop:higher_index_crop]
	line_processing.gradient_temperatures_array = line_processing.gradient_temperatures_array[lower_index_crop:higher_index_crop]
	line_processing.width = int64(higher_index_crop - lower_index_crop)
	return nil
}

func (line_processing *LineProcessing) compute_calculations() error {
	half_index := int((line_processing.width+1)/2 - 1)
	filtered_temperature_array := line_processing.processed_temperatures_array
	//Max, Mean of Tr1
	sum_Tr1 := float64(0)
	max_index_Tr1 := int64(0)
	max_Tr1 := float64(0)
	for index_Tr1 := 0; index_Tr1 <= half_index; index_Tr1++ {
		temperature_float := filtered_temperature_array[index_Tr1]
		sum_Tr1 += temperature_float
		if temperature_float > max_Tr1 {
			max_Tr1 = temperature_float
			max_index_Tr1 = int64(index_Tr1)
		}
	}
	line_processing.max_Tr1 = max_Tr1
	line_processing.mean_Tr1 = sum_Tr1 / float64(half_index+1)
	//Max, Mean of Tr3
	sum_Tr3 := float64(0)
	max_index_Tr3 := int64(0)
	max_Tr3 := float64(0)
	for index_Tr3 := half_index + 1; index_Tr3 < len(filtered_temperature_array); index_Tr3++ {
		temperature_float := filtered_temperature_array[index_Tr3]
		sum_Tr3 += temperature_float
		if temperature_float >= max_Tr3 {
			max_Tr3 = temperature_float
			max_index_Tr3 = int64(index_Tr3)
		}
	}
	line_processing.max_Tr3 = max_Tr3
	line_processing.mean_Tr3 = sum_Tr3 / float64(len(filtered_temperature_array)-half_index)
	//Min, Mean Web
	sum_Web := float64(0)
	min_Web := filtered_temperature_array[max_index_Tr1]
	for index_Web := max_index_Tr1; index_Web <= max_index_Tr3; index_Web++ {
		temperature_float := filtered_temperature_array[index_Web]
		sum_Web += temperature_float
		if temperature_float < min_Web {
			min_Web = temperature_float
		}
	}
	line_processing.min_Web = min_Web
	line_processing.mean_Web = sum_Web / float64(max_index_Tr3-max_index_Tr1)
	return nil
}

// Receive the string for a single line, process it and save the values in the database
func Process_line(string_received string, filename string) error {
	var line_processing LineProcessing
	line_processing.filename = filename

	parsing_error := line_processing.parse_string_received(string_received)
	if parsing_error != nil {
		return parsing_error
	}
	threshold_gradient_error := line_processing.threshold_compute_gradient()
	if threshold_gradient_error != nil {
		return parsing_error
	}
	cropping_error := line_processing.gradient_cropping()
	if cropping_error != nil {
		return parsing_error
	}
	if line_processing.width > global.WIDTH_MINIMUM {
		computing_error := line_processing.compute_calculations()
		if computing_error != nil {
			return parsing_error
		}
		database_line := Database_Line{}
		database_line.Import_line_processing(line_processing)
		process_error := DATABASE.Insert_line(database_line)
		if process_error != nil {
			return process_error
		}
	}

	return nil
}

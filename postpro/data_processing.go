package postpro

import (
	"errors"
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
		return errors.New("error : the parsed line has not enough measures")
	}
	// Parse timestamp
	timestamp, parsing_error := time.Parse(global.PostProParams.TimeFormat, splited_string_array[0])
	if parsing_error != nil {
		return parsing_error
	}
	line_processing.timestamp = timestamp
	// Removed the unwanted measures
	measures_string_array := append(
		splited_string_array[1+global.PostProParams.FirstMeasuresRemoved:500],
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
		if index == 0 { // To initialize the values
			max_temperature = temperature_float
			min_temperature = temperature_float
		}
		max_temperature = math.Max(max_temperature, temperature_float)
		min_temperature = math.Min(min_temperature, temperature_float)
		line_processing.processed_temperatures_array[index] = temperature_float
	}
	// Calcul the threshold that will be used
	line_processing.threshold = math.Max(
		min_temperature*(1-global.PostProParams.AdaptativeFactor)+max_temperature*global.PostProParams.AdaptativeFactor,
		global.PostProParams.MinTemperatureThreshold,
	)
	return nil
}

func (line_processing *LineProcessing) clean_int_received(int_array []int16) error {
	if len(int_array) < 516 {
		return errors.New("error : the parsed line has not enough measures")
	}
	line_processing.timestamp = time.Now()
	measures_int_array := append(
		int_array[1+global.PostProParams.FirstMeasuresRemoved:500],
		int_array[510:len(int_array)-4]...,
	)
	var max_temperature float64
	var min_temperature float64
	line_processing.processed_temperatures_array = make([]float64, len(measures_int_array))
	for index, temperature_int := range measures_int_array {
		temperature_float := float64(temperature_int)
		if index == 0 { // To initialize the values
			max_temperature = temperature_float
			min_temperature = temperature_float
		}
		max_temperature = math.Max(max_temperature, temperature_float)
		min_temperature = math.Min(min_temperature, temperature_float)
		line_processing.processed_temperatures_array[index] = temperature_float
	}
	// Calcul the threshold that will be used
	line_processing.threshold = math.Max(
		min_temperature*(1-global.PostProParams.AdaptativeFactor)+max_temperature*global.PostProParams.AdaptativeFactor,
		global.PostProParams.MinTemperatureThreshold,
	)
	return nil
}

// Threshold the temperatures and compute the gradient array
func (line_processing *LineProcessing) threshold_compute_gradient() error {
	if len(line_processing.processed_temperatures_array) < 2 {
		return errors.New("error : the processed temperatures line has not enough measures")
	}
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
	if global.PostProParams.GradientFactor <= 0 {
		return errors.New("error : the gradient limit factor is not valid")
	}
	line_processing.gradient_limit = max_gradient / global.PostProParams.GradientFactor
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
	// QUESTION : Should we take one processed temperature before to have the values that led to the first gradient?
	line_processing.processed_temperatures_array = line_processing.processed_temperatures_array[lower_index_crop:higher_index_crop]
	line_processing.gradient_temperatures_array = line_processing.gradient_temperatures_array[lower_index_crop:higher_index_crop]
	line_processing.width = int64(len(line_processing.processed_temperatures_array))
	return nil
}

func (line_processing *LineProcessing) compute_calculations() error {
	if line_processing.width < 2 {
		return errors.New("error : not enough measures for calculation")
	}
	half_index := int((line_processing.width+1)/2 - 1) // Round the half to the upper value (+1) and then convert in 0-indexed index (-1)
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
	line_processing.mean_Tr3 = sum_Tr3 / float64(len(filtered_temperature_array)-(half_index+1))
	//Min, Mean Web
	if max_index_Tr1 == max_index_Tr3 {
		return errors.New("error : there is a problem in the max indexes for web calculation")
	}
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
	line_processing.mean_Web = sum_Web / float64(max_index_Tr3-max_index_Tr1+1)
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
		return threshold_gradient_error
	}
	cropping_error := line_processing.gradient_cropping()
	if cropping_error != nil {
		return cropping_error
	}
	if line_processing.width > global.PostProParams.MinWidth {
		computing_error := line_processing.compute_calculations()
		if computing_error != nil {
			return computing_error
		}
		insertion_error := DATABASE.Insert_line_processing(line_processing)
		if insertion_error != nil {
			return insertion_error
		}
	}

	return nil
}

func Process_live_line(int_array_received []int16) error {
	var line_processing LineProcessing
	line_processing.filename = "Live_Recording"

	parsing_error := line_processing.clean_int_received(int_array_received)
	if parsing_error != nil {
		return parsing_error
	}
	threshold_gradient_error := line_processing.threshold_compute_gradient()
	if threshold_gradient_error != nil {
		return threshold_gradient_error
	}
	cropping_error := line_processing.gradient_cropping()
	if cropping_error != nil {
		return cropping_error
	}
	if line_processing.width > global.PostProParams.MinWidth {
		computing_error := line_processing.compute_calculations()
		if computing_error != nil {
			return computing_error
		}
		insertion_error := DATABASE.Insert_line_processing(line_processing)
		if insertion_error != nil {
			return insertion_error
		}
	}

	return nil
}

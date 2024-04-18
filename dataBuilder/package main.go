package main

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	_ "image"
	"image/color"
	_ "image/color"
	"image/png"
	_ "image/png"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)
// Variables used to print the image
var(
    SEUIL_MAX_IMAGE = 1200
    LINE_COUNT int = -1
    FINAL_IMAGE *image.RGBA = image.NewRGBA(image.Rect(0, 0, 1200, 1200))
    local_lower_index int = 0
)

var(
    PROCESSED_LINES []KeyValues
    NUMBER_FIRST_MEASURES_REMOVED int = 5
    TIME_FORMAT string = "2006-01-02 15:04:05,999"
    TIME_FORMAT_REQUESTS string = "2006-01-02 15:04:05"
    TEMPERATURE_THRESHOLD float64 = 800
    GRADIENT_LIMIT_FACTOR float64 = 3
    WIDTH_LIMIT int32 = 2
)

// Contains the important value of a line
type KeyValues struct{
    Timestamp time.Time
    Max_Tr1 float64
    Mean_Tr1 float64
    Mean_Web float64
    Min_Web float64
    Max_Tr3 float64
    Mean_Tr3 float64
    Width int32
}

// Overwrite the printing function
func (values KeyValues) String() string{
    return fmt.Sprint(math.Round(values.Max_Tr1), math.Round(values.Mean_Tr1), math.Round(values.Mean_Web), math.Round(values.Min_Web), math.Round(values.Max_Tr3), math.Round(values.Mean_Tr3), math.Round(float64(values.Width)))
    //return fmt.Sprint(values.Timestamp.Format("15:04:05.000 "),values.Max_Tr1, values.Mean_Tr1, values.Mean_Web, values.Min_Web, values.Max_Tr3, values.Mean_Tr3, values.Width)
}

// To reset the list of KeyValues 
func reset_dataframe(){
    PROCESSED_LINES = []KeyValues{}
}

// Returns the indexs to use based on the timestamps given
func find_index_timestamps(begin_timestamp time.Time, end_timestamp time.Time)(int32, int32, error){
    begin_index := int32(0)
    end_index := int32(-1)
    for index, key_value := range PROCESSED_LINES {
        if (key_value.Timestamp.Before(begin_timestamp)){
            begin_index = int32(index) + 1
        }
        if !(key_value.Timestamp.After(end_timestamp)){
            end_index = int32(index)
        }
    }
    if ((int(begin_index)>=len(PROCESSED_LINES))||(end_index==-1)){
        return -1, -1, errors.New("one or two timestamps are not inside datas")
    }
    return begin_index, end_index, nil
}

// Compute the global values between two timestamps given
func compute_values(begin_string_timestamp string, end_string_timestamp string)error{
    // Convert the timestamps into indexes 
    begin_timestamp, parsing_error := time.Parse(TIME_FORMAT_REQUESTS, begin_string_timestamp)
    if (parsing_error != nil){
        return parsing_error
    }
    end_timestamp, parsing_error := time.Parse(TIME_FORMAT_REQUESTS, end_string_timestamp)
    if (parsing_error != nil){
        return parsing_error
    }
    begin_index, end_index, timestamps_error := find_index_timestamps(begin_timestamp, end_timestamp)
    if (timestamps_error != nil){
        return timestamps_error
    }

    // Initialise local variables
    global_max_Tr1 := float64(0)
    global_sum_Tr1 := float64(0)
    global_sum_Web := float64(0)
    global_min_Web := PROCESSED_LINES[begin_index].Min_Web
    global_max_Tr3 := float64(0)
    global_sum_Tr3 := float64(0)
    global_sum_Width := int64(0)

    // Iterate on all the KeyValues of the lines
    for index:=begin_index; index<=end_index; index++{
        line_values := PROCESSED_LINES[index]
        global_max_Tr1 = math.Max(global_max_Tr1, line_values.Max_Tr1)
        global_sum_Tr1 += line_values.Mean_Tr1
        global_sum_Web += line_values.Mean_Web
        global_min_Web = math.Min(global_min_Web, line_values.Min_Web)
        global_max_Tr3 = math.Max(global_max_Tr3, line_values.Max_Tr3)
        global_sum_Tr3 += line_values.Mean_Tr3
        global_sum_Width += int64(line_values.Width)
    }
    number_of_line := float64(end_index - begin_index + 1)
    global_mean_Web := global_sum_Web/number_of_line

    sum_squared_diff_web := float64(0)
    for index:=begin_index; index<=end_index; index++{
        sum_squared_diff_web += math.Pow(PROCESSED_LINES[index].Mean_Web - global_mean_Web, 2)
    }
    global_std_Web := math.Sqrt(sum_squared_diff_web/number_of_line)
    global_mean_Tr1 := global_sum_Tr1/number_of_line
    global_mean_Tr3 := global_sum_Tr3/number_of_line
    global_mean_Width := global_sum_Width/int64(number_of_line)

    fmt.Println("Max Tr1 : ", global_max_Tr1)
    fmt.Println("Mean Tr1 : ", global_mean_Tr1)
    fmt.Println("Min Web : ", global_min_Web)
    fmt.Println("Mean Web : ", global_mean_Web)
    fmt.Println("Max Tr3 : ", global_max_Tr3)
    fmt.Println("Mean Tr3 : ", global_mean_Tr3)
    fmt.Println("Std : ", global_std_Web)
    fmt.Println("Pix : ", global_mean_Width)
    return nil
}

// function to test a txt file
func test(file_path string){
	file, err := os.Open(file_path)
    if err != nil {
        fmt.Println("Erreur lors de l'ouverture du fichier:", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    scanner.Scan()
    lineCount := 0

    // Read and process each line
    for scanner.Scan() {
        lineCount++
        line := scanner.Text()
        process_line(line)
    }

    // Scanner error
    if err := scanner.Err(); err != nil {
        fmt.Println("Erreur lors de la lecture du fichier:", err)
    }

    // Save the image in .png
    fichierImage, err := os.Create("image.png")
    if err != nil {
        panic(err)
    }
    defer fichierImage.Close()
    err = png.Encode(fichierImage, FINAL_IMAGE)
    if err != nil {
        panic(err)
    }
}

// Threshold the values then crop 
func threshold_crop_line(measurement_string_array []string)([]float64, error){
    LINE_COUNT ++ // For image
    fmt.Println("Size : ", len(measurement_string_array))

    thresholded_temperatures_array := make([]float64, len(measurement_string_array))
    if len(measurement_string_array)<1{ // Case : empty string array
        return thresholded_temperatures_array, nil
    }
    gradient_array := make([]float64, len(measurement_string_array))
    gradient_array[0] = 0
    max_grad := float64(0)
    // Convert the string array to a float64 array with "." instead of "," and threshold it
    for index, temperature_string := range measurement_string_array {
        temperature_string = strings.Replace(temperature_string, ",", ".", -1)
        temperature_float, parse_error := strconv.ParseFloat(temperature_string, 64)
        if (parse_error != nil){ // Case : Parsing error
            return thresholded_temperatures_array, parse_error
        }
        thresholded_float := math.Max(TEMPERATURE_THRESHOLD, temperature_float)
        // Draw on image
        FINAL_IMAGE.Set(index, LINE_COUNT, color.RGBA{255, 255-uint8((thresholded_float-TEMPERATURE_THRESHOLD)*255/float64(SEUIL_MAX_IMAGE-int(TEMPERATURE_THRESHOLD))), 255-uint8((thresholded_float-TEMPERATURE_THRESHOLD)*255/float64(SEUIL_MAX_IMAGE-int(TEMPERATURE_THRESHOLD))), 255})
        thresholded_temperatures_array[index] = thresholded_float
        // Prepare the gradient array by the same time and find the max gradient
        if index > 0{
            gradient_float := math.Abs(thresholded_float - thresholded_temperatures_array[index-1])
            gradient_array[index] = gradient_float
            max_grad = math.Max(max_grad, gradient_float)
        }
    }
    gradient_limit := max_grad/GRADIENT_LIMIT_FACTOR
    lower_index_crop := int(0)
    higher_index_crop := int(0)
    // Find the lower and the higher index for which the gradient is above the gradient_limit
    for index, gradient_float := range gradient_array {
        if gradient_float > gradient_limit{
            if lower_index_crop == 0{
                lower_index_crop = index
            }
            higher_index_crop = index
        }
    }

    // For image
    FINAL_IMAGE.Set(lower_index_crop, LINE_COUNT, color.RGBA{0, 0, 0, 255})
    local_lower_index = lower_index_crop
    FINAL_IMAGE.Set(higher_index_crop, LINE_COUNT, color.RGBA{0, 0, 0, 255})
    return thresholded_temperatures_array[lower_index_crop:higher_index_crop+1], nil
}

func write_values (timestamp time.Time, filtered_temperature_array []float64)error{
    // Initialise local vaariables
    width := len(filtered_temperature_array)
    half_index := (width+1)/2-1 // round up the index
    FINAL_IMAGE.Set(local_lower_index+half_index, LINE_COUNT, color.RGBA{0, 255, 255, 255})
    max_index_tr1 := int32(0)
    max_index_tr3 := int32(0)
    sum_tr1 := float64(0)
    max_tr1 := float64(0)
    sum_tr3 := float64(0)
    max_tr3 := float64(0)
    sum_web := float64(0)

    if width < int(WIDTH_LIMIT){ //Case : empty array
        return nil
    }

    //Max, Sum of Tr1
    for index:=0; index <= half_index; index++{
        temperature_float := filtered_temperature_array[index]
        sum_tr1 += temperature_float
        if (temperature_float > max_tr1){
            max_tr1 = temperature_float
            max_index_tr1 = int32(index)
        }
    }

    //Max, Sum of Tr3
    for index:=half_index+1; index < len(filtered_temperature_array); index++{
        temperature_float := filtered_temperature_array[index]
        sum_tr3 += temperature_float
        if (temperature_float >= max_tr3){
            max_tr3 = temperature_float
            max_index_tr3 = int32(index)
        }
    }

    min_web := filtered_temperature_array[max_index_tr1]
    //Mean, Min Web
    for index:=max_index_tr1; index<=max_index_tr3; index++{
        temperature_float := filtered_temperature_array[index]
        sum_web += temperature_float
        
        if (temperature_float < min_web){
            min_web = temperature_float
        }
    }
    
    // Write in the list of KeyValues
    mean_tr1 := sum_tr1/float64(half_index+1)
    mean_web := sum_web/float64(width)
    mean_tr3 := sum_tr3/float64(len(filtered_temperature_array) - half_index)
    PROCESSED_LINES = append(PROCESSED_LINES, KeyValues{
        Timestamp: timestamp,
        Max_Tr1: max_tr1,
        Mean_Tr1: mean_tr1,
        Mean_Web: mean_web,
        Min_Web: min_web,
        Max_Tr3: max_tr3,
        Mean_Tr3: mean_tr3,
        Width: int32(width),
    })

    // For final image 
    FINAL_IMAGE.Set(int(max_index_tr1)+local_lower_index, LINE_COUNT, color.RGBA{0, 255, 0, 255})
    FINAL_IMAGE.Set(int(max_index_tr3)+local_lower_index, LINE_COUNT, color.RGBA{0, 255, 0, 255})
    
    return nil
}

// Function to automatically process a string line
func process_line(line_string string)error{
    splited_line := strings.Split(line_string, "\t")
    // Parse the timestamp
    timestamp, parsing_error := time.Parse(TIME_FORMAT, splited_line[0])
    if (parsing_error != nil){
        return parsing_error
    }
    // Remove the columns we don't use
    measures := append(splited_line[1+NUMBER_FIRST_MEASURES_REMOVED:500], splited_line[510:len(splited_line)-4]...)
    // Do the thresholding and the crop of the measures
    thresholded_temperatures, threshold_crop_error := threshold_crop_line(measures)
    if (threshold_crop_error != nil){
        return threshold_crop_error
    }
    // Write the KeyValues of the line
    writing_error := write_values(timestamp, thresholded_temperatures)
    if (writing_error != nil){
        return writing_error
    }
    return nil
}

func main() {
    reset_dataframe()
    // Test on different txt files
    test("DUO01-02_0891.txt")
    test("DUO01-02_0892.txt")
    test("DUO01-02_0894.txt")
    test("DUO01-02_0895.txt")
    test("DUO01-02_0896.txt")
    test("DUO01-02_0897.txt")
    test("DUO01-02_0898.txt")
    test("DUO01-02_0899.txt")
    test("DUO01-02_0900.txt")
    // Test the result on asking the values
    compute_values("2024-02-13 11:05:06", "2024-02-13 11:05:14")
    
}

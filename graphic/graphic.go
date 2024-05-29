/*
 * File:    graphic.go.go
 * Date:    May 15, 2024
 * Author:  T.V
 * Email:   theo.verbrugge77@gmail.com
 * Project: goPostPro
 * Description:
 *   Contains some functions to display an image of the processing
 *
 */

package graphic

import (
	"goPostPro/global"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/mazznoer/colorgrad"
)

// Variables used to write correctly in the global image
var recording_image *image.RGBA
var image_line int = 0
var first_timestamp time.Time = time.Now()
var offset int = 0
var beam_id string = ""

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{0, 0, 0, 255}
	point := fixed.Point26_6{fixed.I(x), fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

// Used to convert a temperature into a thermal color
func thermalColor(temperature float64) color.Color {
	domain_value := (temperature - float64(global.Graphics.ThermalScaleStart)) / (float64(global.Graphics.ThermalScaleEnd) - float64(global.Graphics.ThermalScaleStart))
	return colorgrad.Inferno().At(domain_value)
}

func WriteCenteredText(text string) error {
	addLabel(recording_image, 900, image_line, text)
	return nil
}

// Save the global image with the timestamps of beginning and end of measurement
func saveImage() error {
	//Create the file
	recording_image.Rect = image.Rectangle{image.Point{0, 0}, image.Point{recording_image.Rect.Dx(), image_line}}
	filename := ""
	if beam_id == "" {
		filename = "results/000000[ "
	} else {
		filename = "results/" + beam_id + "[ "
	}
	filename = filename + first_timestamp.Format("15-04-05") + "_" + time.Now().Format("15-04-05") + "].png"

	imageFile, creation_error := os.Create(filename)
	if creation_error != nil {
		return creation_error
	}
	defer imageFile.Close()
	//Write our image in the file
	encoding_error := png.Encode(imageFile, recording_image)
	if encoding_error != nil {
		return encoding_error
	}
	log.Println("[GRAPHIC] Saved image as : ", filename)
	return nil
}

// Create a new image by reseting the variables used
func NewImage() error {
	recording_image = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{global.Graphics.ImageWidth, global.Graphics.ImageHeight}})
	image_line = 0
	first_timestamp = time.Now()
	beam_id = ""
	return nil
}

func ChangeName(beam_id_string string) error {
	beam_id = beam_id_string
	return nil
}

func ChangeImage() error {
	saving_error := saveImage()
	if saving_error != nil {
		return saving_error
	}
	creation_error := NewImage()
	if creation_error != nil {
		return creation_error
	}
	return nil
}

// Go to the next line of the image or create a new image if we reached the bottom of the picture
func NewLine() error {
	image_line++
	if image_line == global.Graphics.ImageHeight {
		ChangeImage()
	}
	image_line = image_line % global.Graphics.ImageHeight
	return nil
}

// Draw the left part which is the original thermal image
func DrawBeforeProcessing(temperature_array []float64) error {
	for index := 0; index < len(temperature_array); index++ {
		recording_image.Set(index, image_line, thermalColor(temperature_array[index]))
	}
	return nil
}

// Draw the right part which is the image after thresholding
func DrawAfterProcessing(processed_temperature_array []float64) error {
	for index := 0; index < len(processed_temperature_array); index++ {
		recording_image.Set(global.Graphics.ImageWidth/2+index, image_line, thermalColor(processed_temperature_array[index]))
	}
	return nil
}

// Draw the borders of the detected product
func DrawBorders(left_index int, right_index int) error {
	offset = global.Graphics.ImageWidth/2 + left_index
	recording_image.Set(offset, image_line, color.RGBA{0, 255, 0, 255})
	recording_image.Set(global.Graphics.ImageWidth/2+right_index, image_line, color.RGBA{0, 255, 0, 255})
	return nil
}

// Draw the limits used (max of each side) for the web
func DrawRegions(max_tr1 int, max_tr3 int) error {
	recording_image.Set(offset+max_tr1, image_line, color.RGBA{0, 255, 255, 255})
	recording_image.Set(offset+max_tr3, image_line, color.RGBA{0, 255, 255, 255})
	return nil
}

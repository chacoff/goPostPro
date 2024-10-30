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
	"flag"
	"goPostPro/global"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/golang/freetype"
	"golang.org/x/image/font"

	"github.com/mazznoer/colorgrad"
)

// Variables used to write correctly in the global image
var (
	recording_image                   *image.RGBA
	image_lines_timestamps_associated []string
	Hlines_index                      []int
	Hlines_colors                     []color.Color
	image_line                        int       = 0
	first_timestamp                   time.Time = time.Now()
	offset                            int       = 0
	beam_id                           string    = ""

	// Font variables
	dpi                        = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile                   = flag.String("fontfile", "Poppins-SemiBold.ttf", "filename of the ttf font")
	hinting                    = flag.String("hinting", "none", "none | full")
	size                       = flag.Float64("size", 48, "font size in points")
	fg, _                      = image.NewUniform(color.RGBA{255, 0, 0, 255}), image.White
	c        *freetype.Context = freetype.NewContext()
)

func GraphicInit() {
	NewImage()
	flag.Parse()
	fontBytes, err := os.ReadFile(*fontfile)
	if err != nil {
		log.Println(err)
		return
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	// Initialize the context.

	c.SetDPI(*dpi)
	c.SetFont(f)
	c.SetFontSize(*size)
	c.SetClip(recording_image.Bounds())
	c.SetDst(recording_image)
	c.SetSrc(fg)
	switch *hinting {
	default:
		c.SetHinting(font.HintingNone)
	case "full":
		c.SetHinting(font.HintingFull)
	}
}

func addLabel(x, y int, label string, color *image.Uniform, c *freetype.Context) {
	c.SetDst(recording_image)
	c.SetSrc(color)
	size := 12.0 // font size in pixels
	pt := freetype.Pt(x, y+int(c.PointToFixed(size)>>6))
	if _, err := c.DrawString(label, pt); err != nil {
		log.Println("[GRAPHIC] Error writing in the image")
	}
}

// Used to convert a temperature into a thermal color
func thermalColor(temperature float64) color.Color {
	domain_value := (temperature - float64(global.Graphics.ThermalScaleStart)) / (float64(global.Graphics.ThermalScaleEnd) - float64(global.Graphics.ThermalScaleStart))
	return colorgrad.Inferno().At(domain_value)
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func WriteCenteredText(text string, color color.Color, c *freetype.Context) error {
	addLabel(800, image_line, text, image.NewUniform(color), c)
	return nil
}

func DrawHLine(line int, color color.Color) {
	for horizontal_pixel := 0; horizontal_pixel < global.Graphics.ImageWidth; horizontal_pixel++ {
		recording_image.Set(horizontal_pixel, image_line, color)
	}
}

func DrawHLineAtTimestamp(timestamp_string string, color color.Color) {
	log.Println("[GRAPHIC]Cherche -> ", timestamp_string, "         Lignes de mesures de l'image : ", image_lines_timestamps_associated[0], " -> ", image_lines_timestamps_associated[len(image_lines_timestamps_associated)-1])
	timestamp, err := time.Parse(global.DBParams.TimeFormatRequest, timestamp_string)
	if err != nil {
		log.Println(err)
	}
	for index := 0; index < len(image_lines_timestamps_associated); index++ {
		index_time_object, err := time.Parse(global.PostProParams.TimeFormat, image_lines_timestamps_associated[index])
		if err != nil {
			log.Println(err)
		}
		if timestamp.Before(index_time_object) {
			// log.Println("[GRAPHIC]Trouve : ", index)
			Hlines_index = append(Hlines_index, index)
			Hlines_colors = append(Hlines_colors, color)
		}
	}
}

// saveImage saves the global image with the timestamps of beginning and end of measurement
func saveImage() error {
	var filename string
	var savingFolder string = global.Graphics.Savingfolder
	DrawAllHLines()

	//Create the file
	recording_image.Rect = image.Rectangle{image.Point{0, 0}, image.Point{recording_image.Rect.Dx(), image_line}}

	if beam_id == "" {
		filename = savingFolder + "/000000[ "
	} else {
		filename = savingFolder + "/" + beam_id + "["
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

// NewImage creates a new image by reseting the variables used
func NewImage() error {
	recording_image = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{global.Graphics.ImageWidth, global.Graphics.ImageHeight}})
	image_lines_timestamps_associated = make([]string, 0)
	Hlines_index = make([]int, 0)
	Hlines_colors = make([]color.Color, 0)
	image_line = 0
	first_timestamp = time.Now()
	beam_id = ""
	return nil
}

// ChangeName
func ChangeName(beam_id_string string) error {
	beam_id = beam_id_string
	return nil
}

// ChangeImage
func ChangeImage() error {
	saving_error := saveImage()
	if saving_error != nil {
		log.Println(saving_error)
		return saving_error
	}
	creation_error := NewImage()
	if creation_error != nil {
		log.Println(creation_error)
		return creation_error
	}
	return nil
}

// NewLine goes to the next line of the image or create a new image if we reached the bottom of the picture
func NewLine(timestamp time.Time) error {
	image_line++
	image_lines_timestamps_associated = append(image_lines_timestamps_associated, timestamp.Format(global.PostProParams.TimeFormat))
	if image_line == global.Graphics.ImageHeight {
		ChangeImage()
	}
	image_line = image_line % global.Graphics.ImageHeight
	return nil
}

// DrawBeforeProcessing draws the left part which is the original thermal image
func DrawBeforeProcessing(temperature_array []float64) error {
	for index := 0; index < len(temperature_array); index++ {
		recording_image.Set(index, image_line, thermalColor(temperature_array[index]))
	}
	return nil
}

// DrawAfterProcessing draws the right part which is the image after thresholding
func DrawAfterProcessing(processed_temperature_array []float64) error {
	for index := 0; index < len(processed_temperature_array); index++ {
		recording_image.Set(global.Graphics.ImageWidth/2+index, image_line, thermalColor(processed_temperature_array[index]))
	}
	return nil
}

// DrawBorders draws the borders of the detected product
func DrawBorders(left_index int, right_index int) error {
	offset = global.Graphics.ImageWidth/2 + left_index
	recording_image.Set(offset, image_line, color.RGBA{0, 255, 0, 255})
	recording_image.Set(global.Graphics.ImageWidth/2+right_index, image_line, color.RGBA{0, 255, 0, 255})
	return nil
}

// DrawRegions draws the limits used (max of each side) for the web
func DrawRegions(max_tr1 int, max_tr3 int) error {
	recording_image.Set(offset+max_tr1, image_line, color.RGBA{0, 255, 255, 255})
	recording_image.Set(offset+max_tr3, image_line, color.RGBA{0, 255, 255, 255})
	return nil
}

func DrawAllHLines() {
	for line := 0; line < len(Hlines_index); line++ {
		DrawHLine(Hlines_index[line], Hlines_colors[line])
	}
}

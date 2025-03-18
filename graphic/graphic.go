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
	"goPostPro/tcpServer"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"

	"github.com/mazznoer/colorgrad"
)

// Variables used to write correctly in the global image
var (
	result_image                      *image.RGBA = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})
	over_result_image                 *image.RGBA = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})
	report_image                      *image.RGBA = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})
	base_image                        *image.RGBA = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})
	gradient_image                    *image.RGBA = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})
	final_image                       *image.RGBA = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})
	image_lines_timestamps_associated []string
	image_line                        int        = 0
	first_timestamp                   time.Time  = time.Now().Add(-global.ReRunSynchDifference)
	beam_id                           string     = ""
	offset                                       = 0
	informations_displayed                       = 0
	pass_color                        color.RGBA = color.RGBA{255, 0, 0, 255}

	// Font variables
	dpi        = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile   = flag.String("fontfile", "Poppins-SemiBold.ttf", "filename of the ttf font")
	hinting    = flag.String("hinting", "none", "none | full")
	size       = flag.Float64("size", 24, "font size in points")
	title_size = flag.Float64("title_size", 72, "font size in points")
	fg, _      = image.NewUniform(color.RGBA{255, 0, 0, 255}), image.White

	result_image_context      = createNewContext(result_image)
	over_result_image_context = createNewContext(over_result_image)
	report_image_context      = createNewContext(report_image)
	base_image_context        = createNewContext(base_image)
	gradient_image_context    = createNewContext(gradient_image)
	final_image_context       = createNewContext(final_image)

	font_file *truetype.Font
)

func createNewContext(img *image.RGBA) *freetype.Context {
	c := freetype.NewContext()

	c.SetDPI(*dpi)
	fontBytes, err := os.ReadFile(*fontfile)
	if err != nil {
		log.Println(err)
		return c
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return c
	}
	c.SetFontSize(*size)
	c.SetFont(f)
	c.SetDst(img)
	c.SetClip(img.Rect.Bounds())
	c.SetSrc(fg)
	c.SetHinting(font.HintingNone)

	return c
}

func GraphicInit() {
	flag.Parse()
	// Initialize the context.
	final_image_context.SetFontSize(*title_size)
	NewImage()
}

func addLabel(c *freetype.Context, x, y int, label string, colori color.RGBA) {
	c.SetSrc(image.NewUniform(colori))
	size := 6.0 // font size in pixels
	pt := freetype.Pt(x, y+int(c.PointToFixed(size)>>6))
	if _, err := c.DrawString(label, pt); err != nil {
		log.Println("[GRAPHIC] Error writing in the image : ", err)
	}
}

func imagesTitles() {
	offset_title := 100
	addLabel(final_image_context, offset_title, 50, "Processing", color.RGBA{255, 255, 255, 255})
	addLabel(final_image_context, global.Graphics.ImageWidth+offset_title, 50, "Values", color.RGBA{255, 255, 255, 255})
	addLabel(final_image_context, global.Graphics.ImageWidth*2+offset_title, 50, "Recording", color.RGBA{255, 255, 255, 255})
	addLabel(final_image_context, global.Graphics.ImageWidth*3+offset_title, 50, "Gradient", color.RGBA{255, 255, 255, 255})
}

func SetPassColor(pass int) {
	colori := color.RGBA{0, 0, 0, 255}
	if pass == 1 {
		colori = color.RGBA{255, 0, 0, 255}
	}
	if pass == 2 {
		colori = color.RGBA{0, 255, 0, 255}
	}
	if pass == 3 {
		colori = color.RGBA{0, 0, 255, 255}
	}
	pass_color = colori
}

func AddInformation(text string) {
	report_image_context.SetSrc(image.NewUniform(pass_color))
	size := 6.0 // font size in pixels
	pt := freetype.Pt(20, 10+informations_displayed*20+int(report_image_context.PointToFixed(size)>>6))
	if _, err := report_image_context.DrawString(text, pt); err != nil {
		log.Println("[GRAPHIC] Error writing in the image")
	}
	informations_displayed++
}

// Used to convert a temperature into a thermal color
func thermalColor(temperature float64) color.Color {
	domain_value := (temperature - float64(global.Graphics.ThermalScaleStart)) / (float64(global.Graphics.ThermalScaleEnd) - float64(global.Graphics.ThermalScaleStart))
	return colorgrad.Inferno().At(domain_value)
}

func thermalColorGradient(temperature float64) color.Color {
	dividing_scale_value := 5
	domain_value := (temperature - float64(0)) / (float64(global.Graphics.ThermalScaleEnd/dividing_scale_value) - float64(0))
	return colorgrad.Inferno().At(domain_value)
}

func DrawHLine(line int, colori color.Color) {
	for horizontal_pixel := 0; horizontal_pixel < global.Graphics.ImageWidth; horizontal_pixel++ {
		over_result_image.Set(horizontal_pixel, line-1, colori)
		over_result_image.Set(horizontal_pixel, line, colori)
		over_result_image.Set(horizontal_pixel, line+1, colori)
	}
}

func DrawHLineAtTimestamp(timestamp_string string, label string, label_offset int) {
	timestamp, err := time.Parse(global.PostProParams.TimeFormat, timestamp_string)
	if err != nil {
		log.Println(err)
	}
	if len(image_lines_timestamps_associated) > 0 {
		log.Println("[GRAPHIC]Cherche -> ", timestamp.Format(global.PostProParams.TimeFormat), "         Lignes de mesures de l'image : ", image_lines_timestamps_associated[0], " -> ", image_lines_timestamps_associated[len(image_lines_timestamps_associated)-1])
	}

	for index := 0; index < len(image_lines_timestamps_associated); index++ {
		index_time_object, err := time.Parse(global.PostProParams.TimeFormat, image_lines_timestamps_associated[index])
		if err != nil {
			log.Println(err)
		}
		if timestamp.Before(index_time_object) {
			DrawHLine(index, pass_color)
			addLabel(over_result_image_context, 100, index+offset, label, pass_color)
			return
		}
	}
}

// saveImage saves the global image with the timestamps of beginning and end of measurement
func saveImage() error {
	var filename string
	var savingFolder string = global.Graphics.Savingfolder
	savingFolder += time.Now().Format("2006/01/02")

	// Check if folder exists else create it
	if _, err := os.Stat(savingFolder); os.IsNotExist(err) {
		err := os.MkdirAll(savingFolder, 0755)
		if err != nil {
			log.Println("[GRAPHIC RECORD] : Error ", err)
			return err
		}
		log.Println("[GRAPHIC RECORD] Created folder", savingFolder)
	}

	//Create the file
	final_image.Rect = image.Rectangle{image.Point{0, 0}, image.Point{final_image.Rect.Dx(), max(800, image_line)}}
	draw.Draw(final_image, result_image.Bounds().Add(image.Pt(0, 50)), result_image, image.Point{0, 0}, draw.Over)
	draw.Draw(final_image, over_result_image.Bounds().Add(image.Pt(0, 50)), over_result_image, image.Point{0, 0}, draw.Over)
	draw.Draw(final_image, report_image.Bounds().Add(image.Pt(global.Graphics.ImageWidth, 50)), report_image, image.Point{0, 0}, draw.Over)
	draw.Draw(final_image, base_image.Bounds().Add(image.Pt(global.Graphics.ImageWidth*2, 50)), base_image, image.Point{0, 0}, draw.Over)
	draw.Draw(final_image, gradient_image.Bounds().Add(image.Pt(global.Graphics.ImageWidth*3, 50)), gradient_image, image.Point{0, 0}, draw.Over)
	imagesTitles()

	if beam_id == "" {
		filename = savingFolder + "/000000[ "
	} else {
		filename = savingFolder + "/" + beam_id + "["
	}

	filename = filename + first_timestamp.Format("15-04-05") + "_" + time.Now().Add(-global.ReRunSynchDifference).Format("15-04-05") + "].png"

	imageFile, creation_error := os.Create(filename)
	if creation_error != nil {
		return creation_error
	}
	defer imageFile.Close()
	//Write our image in the file
	encoding_error := png.Encode(imageFile, final_image)
	if encoding_error != nil {
		return encoding_error
	}
	log.Println("[GRAPHIC] Saved image as : ", filename)
	return nil
}

// NewImage creates a new image by reseting the variables used
func NewImage() error {
	result_image = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{global.Graphics.ImageWidth, global.Graphics.ImageHeight}})
	result_image_context = createNewContext(result_image)
	over_result_image = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{global.Graphics.ImageWidth, global.Graphics.ImageHeight}})
	over_result_image_context = createNewContext(over_result_image)
	report_image = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{global.Graphics.ImageWidth, global.Graphics.ImageHeight}})
	report_image_context = createNewContext(report_image)
	base_image = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{global.Graphics.ImageWidth, global.Graphics.ImageHeight}})
	base_image_context = createNewContext(base_image)
	gradient_image = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{global.Graphics.ImageWidth, global.Graphics.ImageHeight}})
	gradient_image_context = createNewContext(gradient_image)
	final_image = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{global.Graphics.ImageWidth * 4, global.Graphics.ImageHeight + 20}})
	final_image_context = createNewContext(final_image)

	imagesTitles()

	image_lines_timestamps_associated = make([]string, 0)
	image_line = 0
	informations_displayed = 0
	first_timestamp = time.Now().Add(-global.ReRunSynchDifference)
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
	tcpServer.RecordAll()
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
		base_image.Set(index, image_line, thermalColor(temperature_array[index]))
	}
	return nil
}

// DrawAfterProcessing draws the right part which is the image after thresholding
func DrawAfterProcessing(processed_temperature_array []float64) error {
	for index := 0; index < len(processed_temperature_array); index++ {
		result_image.Set(index, image_line, thermalColor(processed_temperature_array[index]))
	}
	return nil
}

func DrawGradient(temperature_array []float64) error {
	for index := 0; index < len(temperature_array); index++ {
		gradient_image.Set(index, image_line, thermalColorGradient(temperature_array[index]))
	}
	return nil
}

// DrawBorders draws the borders of the detected product
func DrawBorders(left_index int, right_index int) error {
	over_result_image.Set(left_index, image_line, color.RGBA{0, 255, 0, 255})
	over_result_image.Set(right_index, image_line, color.RGBA{0, 255, 0, 255})
	offset = left_index
	return nil
}

// DrawRegions draws the limits used (max of each side) for the web
func DrawRegions(max_tr1 int, max_tr3 int) error {
	over_result_image.Set(offset+max_tr1, image_line, color.RGBA{0, 255, 255, 255})
	over_result_image.Set(offset+max_tr3, image_line, color.RGBA{0, 255, 255, 255})
	return nil
}

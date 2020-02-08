package palettecalculator

import (
	vision "cloud.google.com/go/vision/apiv1"
	"context"
	gax2 "github.com/googleapis/gax-go/v2"
	"gonum.org/v1/gonum/floats"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	"io"
	"math"
	"os"
)

const RED = 0
const GREEN = 1
const BLUE = 2

// Third party wrapper of the vision.NewImageAnnotatorClient method being used by DI
type Calculator interface {
	DetectImageProperties(ctx context.Context, img *pb.Image, ictx *pb.ImageContext, opts ...gax2.CallOption) (*pb.ImageProperties, error)
}

// Dependency wrapper for os.Open DI
type Opener interface {
	Open(name string) (*os.File, error)
}

type FileOpener struct{}

func (fo *FileOpener) Open(name string) (*os.File, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Third Party wrapper interface for vision.NewImageFromReader
type Reader interface {
	NewImageFromReader(r io.Reader) (*pb.Image, error)
}

type VisionReader struct{}

func (vr *VisionReader) NewImageFromReader(r io.Reader) (*pb.Image, error) {
	image, err := vision.NewImageFromReader(r)
	if err != nil {
		return nil, err
	}

	return image, nil

}

// Calculator for all palette combinations
type PaletteCalculator struct {
	Calculator
	Reader
	Opener
	context.Context
}

func NewPaletteCalculator(ctx context.Context) (*PaletteCalculator, error) {

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, err
	}

	return &PaletteCalculator{Calculator: client, Reader: new(VisionReader), Opener: new(FileOpener), Context: ctx}, nil

}

// Representation of RGB (red, green, blue) color
type RGB struct {
	red   float64 `json:"red"`
	green float64 `json:"green"`
	blue  float64 `json:"blue"`
}

// Representation of HSL (hue, saturation, luminosity) color
type HSL struct {
	hue        float64
	saturation float64
	luminosity float64
}

const RGBMax = float64(255)

// Calculates predominant color in image given file path to image
func (pc *PaletteCalculator) CalculatePredominantColor(file string) (*RGB, error) {
	dc := new(RGB)

	// Open file
	f, err := pc.Opener.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// generate image from file
	image, err := pc.Reader.NewImageFromReader(f)
	if err != nil {
		return nil, err
	}

	// calculate properties of generated image with
	properties, err := pc.Calculator.DetectImageProperties(pc.Context, image, nil)
	if err != nil {
		return nil, err
	}

	// iterate through resulting colors and add to dc's attributes
	for _, quantized := range properties.DominantColors.Colors {
		color := quantized.Color
		dc.red = math.Round(floats.Round(float64(color.Red*255), 1))
		dc.green = math.Round(floats.Round(float64(color.Green*255), 1))
		dc.blue = math.Round(floats.Round(float64(color.Blue*255), 1))
	}

	return dc, nil
}

// Calculates complimentary colors based on dominant color. Returns array of two RGB{}
func (pc *PaletteCalculator) CalculateComplimentaryColorScheme(dc *RGB) []RGB {

	complimentaryColors, hsl := pc.generateInitialRGBAndHSLForColor(dc)

	// Calculate complimentary color
	transformedHSL := &HSL{
		hue:        math.Abs(hsl.hue + 180 - 360),
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
	}

	// Convert complimentary HSL to RGB and append
	return append(complimentaryColors, *pc.ConvertHSLToRGB(transformedHSL))

}

// Calculates split complimentary colors based on dominant color. Returns array of three RGB{}
func (pc *PaletteCalculator) CalculateSplitComplimentaryColorScheme(dc *RGB) []RGB {

	splitComplimentaryColors, hsl := pc.generateInitialRGBAndHSLForColor(dc)

	// Calculate split complimentary colors
	transformedHSLCompliment1 := &HSL{
		hue:        math.Mod(hsl.hue+150, 360),
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
	}

	transformedHSLCompliment2 := &HSL{
		hue:        math.Mod(hsl.hue+210, 360),
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
	}

	// Convert split complimentary color HSL to RGB and append
	return append(splitComplimentaryColors, *pc.ConvertHSLToRGB(transformedHSLCompliment1), *pc.ConvertHSLToRGB(transformedHSLCompliment2))

}

// Calculates Triadic colors based on dominant color. Returns array of three RGB{}
func (pc *PaletteCalculator) CalculateTriadicColorScheme(dc *RGB) []RGB {

	triadicColors, hsl := pc.generateInitialRGBAndHSLForColor(dc)

	// Calculate triadic colors
	transformedTriadicColor1 := &HSL{
		hue:        math.Mod(hsl.hue+120, 360),
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
	}

	transformedTriadicColor2 := &HSL{
		hue:        math.Mod(hsl.hue+240, 360),
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
	}

	// Convert triadic HSL to RGB and append
	return append(triadicColors, *pc.ConvertHSLToRGB(transformedTriadicColor1), *pc.ConvertHSLToRGB(transformedTriadicColor2))

}

// Calculates Tetradic colors based on dominant color. Returns array of four RGB{}
func (pc *PaletteCalculator) CalculateTetradicColorScheme(dc *RGB) []RGB {

	tetradicColors, hsl := pc.generateInitialRGBAndHSLForColor(dc)

	// Calculate tetradic colors
	transformedTetradicColor1 := &HSL{
		hue:        math.Mod(hsl.hue+60, 360),
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
	}
	transformedTetradicColor2 := &HSL{
		hue:        math.Mod(hsl.hue+180, 360),
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
	}
	transformedTetradicColor3 := &HSL{
		hue:        math.Mod(hsl.hue+240, 360),
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
	}

	// Convert tertradic HSL to RGB and append
	return append(tetradicColors, *pc.ConvertHSLToRGB(transformedTetradicColor1), *pc.ConvertHSLToRGB(transformedTetradicColor2), *pc.ConvertHSLToRGB(transformedTetradicColor3))

}

func (pc *PaletteCalculator) generateInitialRGBAndHSLForColor(rgb *RGB) ([]RGB, *HSL) {
	var colors []RGB

	// Create RGB From dominant color
	dcToRGB := RGB{red: rgb.red, green: rgb.green, blue: rgb.blue}
	colors = append(colors, dcToRGB)

	// Convert to HSL
	hsl := pc.ConvertRGBToHSL(&dcToRGB)
	return colors, hsl
}

// Converting method for RGB to HSL
func (pc *PaletteCalculator) ConvertRGBToHSL(rgb *RGB) *HSL {
	rgbArr := []float64{rgb.red, rgb.green, rgb.blue}

	min := floats.Min(rgbArr) / RGBMax
	max := floats.Max(rgbArr) / RGBMax
	delta := max - min
	luminosity := floats.Round((max+min)/float64(2), 2)

	if delta > 0 {
		return pc.CalculateHSL(rgbArr, luminosity, delta)
	}
	return &HSL{hue: 0, saturation: 0, luminosity: luminosity}
}

// Converting method for HSL to RGB
func (pc *PaletteCalculator) ConvertHSLToRGB(hsl *HSL) *RGB {
	var temp1 float64
	var temp2 float64

	if hsl.hue > 0 && hsl.saturation > 0 {
		if hsl.luminosity < .5 {
			temp1 = hsl.luminosity * (1 + hsl.saturation)
		} else {
			temp1 = (hsl.luminosity + hsl.saturation) - (hsl.luminosity * hsl.saturation)
		}

		temp2 = 2*hsl.luminosity - temp1

		tempRed := floats.Round(hsl.hue/360+float64(1)/float64(3), 2)
		tempGreen := floats.Round(hsl.hue/360, 3)
		tempBlue := floats.Round(hsl.hue/360-float64(1)/float64(3), 2)
		return pc.calculateRGB([]float64{tempRed, tempGreen, tempBlue}, []float64{temp1, temp2})
	}
	return &RGB{
		red:   hsl.luminosity * 255,
		green: hsl.luminosity * 255,
		blue:  hsl.luminosity * 255,
	}

}

// RGB to HSL helper method
func (pc *PaletteCalculator) CalculateHSL(rgb []float64, luminosity float64, delta float64) *HSL {
	var saturation float64
	var hue float64
	min := floats.Round(floats.Min(rgb)/RGBMax, 3)
	max := floats.Round(floats.Max(rgb)/RGBMax, 3)
	red := floats.Round(rgb[RED]/RGBMax, 3)
	green := floats.Round(rgb[GREEN]/RGBMax, 3)
	blue := floats.Round(rgb[BLUE]/RGBMax, 3)

	if luminosity < .5 {
		saturation = floats.Round(delta/(max+min), 3)
	} else {
		saturation = floats.Round(delta/(2-max-min), 3)
	}

	if red == max {
		hue = (green - blue) / (max - min)
	}
	if green == max {
		hue = 2 + (blue-red)/(max-min)
	}
	if blue == max {
		hue = 4 + (red-green)/(max-min)
	}

	return &HSL{
		hue:        floats.Round(hue*60, 0),
		saturation: floats.Round(saturation, 2),
		luminosity: floats.Round(luminosity, 2),
	}

}

// HSL to RGB helper method
func (pc *PaletteCalculator) calculateRGB(tempRGB []float64, tempVar []float64) *RGB {
	for i, tempColor := range tempRGB {
		if tempColor < 0 {
			tempRGB[i] = tempColor + 1
		}
		if tempColor > 1 {
			tempRGB[i] = tempColor - 1
		}
	}

	red := floats.Round(pc.calculateRGBByColor(tempRGB[RED], tempVar)*255, 0)

	green := floats.Round(pc.calculateRGBByColor(tempRGB[GREEN], tempVar)*255, 0)

	blue := floats.Round(pc.calculateRGBByColor(tempRGB[BLUE], tempVar)*255, 0)

	return &RGB{red: red, green: green, blue: blue}

}

// HSL to RGB helper method
func (pc *PaletteCalculator) calculateRGBByColor(tempColor float64, tempVar []float64) float64 {
	if tempColor*6 < 1 {
		return floats.Round(tempVar[1]+(tempVar[0]-tempVar[1])*6*tempColor, 3)
	}
	if tempColor*2 < 1 {
		return floats.Round(tempVar[0], 3)
	}
	if tempColor*3 < 2 {
		return floats.Round(tempVar[1]+(tempVar[0]-tempVar[1])*(float64(2)/float64(3)-tempColor)*6, 3)
	}

	return floats.Round(tempVar[1], 3)
}

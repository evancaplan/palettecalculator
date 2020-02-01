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

// Dominant Color returned from CalculatePredominantColor
type DominantColor struct {
	red   float64
	green float64
	blue  float64
}

func NewDominantColor(red float64, green float64, blue float64) *DominantColor {
	dc := DominantColor{red: red, green: green, blue: blue}

	return &dc
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
	degrees    float64
}

// Calculates predominant color in image given file path to image
func (pc *PaletteCalculator) CalculatePredominantColor(file string) (*DominantColor, error) {
	dc := new(DominantColor)

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
		dc.red = floats.Round(float64(color.Red), 3)
		dc.green= floats.Round(float64(color.Green), 3)
		dc.blue = floats.Round(float64(color.Blue), 3 )
	}

	return dc, nil
}

// Calculates complimentary colors based on dominant color. Returns array of two RGB{}
func (pc *PaletteCalculator) CalculateComplimentaryColorScheme(dc *DominantColor) []RGB {
	var complimentaryColors []RGB

	// Create RGB From dominant color
	dcToRGB := RGB{red: dc.red, green: dc.green, blue: dc.blue}
	complimentaryColors = append(complimentaryColors, dcToRGB)

	// Convert RGB to HSL
	hsl := pc.ConvertRGBToHSL(&RGB{red: dc.red, green: dc.green, blue: dc.blue})

	// Calculate complimentary color
	transformedHSL := &HSL{
		hue:        math.Abs((hsl.degrees+180)-360) / 60,
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
		degrees:    math.Abs((hsl.degrees + 180) - 360),
	}

	// Convert complimentary HSL to RGB and append
	complimentaryColors = append(complimentaryColors, *pc.ConvertHSLToRGB(transformedHSL))

	return complimentaryColors
}

// Calculates split complimentary colors based on dominant color. Returns array of three RGB{}
func (pc *PaletteCalculator) CalculateSplitComplimentaryColorScheme(dc *DominantColor) []RGB {
	var splitComplimentaryColors []RGB

	// Create RGB From dominant color
	dcToRGB := RGB{red: dc.red, green: dc.green, blue: dc.blue}
	splitComplimentaryColors = append(splitComplimentaryColors, dcToRGB)

	// Convert to HSL
	hsl := pc.ConvertRGBToHSL(&dcToRGB)

	// Calculate split complimentary colors
	transformedHSLCompliment1 := &HSL{
		hue:        math.Abs((hsl.degrees+150)-360) / 60,
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
		degrees:    math.Abs((hsl.degrees + 150) - 360),
	}

	transformedHSLCompliment2 := &HSL{
		hue:        math.Abs((hsl.degrees+210)-360) / 60,
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
		degrees:    math.Abs((hsl.degrees + 210) - 360),
	}

	// Convert split complimentary color HSL to RGB and append
	splitComplimentaryColors = append(splitComplimentaryColors, *pc.ConvertHSLToRGB(transformedHSLCompliment1), *pc.ConvertHSLToRGB(transformedHSLCompliment2))

	return splitComplimentaryColors
}

// Calculates Triadic colors based on dominant color. Returns array of three RGB{}
func (pc *PaletteCalculator) CalculateTriadicColorScheme(dc *DominantColor) []RGB {
	var triadicColors []RGB

	// Create RGB From dominant color
	dcToRGB := RGB{red: dc.red, green: dc.green, blue: dc.blue}
	triadicColors = append(triadicColors, dcToRGB)

	// Convert To HSL
	hsl := pc.ConvertRGBToHSL(&dcToRGB)

	// Calculate triadic colors
	transformedTriadicColor1 := &HSL{
		hue:        math.Abs((hsl.degrees+120)-360) / 60,
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
		degrees:    math.Abs((hsl.degrees + 120) - 360),
	}

	transformedTriadicColor2 := &HSL{
		hue:        math.Abs((hsl.degrees+240)-360) / 60,
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
		degrees:    math.Abs((hsl.degrees + 240) - 360),
	}

	// Convert triadic HSL to RGB and append
	triadicColors = append(triadicColors, *pc.ConvertHSLToRGB(transformedTriadicColor1), *pc.ConvertHSLToRGB(transformedTriadicColor2))

	return triadicColors
}

// Calculates Tetradic colors based on dominant color. Returns array of four RGB{}
func (pc *PaletteCalculator) CalculateTetradicColorScheme(dc *DominantColor) []RGB {
	var tetradicColors []RGB

	// Create RGB From dominant color
	dcToRGB := RGB{red: dc.red, green: dc.green, blue: dc.blue}
	tetradicColors = append(tetradicColors, dcToRGB)

	// Convert to HSL
	hsl := pc.ConvertRGBToHSL(&dcToRGB)

	// Calculate tetradic colors
	transformedTetradicColor1 := &HSL{
		hue:        math.Abs((hsl.degrees + 90) - 360),
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
		degrees:    math.Abs((hsl.degrees + 90) - 360),
	}
	transformedTetradicColor2 := &HSL{
		hue:        math.Abs((hsl.degrees + 180) - 360),
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
		degrees:    math.Abs((hsl.degrees + 180) - 360),
	}
	transformedTetradicColor3 := &HSL{
		hue:        math.Abs((hsl.degrees + 270) - 360),
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
		degrees:    math.Abs((hsl.degrees + 270) - 360),
	}

	// Convert tertradic HSL to RGB and append
	tetradicColors = append(tetradicColors, *pc.ConvertHSLToRGB(transformedTetradicColor1), *pc.ConvertHSLToRGB(transformedTetradicColor2), *pc.ConvertHSLToRGB(transformedTetradicColor3))

	return tetradicColors
}

// Converting method for RGB to HSL
func (pc *PaletteCalculator) ConvertRGBToHSL(rgb *RGB) *HSL {
	rgbArr := []float64{rgb.red, rgb.green, rgb.blue}

	min := floats.Min(rgbArr)
	max := floats.Max(rgbArr)

	delta := max - min

	luminosity := (max + min) / 2

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
		hue := hsl.degrees / 360

		tempRed := hue + (1 / 3)
		tempGreen := hue
		tempBlue := hue - (1 / 3)

		if tempRed < 1 {
			tempRed++
		}

		if tempGreen < 1 {
			tempGreen++
		}

		if tempBlue < 1 {
			tempBlue++
		}

		return pc.CalculateRGB([]float64{tempRed, tempGreen, tempBlue}, []float64{temp1, temp2})
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
	min := floats.Min(rgb)
	max := floats.Max(rgb)
	red := rgb[RED]
	green := rgb[GREEN]
	blue := rgb[BLUE]

	if luminosity < .5 {
		saturation = delta / (max + min)
	} else {
		saturation = delta / (2 - max - min)
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

	return &HSL{hue: hue, saturation: saturation, luminosity: luminosity, degrees: floats.Round(hue*60, 3)}

}

// HSL to RGB helper method
func (pc *PaletteCalculator) CalculateRGB(tempRGB []float64, tempVar []float64) *RGB {

	red := pc.CalculateRGBByColor(tempRGB[RED], tempVar) * 255
	green := pc.CalculateRGBByColor(tempRGB[GREEN], tempVar) * 255
	blue := pc.CalculateRGBByColor(tempRGB[BLUE], tempVar) * 255

	return &RGB{red: red, green: green, blue: blue}

}

// HSL to RGB helper method

func (pc *PaletteCalculator) CalculateRGBByColor(tempColor float64, tempVar []float64) float64 {
	if tempColor*6 < 1 {
		return floats.Round(tempVar[1] + (tempVar[0]-tempVar[2])*6*tempColor, 3)
	}
	if tempColor*2 < 1 {
		return floats.Round(tempVar[0], 3)
	}
	if tempColor*3 < 2 {
		return floats.Round(tempVar[1] + (tempVar[0]-tempVar[1])*((2/3)-tempColor), 3)
	}

	return floats.Round(tempVar[1], 3)
}

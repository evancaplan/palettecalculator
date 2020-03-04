package palettecalculator

import (
	vision "cloud.google.com/go/vision/apiv1"
	"context"
	gax2 "github.com/googleapis/gax-go/v2"
	"gonum.org/v1/gonum/floats"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	col "google.golang.org/genproto/googleapis/type/color"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

const RED = 0
const GREEN = 1
const BLUE = 2
const RGBMax = float64(255)

// Representation of Color (red, green, blue) color
type Color struct {
	Red   float64 `json:"red"`
	Green float64 `json:"green"`
	Blue  float64 `json:"blue"`
	Hex   string  `json:"hex"`
}

// Representation of HSL (hue, saturation, luminosity) color
type HSL struct {
	hue        float64
	saturation float64
	luminosity float64
}

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

func NewPaletteCalculator() (*PaletteCalculator, error) {
	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, err
	}

	return &PaletteCalculator{Calculator: client, Reader: new(VisionReader), Opener: new(FileOpener), Context: ctx}, nil

}

// Calculates predominant color in image given file path to image
func (pc *PaletteCalculator) CalculatePredominantColor(file string) (*Color, error) {
	dc := new(Color)

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

	// iterate through resulting colors, get most dominant and add to dc's attributes
	var c *col.Color
	max := float32(0)
	for _, quantized := range properties.DominantColors.Colors {
		color := quantized.Color
		score := quantized.Score
		if score > max {
			max = score
			c = color
		}
	}

	dc.Red = float64(c.GetRed())
	dc.Green = float64(c.GetGreen())
	dc.Blue = float64(c.GetBlue())
	dc.Hex = pc.generateHex(dc.Red, dc.Green, dc.Blue)
	return dc, nil
}

// Calculates complimentary colors based on dominant color. Returns array of two Color{}
func (pc *PaletteCalculator) CalculateComplimentaryColorScheme(dc *Color) []Color {

	complimentaryColors, hsl := pc.generateInitialRGBAndHSLForColor(dc)

	// Calculate complimentary color
	transformedHSL := pc.transformHue(hsl, 180)

	// Convert complimentary HSL to Color and append
	return append(complimentaryColors, *pc.ConvertHSLToRGB(transformedHSL))

}

// Calculates split complimentary colors based on dominant color. Returns array of three Color{}
func (pc *PaletteCalculator) CalculateSplitComplimentaryColorScheme(dc *Color) []Color {

	splitComplimentaryColors, hsl := pc.generateInitialRGBAndHSLForColor(dc)

	// Calculate split complimentary colors
	transformedHSLCompliment1 := pc.transformHue(hsl, 150)

	transformedHSLCompliment2 := pc.transformHue(hsl, 210)

	// Convert split complimentary color HSL to Color and append
	return append(splitComplimentaryColors, *pc.ConvertHSLToRGB(transformedHSLCompliment1), *pc.ConvertHSLToRGB(transformedHSLCompliment2))

}

// Calculates Triadic colors based on dominant color. Returns array of three Color{}
func (pc *PaletteCalculator) CalculateTriadicColorScheme(dc *Color) []Color {

	triadicColors, hsl := pc.generateInitialRGBAndHSLForColor(dc)

	// Calculate triadic colors
	transformedTriadicColor1 := pc.transformHue(hsl, 120)

	transformedTriadicColor2 := pc.transformHue(hsl, 240)

	// Convert triadic HSL to Color and append
	return append(triadicColors, *pc.ConvertHSLToRGB(transformedTriadicColor1), *pc.ConvertHSLToRGB(transformedTriadicColor2))

}

// Calculates Tetradic colors based on dominant color. Returns array of four Color{}
func (pc *PaletteCalculator) CalculateTetradicColorScheme(dc *Color) []Color {

	tetradicColors, hsl := pc.generateInitialRGBAndHSLForColor(dc)

	// Calculate tetradic colors
	transformedTetradicColor1 := pc.transformHue(hsl, 60)

	transformedTetradicColor2 := pc.transformHue(hsl, 180)

	transformedTetradicColor3 := pc.transformHue(hsl, 240)

	// Convert tertradic HSL to Color and append
	return append(tetradicColors, *pc.ConvertHSLToRGB(transformedTetradicColor1), *pc.ConvertHSLToRGB(transformedTetradicColor2), *pc.ConvertHSLToRGB(transformedTetradicColor3))

}

func (pc *PaletteCalculator) generateInitialRGBAndHSLForColor(c *Color) ([]Color, *HSL) {
	var colors []Color

	// Create Color From dominant color
	dcToRGB := Color{Red: c.Red, Green: c.Green, Blue: c.Blue, Hex: c.Hex}
	colors = append(colors, dcToRGB)

	// Convert to HSL
	hsl := pc.ConvertRGBToHSL(&dcToRGB)
	return colors, hsl
}

func (pc *PaletteCalculator) transformHue(hsl *HSL, off float64) *HSL {
	return &HSL{
		hue:        math.Mod(hsl.hue+off, 360),
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
	}
}

func (pc *PaletteCalculator) generateHex(r float64, g float64, b float64) string {
	hex := []string{strconv.FormatInt(int64(r), 16), strconv.FormatInt(int64(g), 16), strconv.FormatInt(int64(b), 16)}

	return strings.Join(hex[:], "")
}

// Converting method for Color to HSL
func (pc *PaletteCalculator) ConvertRGBToHSL(rgb *Color) *HSL {
	rgbArr := []float64{rgb.Red, rgb.Green, rgb.Blue}

	min := floats.Min(rgbArr) / RGBMax
	max := floats.Max(rgbArr) / RGBMax
	delta := max - min
	luminosity := floats.Round((max+min)/float64(2), 2)

	if delta > 0 {
		return pc.CalculateHSL(rgbArr, luminosity, delta)
	}
	return &HSL{hue: 0, saturation: 0, luminosity: luminosity}
}

// Color to HSL helper method
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

// Converting method for HSL to Color
func (pc *PaletteCalculator) ConvertHSLToRGB(hsl *HSL) *Color {
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
	return &Color{
		Red:   hsl.luminosity * 255,
		Green: hsl.luminosity * 255,
		Blue:  hsl.luminosity * 255,
		Hex:   pc.generateHex(hsl.luminosity*255, hsl.luminosity*255, hsl.luminosity*255),
	}

}

// HSL to Color helper method
func (pc *PaletteCalculator) calculateRGB(tempRGB []float64, tempVar []float64) *Color {
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

	hex := pc.generateHex(red, green, blue)

	return &Color{Red: red, Green: green, Blue: blue, Hex: hex}

}

// HSL to Color helper method
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

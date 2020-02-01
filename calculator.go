package palettecalculator

import (
	vision "cloud.google.com/go/vision/apiv1"
	"context"
	"github.com/googleapis/gax-go/v2"
	"gonum.org/v1/gonum/floats"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	"io"
	"math"
	"os"
)

const RED = 0
const GREEN = 1
const BLUE = 2

type Calculator interface {
	DetectImageProperties(ctx context.Context, img *pb.Image, ictx *pb.ImageContext, opts ...gax.CallOption) (*pb.ImageProperties, error)
}

type Reader interface {
	NewImageFromReader(r io.Reader) (*pb.Image, error)
}

type PaletteCalculator struct {
	Calculator
	Reader
	context.Context
}

func NewPaletteCalculator(ctx context.Context) (*PaletteCalculator, error) {

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, err
	}

	return &PaletteCalculator{Calculator: client, Context: ctx}, nil

}

type DominantColor struct {
	red   float64
	green float64
	blue  float64
}

func NewDominantColor(red float64, green float64, blue float64) *DominantColor {
	dc := DominantColor{red: red, green: green, blue: blue}

	return &dc
}

type RGB struct {
	red   float64
	green float64
	blue  float64
}

type HSL struct {
	hue        float64
	saturation float64
	luminosity float64
	degrees    float64
}

func (pc *PaletteCalculator) CalculatePredominantColor(file string) (*DominantColor, error) {
	dc := new(DominantColor)

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	image, err := pc.Reader.NewImageFromReader(f)
	if err != nil {
		return nil, err
	}
	properties, err := pc.Calculator.DetectImageProperties(pc.Context, image, nil)
	if err != nil {
		return nil, err
	}

	for _, quantized := range properties.DominantColors.Colors {
		color := quantized.Color
		dc.red = float64(color.Red)
		dc.green = float64(color.Green)
		dc.blue = float64(color.Blue)
	}

	return dc, nil
}

func (pc *PaletteCalculator) CalculateComplimentaryColorScheme(dc *DominantColor) []RGB {
	var complimentaryColors []RGB
	dcToRGB := RGB{red: dc.red, green: dc.green, blue: dc.blue}
	complimentaryColors = append(complimentaryColors, dcToRGB)

	hsl := pc.ConvertRGBToHSL(&RGB{red: dc.red, green: dc.green, blue: dc.blue})
	transformedHSL := &HSL{
		hue:        math.Abs((hsl.degrees+180)-360) / 60,
		saturation: hsl.saturation,
		luminosity: hsl.luminosity,
		degrees:    math.Abs((hsl.degrees + 180) - 360),
	}

	complimentaryColors = append(complimentaryColors, *pc.ConvertHSLToRGB(transformedHSL))

	return complimentaryColors
}

func (pc *PaletteCalculator) CalculateSplitComplimentaryColorScheme(dc *DominantColor) []RGB {

	var splitComplimentaryColors []RGB
	dcToRGB := RGB{red: dc.red, green: dc.green, blue: dc.blue}
	splitComplimentaryColors = append(splitComplimentaryColors, dcToRGB)

	hsl := pc.ConvertRGBToHSL(&dcToRGB)
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

	splitComplimentaryColors = append(splitComplimentaryColors, *pc.ConvertHSLToRGB(transformedHSLCompliment1), *pc.ConvertHSLToRGB(transformedHSLCompliment2))

	return splitComplimentaryColors
}

func (pc *PaletteCalculator) CalculateTriadicColorScheme(dc *DominantColor) []RGB {
	var triadicColors []RGB
	dcToRGB := RGB{red: dc.red, green: dc.green, blue: dc.blue}
	triadicColors = append(triadicColors, dcToRGB)

	hsl := pc.ConvertRGBToHSL(&dcToRGB)
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

	triadicColors = append(triadicColors, *pc.ConvertHSLToRGB(transformedTriadicColor1), *pc.ConvertHSLToRGB(transformedTriadicColor2))

	return triadicColors
}

func (pc *PaletteCalculator) CalculateTetradicColorScheme(dc *DominantColor) []RGB {
	var tetradicColors []RGB
	dcToRGB := RGB{red: dc.red, green: dc.green, blue: dc.blue}
	tetradicColors = append(tetradicColors, dcToRGB)

	hsl := pc.ConvertRGBToHSL(&dcToRGB)
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

	tetradicColors = append(tetradicColors, *pc.ConvertHSLToRGB(transformedTetradicColor1), *pc.ConvertHSLToRGB(transformedTetradicColor2), *pc.ConvertHSLToRGB(transformedTetradicColor3))

	return tetradicColors
}

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

func (pc *PaletteCalculator) CalculateRGB(tempRGB []float64, tempVar []float64) *RGB {

	red := pc.CalculateRGBByColor(tempRGB[RED], tempVar) * 255
	green := pc.CalculateRGBByColor(tempRGB[GREEN], tempVar) * 255
	blue := pc.CalculateRGBByColor(tempRGB[BLUE], tempVar) * 255

	return &RGB{red: red, green: green, blue: blue}

}

func (pc *PaletteCalculator) CalculateRGBByColor(tempColor float64, tempVar []float64) float64 {
	if tempColor*6 < 1 {
		return tempVar[1] + (tempVar[0]-tempVar[2])*6*tempColor
	}
	if tempColor*2 < 1 {
		return tempVar[0]
	}
	if tempColor*3 < 2 {
		return tempVar[1] + (tempVar[0]-tempVar[1])*((2/3)-tempColor)
	}

	return tempVar[1]
}

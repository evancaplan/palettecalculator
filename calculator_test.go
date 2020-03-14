package palettecalculator

import (
	"context"
	"errors"
	"fmt"
	"github.com/googleapis/gax-go/v2"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	"google.golang.org/genproto/googleapis/type/color"
	"io"
	"os"
	"reflect"
	"testing"
)

const Red = 24
const Green = 98
const Blue = 119
const hue = 193
const saturation = .66
const luminosity = .28
const Hex = "186277"

func TestCalculatePredominantColorFromFile(t *testing.T) {
	for _, test := range []struct {
		name                  string
		file                  os.File
		filePath              string
		data                  []*pb.ColorInfo
		visionData            []byte
		expectedDominantColor *Color
		calculatorErr         error
		openerErr             error
		readerErr             error
		expectedErr           error
	}{
		{
			name:                  "should return dominant color with no error",
			file:                  *new(os.File),
			filePath:              "test/file.path",
			data:                  []*pb.ColorInfo{&pb.ColorInfo{Color: &color.Color{Red: Red, Green: Green, Blue: Blue}, Score: .01}},
			visionData:            []byte{},
			expectedDominantColor: &Color{Red: Red, Green: Green, Blue: Blue, Hex: Hex},
			calculatorErr:         nil,
			openerErr:             nil,
			readerErr:             nil,
			expectedErr:           nil,
		},
		{
			name:                  "error occurs when file is opened",
			file:                  *new(os.File),
			filePath:              "test/file.path",
			data:                  nil,
			visionData:            nil,
			expectedDominantColor: nil,
			calculatorErr:         nil,
			openerErr:             errors.New("os error has occurRed. file not found"),
			readerErr:             nil,
			expectedErr:           errors.New("os error has occurRed. file not found"),
		},
		{
			name:                  "error occurs when file is read as image",
			file:                  *new(os.File),
			filePath:              "test/file.path",
			data:                  nil,
			visionData:            nil,
			expectedDominantColor: nil,
			calculatorErr:         nil,
			openerErr:             nil,
			readerErr:             errors.New("unable to read from file"),
			expectedErr:           errors.New("unable to read from file"),
		}, {
			name:                  "error occurs when image properties are calculated",
			file:                  *new(os.File),
			filePath:              "test/file.path",
			data:                  nil,
			visionData:            nil,
			expectedDominantColor: nil,
			calculatorErr:         errors.New("unable to calculate image properties"),
			openerErr:             nil,
			readerErr:             nil,
			expectedErr:           errors.New("unable to calculate image properties"),
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			paletteCalculator := new(PaletteCalculator)

			paletteCalculator.Calculator = &MockCalculator{data: test.data, err: test.calculatorErr}
			paletteCalculator.Opener = &MockFileOpener{data: &test.file, err: test.openerErr}
			paletteCalculator.Reader = &MockVisionReader{data: test.visionData, err: test.readerErr}

			returnedDominantColor, err := paletteCalculator.CalculatePredominantColorFromFile(test.filePath)

			if !reflect.DeepEqual(test.expectedDominantColor, returnedDominantColor) {
				t.Errorf("expected: %+v\n returned: %+v\n ", test.expectedDominantColor, returnedDominantColor)
			}

			if !reflect.DeepEqual(test.expectedErr, err) {
				t.Errorf("expected error: %s returned error: %s", test.expectedErr.Error(), err.Error())
			}
		})
	}
}
func TestCalculatePredominantColorFromURI(t *testing.T) {
	for _, test := range []struct {
		name                  string
		uri                   string
		data                  []*pb.ColorInfo
		visionData            []byte
		expectedDominantColor *Color
		calculatorErr         error
		expectedErr           error
	}{
		{
			name:                  "should return dominant color with no error",
			uri:                   "test.uri",
			data:                  []*pb.ColorInfo{&pb.ColorInfo{Color: &color.Color{Red: Red, Green: Green, Blue: Blue}, Score: .01}},
			visionData:            []byte{},
			expectedDominantColor: &Color{Red: Red, Green: Green, Blue: Blue, Hex: Hex},
			calculatorErr:         nil,
			expectedErr:           nil,
		}, {
			name:                  "error occurs when image properties are calculated",
			uri:                   "test.uri",
			data:                  nil,
			visionData:            nil,
			expectedDominantColor: nil,
			calculatorErr:         errors.New("unable to calculate image properties"),
			expectedErr:           errors.New("unable to calculate image properties"),
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			paletteCalculator := new(PaletteCalculator)

			paletteCalculator.Calculator = &MockCalculator{data: test.data, err: test.calculatorErr}
			paletteCalculator.Reader = &MockVisionReader{data: test.visionData}

			returnedDominantColor, err := paletteCalculator.CalculatePredominantColorFromURI(test.uri)

			if !reflect.DeepEqual(test.expectedDominantColor, returnedDominantColor) {
				t.Errorf("expected: %+v\n returned: %+v\n ", test.expectedDominantColor, returnedDominantColor)
			}

			if !reflect.DeepEqual(test.expectedErr, err) {
				t.Errorf("expected error: %s returned error: %s", test.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestCalculateComplimentaryColorScheme(t *testing.T) {
	dominantColors := Color{Red: Red, Green: Green, Blue: Blue, Hex: Hex}
	expectedRGB := []Color{{Red: Red, Green: Green, Blue: Blue, Hex: Hex}, {Red: 119, Green: 45, Blue: 24, Hex: "772d18"}}
	paletteCalculator := new(PaletteCalculator)

	returnedRGB := paletteCalculator.CalculateComplimentaryColorScheme(&dominantColors)

	if !reflect.DeepEqual(expectedRGB, returnedRGB) {
		t.Errorf("expected: %v\n returned %v\n", expectedRGB, returnedRGB)
	}

}

func TestCalculateSplitComplimentaryColorScheme(t *testing.T) {
	dominantColors := &Color{Red: Red, Green: Green, Blue: Blue, Hex: Hex}
	expectedRGB := []Color{{Red: Red, Green: Green, Blue: Blue, Hex: Hex}, {119, 24, 51, "771833"}, {119, 92, 24, "775c18"}}
	paletteCalculator := new(PaletteCalculator)

	returnedRGB := paletteCalculator.CalculateSplitComplimentaryColorScheme(dominantColors)

	if !reflect.DeepEqual(expectedRGB, returnedRGB) {
		t.Errorf("expected: %v\n returned %v\n", expectedRGB, returnedRGB)
	}

}

func TestCalculateTriadicColorScheme(t *testing.T) {
	dominantColors := &Color{Red: Red, Green: Green, Blue: Blue, Hex: Hex}
	expectedRGB := []Color{{Red: Red, Green: Green, Blue: Blue, Hex: Hex}, {119, 24, 96, "771860"}, {96, 119, 24, "607718"}}
	paletteCalculator := new(PaletteCalculator)

	returnedRGB := paletteCalculator.CalculateTriadicColorScheme(dominantColors)

	if !reflect.DeepEqual(expectedRGB, returnedRGB) {
		t.Errorf("expected: %v\n returned %v\n", expectedRGB, returnedRGB)
	}

}

func TestCalculateTetradicColorScheme(t *testing.T) {
	dominantColors := &Color{Red: Red, Green: Green, Blue: Blue, Hex: Hex}
	expectedRGB := []Color{{Red: Red, Green: Green, Blue: Blue, Hex: Hex}, {47, 24, 119, "2f1877"}, {119, 45, 24, "772d18"}, {96, 119, 24, "607718"}}
	paletteCalculator := new(PaletteCalculator)

	returnedRGB := paletteCalculator.CalculateTetradicColorScheme(dominantColors)

	if !reflect.DeepEqual(expectedRGB, returnedRGB) {
		t.Errorf("expected: %v\n returned %v\n", expectedRGB, returnedRGB)
	}

}

func TestConvertRGBToHSL(t *testing.T) {
	testRGB := &Color{Red: Red, Green: Green, Blue: Blue}
	paletteCalculator := new(PaletteCalculator)
	expectedHSL := &HSL{hue: hue, saturation: saturation, luminosity: luminosity}

	returnedHSL := paletteCalculator.ConvertRGBToHSL(testRGB)

	if !reflect.DeepEqual(expectedHSL, returnedHSL) {
		t.Errorf("expected: %v\n returned: %v\n", expectedHSL, returnedHSL)
	}

}

func TestConvertHSLToRGB(t *testing.T) {
	testHSL := &HSL{hue: hue, saturation: saturation, luminosity: luminosity}
	paletteCalculator := new(PaletteCalculator)
	expectedRGB := &Color{Red: Red, Green: Green, Blue: Blue, Hex: Hex}

	returnedRGB := paletteCalculator.ConvertHSLToRGB(testHSL)
	if !reflect.DeepEqual(expectedRGB, returnedRGB) {
		t.Errorf("expected: %v\n returned: %v\n", expectedRGB, returnedRGB)
	}

}

type MockCalculator struct {
	data []*pb.ColorInfo
	err  error
}

func (m *MockCalculator) DetectImageProperties(ctx context.Context, img *pb.Image, ictx *pb.ImageContext, opts ...gax.CallOption) (*pb.ImageProperties, error) {
	return &pb.ImageProperties{DominantColors: &pb.DominantColorsAnnotation{Colors: m.data}}, m.err
}

type MockFileOpener struct {
	data *os.File
	err  error
}

func (m *MockFileOpener) Open(name string) (*os.File, error) {
	return m.data, m.err
}

type MockVisionReader struct {
	data []byte
	err  error
}

func (m *MockVisionReader) NewImageFromReader(r io.Reader) (*pb.Image, error) {
	return &pb.Image{Content: m.data}, m.err
}

func (m *MockVisionReader) NewImageFromURI(uri string) *pb.Image {
	return &pb.Image{Content: m.data}
}

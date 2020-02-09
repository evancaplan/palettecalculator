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

const red = 24
const green = 98
const blue = 119
const hue = 193
const saturation = .66
const luminosity = .28
const hex = "186277"

func TestCalculatePredominantColor(t *testing.T) {
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
			data:                  []*pb.ColorInfo{&pb.ColorInfo{Color: &color.Color{Red: .094, Green: .384, Blue: .466}}},
			visionData:            []byte{},
			expectedDominantColor: &Color{red: red, green: green, blue: blue, hex: hex},
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
			openerErr:             errors.New("os error has occurred. file not found"),
			readerErr:             nil,
			expectedErr:           errors.New("os error has occurred. file not found"),
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

			returnedDominantColor, err := paletteCalculator.CalculatePredominantColor(test.filePath)

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
	dominantColors := Color{red: red, green: green, blue: blue, hex: hex}
	expectedRGB := []Color{{red: red, green: green, blue: blue, hex: hex}, {red: 119, green: 45, blue: 24, hex: "772d18"}}
	paletteCalculator := new(PaletteCalculator)

	returnedRGB := paletteCalculator.CalculateComplimentaryColorScheme(&dominantColors)

	if !reflect.DeepEqual(expectedRGB, returnedRGB) {
		t.Errorf("expected: %v\n returned %v\n", expectedRGB, returnedRGB)
	}

}

func TestCalculateSplitComplimentaryColorScheme(t *testing.T) {
	dominantColors := &Color{red: red, green: green, blue: blue, hex: hex}
	expectedRGB := []Color{{red: red, green: green, blue: blue, hex: hex}, {119, 24, 51, "771833"}, {119, 92, 24, "775c18"}}
	paletteCalculator := new(PaletteCalculator)

	returnedRGB := paletteCalculator.CalculateSplitComplimentaryColorScheme(dominantColors)

	if !reflect.DeepEqual(expectedRGB, returnedRGB) {
		t.Errorf("expected: %v\n returned %v\n", expectedRGB, returnedRGB)
	}

}

func TestCalculateTriadicColorScheme(t *testing.T) {
	dominantColors := &Color{red: red, green: green, blue: blue, hex: hex}
	expectedRGB := []Color{{red: red, green: green, blue: blue, hex: hex}, {119, 24, 96, "771860"}, {96, 119, 24, "607718"}}
	paletteCalculator := new(PaletteCalculator)

	returnedRGB := paletteCalculator.CalculateTriadicColorScheme(dominantColors)

	if !reflect.DeepEqual(expectedRGB, returnedRGB) {
		t.Errorf("expected: %v\n returned %v\n", expectedRGB, returnedRGB)
	}

}

func TestCalculateTetradicColorScheme(t *testing.T) {
	dominantColors := &Color{red: red, green: green, blue: blue, hex: hex}
	expectedRGB := []Color{{red: red, green: green, blue: blue, hex: hex}, {47, 24, 119, "2f1877"}, {119, 45, 24, "772d18"}, {96, 119, 24, "607718"}}
	paletteCalculator := new(PaletteCalculator)

	returnedRGB := paletteCalculator.CalculateTetradicColorScheme(dominantColors)

	if !reflect.DeepEqual(expectedRGB, returnedRGB) {
		t.Errorf("expected: %v\n returned %v\n", expectedRGB, returnedRGB)
	}

}

func TestConvertRGBToHSL(t *testing.T) {
	testRGB := &Color{red: red, green: green, blue: blue}
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
	expectedRGB := &Color{red: red, green: green, blue: blue, hex: hex}

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

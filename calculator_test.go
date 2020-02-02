package palettecalculator

import (
	"context"
	"errors"
	"fmt"
	"github.com/googleapis/gax-go/v2"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	"google.golang.org/genproto/googleapis/type/color"
	"io"
	"math"
	"os"
	"reflect"
	"testing"
)

const red = 128
const green = 51
const blue = 77
const hue = .34
const saturation = .43
const luminosity = .35
const degrees = 340

func TestCalculatePredominantColor(t *testing.T) {
	for _, test := range []struct {
		name                  string
		file                  os.File
		filePath              string
		data                  []*pb.ColorInfo
		visionData            []byte
		expectedDominantColor *RGB
		calculatorErr         error
		openerErr             error
		readerErr             error
		expectedErr           error
	}{
		{
			name:                  "should return dominant color with no error",
			file:                  *new(os.File),
			filePath:              "test/file.path",
			data:                  []*pb.ColorInfo{&pb.ColorInfo{Color: &color.Color{Red: .50, Green: .20, Blue: .30}}},
			visionData:            []byte{},
			expectedDominantColor: &RGB{red: red, green: green, blue: blue},
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
	dominantColors := RGB{red: red, green: green, blue: blue}
	expectedRGB := []RGB{{red: red, green: green, blue: blue}, {red: 51, green: 128, blue: 100}}
	paletteCalculator := new(PaletteCalculator)

	returnedRGB := paletteCalculator.CalculateComplimentaryColorScheme(&dominantColors)

	if !reflect.DeepEqual(expectedRGB, returnedRGB) {
		t.Errorf("expected: %f\n returned %f\n", expectedRGB, returnedRGB)
	}

}

func TestCalculateSplitComplimentaryColorScheme(t *testing.T) {
	dominantColors := &RGB{red: red, green: green, blue: blue}
	expectedRGB := []RGB{{red: red, green: green, blue: blue}, {51, 128, 63}, {51, 10, 128}}
	paletteCalculator := new(PaletteCalculator)

	returnedRGB := paletteCalculator.CalculateSplitComplimentaryColorScheme(dominantColors)

	if !reflect.DeepEqual(expectedRGB, returnedRGB) {
		t.Errorf("expected: %f\n returned %f\n", expectedRGB, returnedRGB)
	}

}

func TestCalculateTriadicColorScheme(t *testing.T) {
	dominantColors := &RGB{red: red, green: green, blue: blue}
	expectedRGB := []RGB{{red: red, green: green, blue: blue}, {4, 128, 26}, {51, 4, 128}}
	paletteCalculator := new(PaletteCalculator)

	returnedRGB := paletteCalculator.CalculateTriadicColorScheme(dominantColors)

	if !reflect.DeepEqual(expectedRGB, returnedRGB) {
		t.Errorf("expected: %f\n returned %f\n", expectedRGB, returnedRGB)
	}

}

func TestCalculateTetradicColorScheme(t *testing.T) {
	dominantColors := &RGB{red: red, green: green, blue: blue}
	expectedRGB := []RGB{{red: red, green: green, blue: blue}, {11, 128, -15}, {51, 128, 100}, {51, 51, 128}}
	paletteCalculator := new(PaletteCalculator)

	returnedRGB := paletteCalculator.CalculateTetradicColorScheme(dominantColors)

	if !reflect.DeepEqual(expectedRGB, returnedRGB) {
		t.Errorf("expected: %f\n returned %f\n", expectedRGB, returnedRGB)
	}

}

func TestConvertRGBToHSL(t *testing.T) {
	testRGB := &RGB{red: red, green: green, blue: blue}
	paletteCalculator := new(PaletteCalculator)
	expectedHSL := &HSL{hue: .34, saturation: .43, luminosity: .35, degrees: 340}

	returnedHSL := paletteCalculator.ConvertRGBToHSL(testRGB)

	if !reflect.DeepEqual(expectedHSL, returnedHSL) {
		t.Errorf("expected: %v\n returned: %v\n", expectedHSL, returnedHSL)
	}

}

func TestConvertHSLToRGB(t *testing.T) {
	testHSL := &HSL{hue: hue, saturation: saturation, luminosity: luminosity, degrees: math.Abs(degrees - 360)}
	paletteCalculator := new(PaletteCalculator)
	expectedRGB := &RGB{red: red, green: green, blue: blue}

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

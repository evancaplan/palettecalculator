package palettecalculator

import (
	"context"
	"errors"
	"fmt"
	"github.com/googleapis/gax-go"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	"google.golang.org/genproto/googleapis/type/color"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestCalculatePredominantColor(t *testing.T) {
	for _, test := range []struct {
		name                  string
		file                  os.File
		filePath              string
		data                  []*pb.ColorInfo
		visionData            []byte
		expectedDominantColor *DominantColor
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
			expectedDominantColor: NewDominantColor(.50, .20, .30),
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
				name: "error occurs when image properties are calculated",
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

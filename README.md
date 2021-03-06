# Go Palette Calculator
### Using [Google Cloud's Vision API](https://cloud.google.com/vision) in order to analyze an image, this Go library creates color palettes available in 4 palette options based on the predominant color from that analysis:
#### Complimentary
##### Usage:
```
c, err := NewPaletteCalculator()
if err != nil {
    handle error
}

predominantColor, err := c.CalculatePredominantColor(filePath)
if err != nil {
    handle error
}

complimentaryColor := c.CalculateComplimentaryColorScheme(predominantColor)
```
 #### Split Complimentary 
##### Usage:
```
c, err := NewPaletteCalculator()
if err != nil {
    handle error
}

predominantColor, err := c.CalculatePredominantColor
if err != nil {
    handle error
}

complimentaryColor := c.CalculateSplitComplimentaryColorScheme(predominantColor)
```
#### Triadic 
##### Usage:
```
c, err := NewPaletteCalculator()
if err != nil {
    handle error
}

predominantColor, err := c.CalculatePredominantColor(filePath)
if err != nil {
    handle error
}

complimentaryColor := c.CalculateTriadicColorScheme(predominantColor)
```
#### Tetradic
##### Usage:
```
c, err := NewPaletteCalculator()
if err != nil {
    handle error
}

predominantColor, err := c.CalculatePredominantColor(filePath)
if err != nil {
    handle error
}

complimentaryColor := c.CalculateTetradicColorScheme(predominantColor)
```
### REST API use case:
##### https://github.com/evancaplan/palette-api/

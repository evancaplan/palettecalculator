# Go Palette Calculator
### Using Google Cloud's Vision Library, this Go library creates color palettes available in 4 palette options:
#### Complimentary
##### Usage:
```
c := NewPaletteCalculator()

predominantColor := c.CalculatePredominantColor(filePath)

complimentaryColor := c.CalculateComplimentaryColorScheme(predominantColor)
```
 #### Split Complimentary 
##### Usage:
```
c := NewPaletteCalculator()

predominantColor := c.CalculatePredominantColor(filePath)

complimentaryColor := c.CalculateSplitComplimentaryColorScheme(predominantColor)
```
#### Triadic 
##### Usage:
```
c := NewPaletteCalculator()

predominantColor := c.CalculatePredominantColor(filePath)

complimentaryColor := c.CalculateTriadicColorScheme(predominantColor)
```
#### Tetradic
##### Usage:
```
c := NewPaletteCalculator()

predominantColor := c.CalculatePredominantColor(filePath)

complimentaryColor := c.CalculateTetradicColorScheme(predominantColor)
```

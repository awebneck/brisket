package brisket

import (
  "image"
)

func ConvertToGrayscale(img image.Image) image.Image {
  bounds := img.Bounds()
  gray := image.NewGray(bounds)
  model := gray.ColorModel()
  for i := 0; i < bounds.Max.X; i++ {
    for j := 0; j < bounds.Max.Y; j++ {
      gray.Set(i, j, model.Convert(img.At(i,j)))
    }
  }
  return gray
}

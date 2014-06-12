package brisket

import (
  "image"
)

func ConvertToGrayscale(img image.Image) *image.Gray {
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

func ConvertToColor(img image.Image) *image.RGBA {
  bounds := img.Bounds()
  rgba := image.NewRGBA(bounds)
  model := rgba.ColorModel()
  for i := 0; i < bounds.Max.X; i++ {
    for j := 0; j < bounds.Max.Y; j++ {
      rgba.Set(i, j, model.Convert(img.At(i,j)))
    }
  }
  return rgba
}

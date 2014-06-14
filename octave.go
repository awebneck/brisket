package brisket

import (
  "image"
)

type octave struct {
  image image.Image
  scale float64
  fast *fast
}


func NewOctave(pyr *pyramid, scale float64, thresh uint8) (*octave, error) {
  oct := new(octave)
  // Need to scale image first
  oct.image = pyr.image
  oct.scale = scale
  oct.fast = NewFast(oct, thresh, PatternSize9_16)
  return oct, nil
}

func NewOctaveFromGray(gray *image.Gray, scale float64, thresh uint8) (*octave, error) {
  oct := new(octave)
  // Need to scale image first
  oct.image = gray
  oct.scale = scale
  oct.fast = NewFast(oct, thresh, PatternSize9_16)
  return oct, nil
}

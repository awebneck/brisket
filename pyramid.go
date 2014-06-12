package brisket

import (
  "image"
)

type pyramid struct {
  image image.Image
  layers []*octave
}

func NewPyramid(image image.Image, octaves int, thresh uint8) (*pyramid, error) {
  var err error
  pyr := new(pyramid)
  pyr.image = image
  pyr.layers = make([]*octave, octaves*2 + 2)
  pyr.layers[0], err = NewOctave(pyr, 1.0, thresh)
  pyr.layers[1], err = NewOctave(pyr, 1.5, thresh)
  for i := 0; i < octaves*2; i += 2 {
    pyr.layers[i + 2], err = NewOctave(pyr, pyr.layers[i].scale*2, thresh)
    pyr.layers[i + 3], err = NewOctave(pyr, pyr.layers[i + 1].scale*2, thresh)
  }
  if err != nil {
  }
  return pyr, nil
}

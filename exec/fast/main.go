package main

import (
  "fmt"
  "os"
  "strconv"
  "image/jpeg"
  "github.com/awebneck/brisket"
)

func main() {
  status, err := run()
  if err != nil {
    fmt.Printf("An error occurred: %s\n", err.Error())
    os.Exit(status)
  }
}

func run() (int, error) {
  filename := os.Args[1]
  thresh, err := strconv.ParseInt(os.Args[2], 10, 8)
  if err != nil {
    return 1, err
  }
  fmt.Printf("Analyzing image %s for FAST/AGAST points with threshold %d\n", filename, thresh)
  filedata, err := os.Open(filename)
  img, err := jpeg.Decode(filedata)
  fmt.Printf("Image Size: %d x %d\n", img.Bounds().Max.X, img.Bounds().Max.Y)
  fmt.Printf("Converting...\n")
  img = brisket.ConvertToGrayscale(img)
  fmt.Printf("Converted To Grayscale\n")
  // nfile, err := os.Create("gray.jpg")
  // jpeg.Encode(nfile, img, nil)
  if err != nil {
    return 1, err
  }
  return 0, nil
}

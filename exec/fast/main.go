package main

import (
  "fmt"
  "os"
  "strconv"
  "image/jpeg"
  "image/png"
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
  thresh64, err := strconv.Atoi(os.Args[2])
  thresh := uint8(thresh64)
  if err != nil {
    return 1, err
  }
  fmt.Printf("Analyzing image %s for FAST/AGAST points with threshold %d\n", filename, thresh)
  filedata, err := os.Open(filename)
  img, err := jpeg.Decode(filedata)
  fmt.Printf("Image Size: %d x %d\n", img.Bounds().Max.X, img.Bounds().Max.Y)
  fmt.Printf("Converting...\n")
  gray := brisket.ConvertToGrayscale(img)
  fmt.Printf("Converted To Grayscale\n")
  fast := brisket.NewFastFromGray(gray, thresh, brisket.PatternSize9_16)
  fmt.Printf("Keypoints Calculated, Rendering Final\n")
  final := fast.RenderKeypoints()
  nfile, err := os.Create("final.png")
  png.Encode(nfile, final)
  nfile.Close()
  fmt.Printf("Rendering Final Keypoints Only\n")
  finalkponly := fast.RenderKeypointsOnly()
  nfile, err = os.Create("final-kponly.png")
  png.Encode(nfile, finalkponly)
  nfile.Close()
  if err != nil {
    return 1, err
  }
  return 0, nil
}

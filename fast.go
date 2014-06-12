package brisket

import (
  "image"
  "image/color"
)

// 3.5px Bresenham's Circle
var pattern = [16][2]int {
  { 0,  3},
  { 1,  3},
  { 2,  2},
  { 3,  1},
  { 3,  0},
  { 3, -1},
  { 2, -2},
  { 1, -3},
  { 0, -3},
  {-1, -3},
  {-2, -2},
  {-3, -1},
  {-3,  0},
  {-3,  1},
  {-2,  2},
  {-1,  3},
};

type fast struct {
  image image.Image
  thresh *threshTable
  keypoints []*keypoint
}

type keypoint struct {
  point *image.Point
  score int
}

func NewFast(oct *octave, thresh uint8) *fast {
  f := new(fast)
  f.image = oct.image
  f.thresh = NewThreshTable(f.image.Bounds().Max.X, thresh)
  f.keypoints = make([]*keypoint, 0, 2)
  f.findKeypoints()
  return f
}

func NewFastFromGray(gray *image.Gray, thresh uint8) *fast {
  f := new(fast)
  f.image = gray
  f.thresh = NewThreshTable(f.image.Bounds().Max.X, thresh)
  f.keypoints = make([]*keypoint, 0, 2)
  f.findKeypoints()
  return f
}

func (f *fast) findKeypoints() {
  for i := 3; i < f.image.Bounds().Max.X - 2; i++ {
    for j := 3; j < f.image.Bounds().Max.Y - 2; j++ {
      // I lifted this unapologetically straight from the OpenCV source,
      // since Mair et al. describes an ML algorithm to find the optimal
      // tree and - frankly - ain't nobody got time for dat.
      v := f.image.At(i, j).(color.Gray).Y
      tab := f.thresh.table[-v + 255:]
      d := tab[f.image.At(i + pattern[0][0], j + pattern[0][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[8][0], j + pattern[8][1]).(color.Gray).Y]
      if d == 0 {
        continue
      }
      d &= tab[f.image.At(i + pattern[2][0], j + pattern[2][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[10][0], j + pattern[10][1]).(color.Gray).Y]
      d &= tab[f.image.At(i + pattern[4][0], j + pattern[4][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[12][0], j + pattern[12][1]).(color.Gray).Y]
      d &= tab[f.image.At(i + pattern[6][0], j + pattern[6][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[14][0], j + pattern[14][1]).(color.Gray).Y]
      if d == 0 {
        continue
      }
      d &= tab[f.image.At(i + pattern[1][0], j + pattern[1][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[9][0], j + pattern[9][1]).(color.Gray).Y]
      d &= tab[f.image.At(i + pattern[3][0], j + pattern[3][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[11][0], j + pattern[11][1]).(color.Gray).Y]
      d &= tab[f.image.At(i + pattern[5][0], j + pattern[5][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[13][0], j + pattern[13][1]).(color.Gray).Y]
      d &= tab[f.image.At(i + pattern[7][0], j + pattern[7][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[15][0], j + pattern[15][1]).(color.Gray).Y]
      if !f.scanContiguous(1, d, int(v - f.thresh.threshold), i, j) {
        f.scanContiguous(2, d, int(v + f.thresh.threshold), i, j)
      }
    }
  }
}

func (f *fast) scanContiguous(comp, d uint8, thr, x, y int) (bool) {
  if d & comp != 0 {
    count := 0
    for i := 0; i < 25; i++ {
      pv := f.image.At(x + pattern[i % 16][0], y + pattern[i % 16][1]).(color.Gray).Y
      if int(pv) < thr {
        count++
        if count > 8 {
          k := new(keypoint)
          k.point = new(image.Point)
          k.point.X = x
          k.point.Y = y
          k.score = 0
          f.keypoints = append(f.keypoints, k)
          return true
        }
      } else {
        count = 0
      }
    }
  }
  return false
}

func (f *fast) RenderKeypoints() *image.RGBA {
  kpcolor := color.RGBA{0,255,0,255}
  img := ConvertToColor(f.image)
  for i := 0; i < len(f.keypoints); i++ {
    point := f.keypoints[i].point
    img.Set(point.X, point.Y, kpcolor)
  }
  return img
}

func (f *fast) RenderKeypointsOnly() *image.RGBA {
  black := color.RGBA{0,0,0,255}
  kpcolor := color.RGBA{0,255,0,255}
  img := image.NewRGBA(f.image.Bounds())
  for i := 0; i < img.Bounds().Max.X; i++ {
    for j := 0; j < img.Bounds().Max.Y; j++ {
      img.Set(i, j, black)
    }
  }
  for i := 0; i < len(f.keypoints); i++ {
    point := f.keypoints[i].point
    img.Set(point.X, point.Y, kpcolor)
  }
  return img
}

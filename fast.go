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
  keypoints []*image.Point
  scoreMap [][]int
}

func NewFast(oct *octave, thresh uint8) *fast {
  f := new(fast)
  f.image = oct.image
  f.thresh = NewThreshTable(thresh)
  f.keypoints = make([]*image.Point, 0, 2)
  f.scoreMap = make([][]int, f.image.Bounds().Max.X)
  for i := 0; i < f.image.Bounds().Max.X; i++ {
    f.scoreMap[i] = make([]int, f.image.Bounds().Max.Y)
    for j := 0; j < f.image.Bounds().Max.Y; j++ {
      f.scoreMap[i][j] = 0
    }
  }
  f.findKeypoints()
  return f
}

func NewFastFromGray(gray *image.Gray, thresh uint8) *fast {
  f := new(fast)
  f.image = gray
  f.thresh = NewThreshTable(thresh)
  f.keypoints = make([]*image.Point, 0, 2)
  f.scoreMap = make([][]int, f.image.Bounds().Max.X)
  for i := 0; i < f.image.Bounds().Max.X; i++ {
    f.scoreMap[i] = make([]int, f.image.Bounds().Max.Y)
    for j := 0; j < f.image.Bounds().Max.Y; j++ {
      f.scoreMap[i][j] = 0
    }
  }
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

      // Compare pixels at top and bottom. If neither differ by the threshold,
      // disregard the rest and continue to the next pixel
      d := tab[f.image.At(i + pattern[0][0], j + pattern[0][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[8][0], j + pattern[8][1]).(color.Gray).Y]
      if d == 0 {
        continue
      }

      // Compare pixels at left, right, and the 45-degree diagonals. If none differ
      // by the threshold, disregard the rest and continue to the next pixel
      d &= tab[f.image.At(i + pattern[2][0], j + pattern[2][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[10][0], j + pattern[10][1]).(color.Gray).Y]
      d &= tab[f.image.At(i + pattern[4][0], j + pattern[4][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[12][0], j + pattern[12][1]).(color.Gray).Y]
      d &= tab[f.image.At(i + pattern[6][0], j + pattern[6][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[14][0], j + pattern[14][1]).(color.Gray).Y]
      if d == 0 {
        continue
      }

      // Compare all remaining pixels of the Bresenham's circle.
      d &= tab[f.image.At(i + pattern[1][0], j + pattern[1][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[9][0], j + pattern[9][1]).(color.Gray).Y]
      d &= tab[f.image.At(i + pattern[3][0], j + pattern[3][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[11][0], j + pattern[11][1]).(color.Gray).Y]
      d &= tab[f.image.At(i + pattern[5][0], j + pattern[5][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[13][0], j + pattern[13][1]).(color.Gray).Y]
      d &= tab[f.image.At(i + pattern[7][0], j + pattern[7][1]).(color.Gray).Y] |
           tab[f.image.At(i + pattern[15][0], j + pattern[15][1]).(color.Gray).Y]

      // scan for contiguity of differing pixels
      if !f.scanContiguous(2, v, d, int(v + f.thresh.threshold), i, j) {
        f.scanContiguous(1, v, d, int(v - f.thresh.threshold), i, j)
      }
    }
  }
  // Non-maximum suppression. Could be made more efficient, but eh,
  // I'll worry about it later
  unsuppressedKeypoints := make([]*image.Point, 0, 2)
  for i := 0; i < len(f.keypoints); i++ {
    kp := f.keypoints[i]
    score := f.scoreMap[kp.X][kp.Y]
    if score > f.scoreMap[kp.X - 1][kp.Y    ] &&
       score > f.scoreMap[kp.X + 1][kp.Y    ] &&
       score > f.scoreMap[kp.X - 1][kp.Y - 1] &&
       score > f.scoreMap[kp.X    ][kp.Y - 1] &&
       score > f.scoreMap[kp.X + 1][kp.Y - 1] &&
       score > f.scoreMap[kp.X - 1][kp.Y + 1] &&
       score > f.scoreMap[kp.X    ][kp.Y + 1] &&
       score > f.scoreMap[kp.X + 1][kp.Y + 1] {
      unsuppressedKeypoints = append(unsuppressedKeypoints, kp)
    }
  }
  f.keypoints = unsuppressedKeypoints
  f.scoreMap = nil
}

func (f *fast) scanContiguous(comp, value, d uint8, thr, x, y int) (bool) {
  // Ensure that the comparison is still valid
  if d & comp != 0 {
    count := 0
    // Loop over the pattern 1.5 times to ensure comparison of all
    // possible contiguities
    for i := 0; i < 25; i++ {
      pv := f.image.At(x + pattern[i % 16][0], y + pattern[i % 16][1]).(color.Gray).Y
      // If this pixel beats the threshold, bump the count. Otherwise,
      // reset it.
      if int(pv) < thr {
        count++
        // If there are at least 9 contiguous pixels meeting the
        // difference criteria, we have a keypoint. Congratulations!
        if count > 8 {
          k := new(image.Point)
          k.X = x
          k.Y = y
          f.scoreMap[x][y] = f.calculateScore(value, x, y)
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

func (f *fast) calculateScore(value uint8, x, y int) int {
  d := make([]int, 25)
  // Set up an array of distances from the threshold to compare
  for i := 0; i < 25; i++ {
    d[i] = int(value - f.image.At(x + pattern[i % 16][0], y + pattern[i % 16][1]).(color.Gray).Y)
  }
  a0 := int(f.thresh.threshold)
  for i := 0; i < 16; i += 2 {
    a := MinInt(d[i + 1], d[i + 2])
    a = MinInt(a, d[i + 3])
    if a <= a0 {
      continue
    }
    a = MinInt(a, d[i + 4])
    a = MinInt(a, d[i + 5])
    a = MinInt(a, d[i + 6])
    a = MinInt(a, d[i + 7])
    a = MinInt(a, d[i + 8])
    a0 = MaxInt(a0, MinInt(a, d[i]))
    a0 = MaxInt(a0, MinInt(a, d[i + 9]))
  }
  b0 := -a0
  for i := 0; i < 16; i += 2 {
    b := MaxInt(d[i + 1], d[i + 2])
    b = MaxInt(b, d[i + 3])
    b = MaxInt(b, d[i + 4])
    b = MaxInt(b, d[i + 5])
    if b >= b0 {
      continue
    }
    b = MaxInt(b, d[i + 6])
    b = MaxInt(b, d[i + 7])
    b = MaxInt(b, d[i + 8])
    b0 = MinInt(b0, MaxInt(b, d[i]))
    b0 = MinInt(b0, MaxInt(b, d[i + 9]))
  }
  return -b0 - 1
}

// Returns the original image (in grayscale RGB) with the keypoints
// rendered as solid green pixels.
func (f *fast) RenderKeypoints() *image.RGBA {
  kpcolor := color.RGBA{0,255,0,255}
  img := ConvertToColor(f.image)
  for i := 0; i < len(f.keypoints); i++ {
    point := f.keypoints[i]
    img.Set(point.X, point.Y, kpcolor)
  }
  return img
}

// Returns a black image (in RGB) with the keypoints rendered as
// solid green pixels.
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
    point := f.keypoints[i]
    img.Set(point.X, point.Y, kpcolor)
  }
  return img
}

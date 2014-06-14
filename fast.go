package brisket

import (
  "image"
  "image/color"
)

const (
  PatternSize9_16 = 16
  PatternSize5_8 = 8
)

// 3.5px Bresenham's Circle
var pattern9_16 = [16][2]int {
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

// 1.5px Bresenham's Circle
var pattern5_8 = [8][2]int {
  { 0,  1},
  { 1,  1},
  { 1,  0},
  { 1, -1},
  { 0, -1},
  {-1, -1},
  {-1,  0},
  {-1,  1},
};

type fast struct {
  image image.Image
  thresh *threshTable
  keypoints []*fastKeypoint
  patternSize int
  scoreMap [][]int
}

type fastKeypoint struct {
  point *image.Point
  score int
}

func NewFast(oct *octave, thresh uint8, patternSize int) *fast {
  f := new(fast)
  f.image = oct.image
  f.thresh = NewThreshTable(thresh)
  f.keypoints = make([]*fastKeypoint, 0, 2)
  f.scoreMap = make([][]int, f.image.Bounds().Max.X)
  f.patternSize = patternSize
  for i := 0; i < f.image.Bounds().Max.X; i++ {
    f.scoreMap[i] = make([]int, f.image.Bounds().Max.Y)
    for j := 0; j < f.image.Bounds().Max.Y; j++ {
      f.scoreMap[i][j] = 0
    }
  }
  f.findKeypoints()
  return f
}

func NewFastFromGray(gray *image.Gray, thresh uint8, patternSize int) *fast {
  f := new(fast)
  f.image = gray
  f.thresh = NewThreshTable(thresh)
  f.keypoints = make([]*fastKeypoint, 0, 2)
  f.scoreMap = make([][]int, f.image.Bounds().Max.X)
  f.patternSize = patternSize
  for i := 0; i < f.image.Bounds().Max.X; i++ {
    f.scoreMap[i] = make([]int, f.image.Bounds().Max.Y)
    for j := 0; j < f.image.Bounds().Max.Y; j++ {
      f.scoreMap[i][j] = 0
    }
  }
  f.findKeypoints()
  return f
}

func (f *fast) calculateDistance5_8(v uint8, tab []uint8, x, y int) uint8 {
  // Compare pixels at top and bottom. If neither differ by the threshold,
  // disregard the rest and continue to the next pixel
  d := tab[f.image.At(x + pattern5_8[0][0], y + pattern5_8[0][1]).(color.Gray).Y] |
       tab[f.image.At(x + pattern5_8[4][0], y + pattern5_8[4][1]).(color.Gray).Y]
  if d == 0 {
    return 0
  }

  // Compare pixels at left and right. If none differ by the threshold, disregard
  // the rest and continue to the next pixel
  d &= tab[f.image.At(x + pattern5_8[2][0], y + pattern5_8[2][1]).(color.Gray).Y] |
       tab[f.image.At(x + pattern5_8[6][0], y + pattern5_8[6][1]).(color.Gray).Y]
  if d == 0 {
    return 0
  }

  // Compare pixels at 45-degree angles. If none differ by the threshold, disregard
  // the rest and continue to the next pixel
  d &= tab[f.image.At(x + pattern5_8[1][0], y + pattern5_8[1][1]).(color.Gray).Y] |
       tab[f.image.At(x + pattern5_8[5][0], y + pattern5_8[5][1]).(color.Gray).Y]
  d &= tab[f.image.At(x + pattern5_8[3][0], y + pattern5_8[3][1]).(color.Gray).Y] |
       tab[f.image.At(x + pattern5_8[7][0], y + pattern5_8[7][1]).(color.Gray).Y]
  return d
}

func (f *fast) calculateDistance9_16(v uint8, tab []uint8, x, y int) uint8 {
  // Compare pixels at top and bottom. If neither differ by the threshold,
  // disregard the rest and continue to the next pixel
  d := tab[f.image.At(x + pattern9_16[0][0], y + pattern9_16[0][1]).(color.Gray).Y] |
       tab[f.image.At(x + pattern9_16[8][0], y + pattern9_16[8][1]).(color.Gray).Y]
  if d == 0 {
    return 0
  }

  // Compare pixels at left, right, and the 45-degree diagonals. If none differ
  // by the threshold, disregard the rest and continue to the next pixel
  d &= tab[f.image.At(x + pattern9_16[2 ][0], y + pattern9_16[2 ][1]).(color.Gray).Y] |
       tab[f.image.At(x + pattern9_16[10][0], y + pattern9_16[10][1]).(color.Gray).Y]
  d &= tab[f.image.At(x + pattern9_16[4 ][0], y + pattern9_16[4 ][1]).(color.Gray).Y] |
       tab[f.image.At(x + pattern9_16[12][0], y + pattern9_16[12][1]).(color.Gray).Y]
  d &= tab[f.image.At(x + pattern9_16[6 ][0], y + pattern9_16[6 ][1]).(color.Gray).Y] |
       tab[f.image.At(x + pattern9_16[14][0], y + pattern9_16[14][1]).(color.Gray).Y]
  if d == 0 {
    return 0
  }

  // Compare all remaining pixels of the Bresenham's circle.
  d &= tab[f.image.At(x + pattern9_16[1 ][0], y + pattern9_16[1 ][1]).(color.Gray).Y] |
       tab[f.image.At(x + pattern9_16[9 ][0], y + pattern9_16[9 ][1]).(color.Gray).Y]
  d &= tab[f.image.At(x + pattern9_16[3 ][0], y + pattern9_16[3 ][1]).(color.Gray).Y] |
       tab[f.image.At(x + pattern9_16[11][0], y + pattern9_16[11][1]).(color.Gray).Y]
  d &= tab[f.image.At(x + pattern9_16[5 ][0], y + pattern9_16[5 ][1]).(color.Gray).Y] |
       tab[f.image.At(x + pattern9_16[13][0], y + pattern9_16[13][1]).(color.Gray).Y]
  d &= tab[f.image.At(x + pattern9_16[7 ][0], y + pattern9_16[7 ][1]).(color.Gray).Y] |
       tab[f.image.At(x + pattern9_16[15][0], y + pattern9_16[15][1]).(color.Gray).Y]
  return d
}

func (f *fast) findKeypoints() {
  for i := 3; i < f.image.Bounds().Max.X - 2; i++ {
    for j := 3; j < f.image.Bounds().Max.Y - 2; j++ {
      var d uint8
      // I lifted this unapologetically straight from the OpenCV source,
      // since Mair et al. describes an ML algorithm to find the optimal
      // tree and - frankly - ain't nobody got time for dat.
      v := f.image.At(i, j).(color.Gray).Y
      tab := f.thresh.table[-v + 255:]
      // get binary keypoint status of pixel, and its value
      if f.patternSize == PatternSize9_16 {
        d = f.calculateDistance9_16(v, tab, i, j)
      } else {
        d = f.calculateDistance5_8(v, tab, i, j)
      }
      if d == 0 {
        continue
      }

      // scan for contiguity of differing pixels
      if !f.scanContiguous(2, v, d, i, j) {
        f.scanContiguous(1, v, d, i, j)
      }
    }
  }
  // Non-maximum suppression. Could be made more efficient, but eh,
  // I'll worry about it later. Looks up scores in the scoreMap field
  // and ensures that the keypoint in question is the local maximum.
  unsuppressedKeypoints := make([]*fastKeypoint, 0, 2)
  for i := 0; i < len(f.keypoints); i++ {
    kp := f.keypoints[i]
    kpp := kp.point
    score := f.keypoints[i].score
    if score > f.scoreMap[kpp.X - 1][kpp.Y    ] &&
       score > f.scoreMap[kpp.X + 1][kpp.Y    ] &&
       score > f.scoreMap[kpp.X - 1][kpp.Y - 1] &&
       score > f.scoreMap[kpp.X    ][kpp.Y - 1] &&
       score > f.scoreMap[kpp.X + 1][kpp.Y - 1] &&
       score > f.scoreMap[kpp.X - 1][kpp.Y + 1] &&
       score > f.scoreMap[kpp.X    ][kpp.Y + 1] &&
       score > f.scoreMap[kpp.X + 1][kpp.Y + 1] {
      unsuppressedKeypoints = append(unsuppressedKeypoints, kp)
    }
  }
  f.keypoints = unsuppressedKeypoints
  f.scoreMap = nil
}

func (f *fast) scanContiguous(comp, value, d uint8, x, y int) (bool) {
  var thr int
  k := f.patternSize / 2
  n := f.patternSize + k + 1
  if comp % 2 == 0 {
    thr = int(value + f.thresh.threshold)
  } else {
    thr = int(value - f.thresh.threshold)
  }
  // Ensure that the comparison is still valid
  if d & comp != 0 {
    count := 0
    // Loop over the pattern 1.5 times to ensure comparison of all
    // possible contiguities
    for i := 0; i < n; i++ {
      var pv uint8
      if f.patternSize == PatternSize9_16 {
        pv = f.image.At(x + pattern9_16[i % f.patternSize][0], y + pattern9_16[i % f.patternSize][1]).(color.Gray).Y
      } else {
        pv = f.image.At(x + pattern5_8[i % f.patternSize][0], y + pattern5_8[i % f.patternSize][1]).(color.Gray).Y
      }
      // If this pixel beats the threshold, bump the count. Otherwise,
      // reset it.
      if int(pv) < thr {
        count++
        // If there are at least 9 contiguous pixels meeting the
        // difference criteria, we have a keypoint. Congratulations!
        if count > k {
          kp := new(fastKeypoint)
          kp.point = new(image.Point)
          kp.point.X = x
          kp.point.Y = y
          // Calculate the score and add it to the scoremap
          f.scoreMap[x][y] = f.calculateScore(value, x, y)
          kp.score = f.scoreMap[x][y]
          f.keypoints = append(f.keypoints, kp)
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
  k := f.patternSize / 2
  n := f.patternSize + k + 1

  if f.patternSize == PatternSize9_16 {
    return f.calculateScore9_16(n, value, x, y)
  } else {
    return f.calculateScore5_8(n, value, x, y)
  }
}

func (f *fast) calculateScore9_16(n int, value uint8, x, y int) int {
  // Set up an array of distances from the threshold to compare
  d := make([]int, n)
  for i := 0; i < n; i++ {
    d[i] = int(value - f.image.At(x + pattern9_16[i % f.patternSize][0], y + pattern9_16[i % f.patternSize][1]).(color.Gray).Y)
  }

  a0 := int(f.thresh.threshold)
  for i := 0; i < f.patternSize; i += 2 {
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
  for i := 0; i < f.patternSize; i += 2 {
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

func (f *fast) calculateScore5_8(n int, value uint8, x, y int) int {
  // Set up an array of distances from the threshold to compare
  d := make([]int, n)
  for i := 0; i < n; i++ {
    d[i] = int(value - f.image.At(x + pattern5_8[i % f.patternSize][0], y + pattern5_8[i % f.patternSize][1]).(color.Gray).Y)
  }

  a0 := int(f.thresh.threshold)
  for i := 0; i < f.patternSize; i += 2 {
    a := MinInt(d[i + 1], d[i + 2])
    if a <= a0 {
      continue
    }
    a = MinInt(a, d[i + 3])
    a = MinInt(a, d[i + 4])
    a0 = MaxInt(a0, MinInt(a, d[i]))
    a0 = MaxInt(a0, MinInt(a, d[i + 5]))
  }
  b0 := -a0
  for i := 0; i < f.patternSize; i += 2 {
    b := MaxInt(d[i + 1], d[i + 2])
    b = MaxInt(b, d[i + 3])
    if b >= b0 {
      continue
    }
    b = MaxInt(b, d[i + 4])
    b0 = MinInt(b0, MaxInt(b, d[i]))
    b0 = MinInt(b0, MaxInt(b, d[i + 5]))
  }
  return -b0 - 1
}

// Returns the original image (in grayscale RGB) with the keypoints
// rendered as solid green pixels.
func (f *fast) RenderKeypoints() *image.RGBA {
  kpcolor := color.RGBA{0,255,0,255}
  img := ConvertToColor(f.image)
  for i := 0; i < len(f.keypoints); i++ {
    point := f.keypoints[i].point
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
    point := f.keypoints[i].point
    img.Set(point.X, point.Y, kpcolor)
  }
  return img
}

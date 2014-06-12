package brisket

type threshTable struct {
  threshold uint8
  table []uint8
}

func NewThreshTable(cols int, threshold uint8) *threshTable {
  tab := new(threshTable)
  tab.threshold = threshold
  tab.table = make([]uint8, 512)
  threshint := int(threshold)
  for i := -255; i < 256; i++ {
    if i < -threshint {
      tab.table[i + 255] = 1
    } else if i > threshint {
      tab.table[i + 255] = 2
    } else {
      tab.table[i + 255] = 0
    }
  }
  return tab
}

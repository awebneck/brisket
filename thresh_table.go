package brisket

type threshTable struct {
  threshold uint8
  table []uint8
}

// Constructs the table for binary comparison based on
// the threshold value. Once sliced to the nexus pixel's
// value, the resulting slice value at the query pixel's
// value will return 1 if the query is less than the
// nexus by at least the threshold, 2 if the query is
// greater than the nexus by at least the threshold, and
// 0 otherwise.
func NewThreshTable(threshold uint8) *threshTable {
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

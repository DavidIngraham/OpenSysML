package export

const (
	largeGraphTripleThreshold = 1 << 18
	largeGraphTripleQuantum   = 1 << 16
)

// graphTripleCapacity keeps nearby large model versions in the same allocation class.
// Small buffers stay exact; rounding adds less than one quarter of the requested size.
func graphTripleCapacity(count int) int {
	if count < largeGraphTripleThreshold {
		return count
	}
	remainder := count % largeGraphTripleQuantum
	if remainder == 0 {
		return count
	}
	extra := largeGraphTripleQuantum - remainder
	maxInt := int(^uint(0) >> 1)
	if count > maxInt-extra {
		return count
	}
	return count + extra
}

package export

import "testing"

func TestGraphTripleCapacityKeepsSmallBuffersExact(t *testing.T) {
	for _, count := range []int{0, 1, 12, 25000, largeGraphTripleThreshold - 1} {
		if got := graphTripleCapacity(count); got != count {
			t.Fatalf("small capacity %d changed to %d", count, got)
		}
	}
}

func TestGraphTripleCapacityBoundsSpareStorage(t *testing.T) {
	for count := largeGraphTripleThreshold; count < largeGraphTripleThreshold*12; count += 31 {
		got := graphTripleCapacity(count)
		if got < count || got-count >= largeGraphTripleQuantum || got-count > count/4 {
			t.Fatalf("capacity %d for %d violates bounded spare storage", got, count)
		}
		if got%largeGraphTripleQuantum != 0 {
			t.Fatalf("capacity %d is not a complete large allocation class", got)
		}
	}
}

func TestGraphTripleCapacitySharesNearbyModelVersions(t *testing.T) {
	for _, count := range []int{436684, 448119, 450000, 458752} {
		if got := graphTripleCapacity(count); got != 458752 {
			t.Fatalf("nearby model %d uses capacity %d instead of shared class", count, got)
		}
	}
	if got := graphTripleCapacity(458753); got != 524288 {
		t.Fatalf("next class is %d", got)
	}
}

func TestGraphTripleCapacityDoesNotOverflow(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, count := range []int{maxInt, maxInt - 1, maxInt - largeGraphTripleQuantum + 2, maxInt - largeGraphTripleQuantum + 1} {
		if got := graphTripleCapacity(count); got != count {
			t.Fatalf("unroundable or aligned near-limit count %d changed to %d", count, got)
		}
	}
}

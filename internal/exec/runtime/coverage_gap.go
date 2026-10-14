package runtime

import (
	"fmt"
	"slices"
	"strings"
)

// CoverageGap is a construct whose unordered alternatives a run performs in one
// order without a choice point, so exploring or checking it enumerates fewer
// orders than the library admits and neither may report its outcome set complete.
type CoverageGap int

const (
	// GapBlockBodyRepetition: the performances of a repeated step in a loop or if
	// body are each run as one move, so no interleaving of them is ever a schedule.
	GapBlockBodyRepetition CoverageGap = iota + 1
)

// String is the gap's identifier, as a conformance case's `exploreGaps` names it.
func (g CoverageGap) String() string {
	switch g {
	case GapBlockBodyRepetition:
		return "block-body-repetition"
	}
	return fmt.Sprintf("coverage gap %d", int(g))
}

// Description says what the gap left unenumerated, as a status prints it.
func (g CoverageGap) Description() string {
	switch g {
	case GapBlockBodyRepetition:
		return "the performances of a repeated step in a loop or if body are each run as one move, so their interleavings were not explored"
	}
	return g.String()
}

// DescribeGaps joins the gaps' descriptions, in their order.
func DescribeGaps(gaps []CoverageGap) string {
	parts := make([]string, len(gaps))
	for i, gap := range gaps {
		parts[i] = gap.Description()
	}
	return strings.Join(parts, "; ")
}

// GapNames lists the gaps' identifiers, in their order.
func GapNames(gaps []CoverageGap) []string {
	names := make([]string, len(gaps))
	for i, gap := range gaps {
		names[i] = gap.String()
	}
	return names
}

// noteGap records a construct the run performed in one order without a choice
// point: explore and check then report what they enumerated short of complete.
func (c *Context) noteGap(gap CoverageGap) {
	if c.coverageGaps == nil {
		c.coverageGaps = make(map[CoverageGap]bool)
	}
	c.coverageGaps[gap] = true
}

// gaps lists the distinct gaps recorded, in canonical order.
func (c *Context) gaps() []CoverageGap {
	out := make([]CoverageGap, 0, len(c.coverageGaps))
	for gap := range c.coverageGaps {
		out = append(out, gap)
	}
	slices.Sort(out)
	return out
}

// mergeGaps adds the gaps not yet in the list and returns it in canonical order.
func mergeGaps(into []CoverageGap, gaps []CoverageGap) []CoverageGap {
	for _, gap := range gaps {
		if !slices.Contains(into, gap) {
			into = append(into, gap)
		}
	}
	slices.Sort(into)
	return into
}

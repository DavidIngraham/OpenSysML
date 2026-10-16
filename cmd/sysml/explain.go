package main

import (
	"fmt"
	"io"

	"github.com/Open-MBEE/OpenSysML/internal/check/codes"
)

// explainDiagnostics answers -explain: each code named is explained in turn, and
// no code at all lists the catalogue. An unknown code is a usage error, as a
// misspelt flag is, so a script that asked about a code that was renamed hears it.
func explainDiagnostics(out, errw io.Writer, names []string) int {
	if len(names) == 0 {
		fmt.Fprint(out, codes.List())
		return 0
	}
	status := 0
	for i, name := range names {
		text, err := codes.Explain(name)
		if err != nil {
			fmt.Fprintf(errw, "sysml: %v\n", err)
			status = 2
			continue
		}
		if i > 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprint(out, text)
	}
	return status
}

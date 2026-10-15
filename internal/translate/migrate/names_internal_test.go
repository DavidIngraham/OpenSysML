package migrate

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"
)

func TestQueryNameMetadataPathSegment(t *testing.T) {
	for _, tc := range []struct {
		names      []string
		definition []string
		feature    string
	}{
		{[]string{"Frank's Braindump", "Composition Example", "review's grade"}, []string{`Frank\'s Braindump`, "Composition Example"}, `review\'s grade`},
		{[]string{"Review Profile", "Reviewed", "Safety::Scope"}, []string{"Review Profile", "Reviewed"}, "Safety::Scope"},
		{[]string{"Profile::Review", "part", "level"}, []string{"Profile::Review", "part"}, "level"},
	} {
		segments := make([]string, len(tc.names))
		for i, name := range tc.names {
			segments[i] = pathSegment(queryName(name))
		}
		path := strings.Join(segments, "::")
		definition, feature, ok := source.MetadataPathSegments(path)
		if !ok || !reflect.DeepEqual(definition, tc.definition) || feature != tc.feature {
			t.Errorf("MetadataPathSegments(%q) = %v, %q, %v", path, definition, feature, ok)
		}
	}
}

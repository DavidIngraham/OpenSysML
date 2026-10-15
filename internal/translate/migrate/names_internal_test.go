package migrate

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"
)

func TestQueryNameMetadataPathSegment(t *testing.T) {
	path := strings.Join([]string{
		pathSegment(queryName("Frank's Braindump")),
		pathSegment(queryName("Composition Example")),
		pathSegment(queryName("review's grade")),
	}, "::")
	definition, feature, ok := source.MetadataPathSegments(path)
	if !ok || !reflect.DeepEqual(definition, []string{`Frank\'s Braindump`, "Composition Example"}) || feature != `review\'s grade` {
		t.Errorf("MetadataPathSegments(%q) = %v, %q, %v", path, definition, feature, ok)
	}
}

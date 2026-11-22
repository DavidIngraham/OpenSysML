package view

import (
	"reflect"
	"testing"
)

func TestMatrixGridRecordsPreserveOriginsAndSnapshotIsolation(t *testing.T) {
	rendering := render(t, "matrix.sysml", "MatrixViews::allRelations")
	data := rendering.Data()
	if len(data.ColumnElements) != len(data.Columns)-1 || len(data.RelationshipCells) == 0 {
		t.Fatalf("semantic matrix records absent: %+v", data)
	}
	for _, column := range data.ColumnElements {
		if column.Index < 1 || column.Index >= len(data.Columns) || column.Name != data.Columns[column.Index] || !column.Origin.Located() {
			t.Fatalf("column origin or axis missing: %+v", column)
		}
	}
	foundDuplicate := false
	for _, cell := range data.RelationshipCells {
		if cell.Row < 0 || cell.Row >= len(data.Rows) || cell.Column < 1 || cell.Column >= len(data.Columns) {
			t.Fatalf("bad cell indices: %+v", cell)
		}
		counts := map[string]int{}
		for _, relation := range cell.Relationships {
			if !relation.Origin.Located() || relation.SourceOrigin != data.Rows[cell.Row].Origin || relation.TargetOrigin != data.ColumnElements[cell.Column-1].Origin {
				t.Fatalf("relationship declaration or direction lost: %+v", relation)
			}
			counts[relation.Kind]++
			if counts[relation.Kind] > 1 {
				foundDuplicate = true
			}
		}
	}
	if !foundDuplicate {
		t.Fatal("fixture's repeated refinement and derivation relationships collapsed")
	}
	before := rendering.Data()
	data.ColumnElements[0].Name = "changed"
	data.RelationshipCells[0].Relationships[0].Label = "changed"
	clone := rendering.Clone()
	clone.ColumnElements[0].Name = "changed"
	clone.RelationshipCells[0].Relationships[0].Label = "changed"
	if !reflect.DeepEqual(before, rendering.Data()) {
		t.Fatal("editing exported data or a clone mutated the native rendering")
	}
}

package view

// ColumnElement identifies a semantic matrix target. Index includes the leading
// label column: the first target column has index 1.
type ColumnElement struct {
	Index          int
	Name           string
	SemanticID     string
	QualifiedName  string
	IdentityStatus string
	Origin         Origin
}

// GridRelationship retains a declaration and its native directed endpoints.
// Matrix derive runs original-to-derived, independent of diagram arrow conventions.
type GridRelationship struct {
	SemanticID       string
	QualifiedName    string
	IdentityStatus   string
	SourceSemanticID string
	TargetSemanticID string
	Kind             string
	Label            string
	Origin           Origin
	SourceOrigin     Origin
	TargetOrigin     Origin
}

// RelationshipCell contains every relationship represented at a nonempty cell.
// Row and Column index Rows and Columns directly, including the label column.
type RelationshipCell struct {
	Row           int
	Column        int
	Relationships []GridRelationship
}

func cloneRelationshipCells(cells []RelationshipCell) []RelationshipCell {
	out := append([]RelationshipCell(nil), cells...)
	for i := range out {
		out[i].Relationships = append([]GridRelationship(nil), cells[i].Relationships...)
	}
	return out
}

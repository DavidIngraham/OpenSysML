package opensysml

import (
	"context"
	"fmt"

	pb "github.com/Open-MBEE/OpenSysML/api/proto"
)

// RenderViewOption configures a view rendering.
type RenderViewOption func(*renderViewOptions)

type renderViewOptions struct {
	ports    string
	artifact *RenderArtifactOptions
}

// WithFullPorts includes ports whether or not a connector uses them.
func WithFullPorts() RenderViewOption {
	return func(options *renderViewOptions) { options.ports = "full" }
}

// RenderArtifactOptions selects a formatted artifact; Form is required.
type RenderArtifactOptions struct {
	Form         string
	Palette      string
	DrawingStyle string
	Overlay      string
	Direction    string
	Unplaced     string
	Width        int32
	LinkTemplate string
}

// WithRenderArtifact requests formatted text alongside the diagram data.
func WithRenderArtifact(artifact RenderArtifactOptions) RenderViewOption {
	return func(options *renderViewOptions) { options.artifact = &artifact }
}

// RenderArtifact is formatted UTF-8 diagram or table source.
type RenderArtifact struct {
	Form      string
	MediaType string
	Text      string
}

// RenderedView is the lossless diagram data returned for a declared or
// targeted pseudo-view.
type RenderedView struct {
	View     string
	Kind     string
	Stated   string
	Nodes    []RenderNode
	Edges    []RenderEdge
	Columns  []string
	Rows     []RenderRow
	Canvas   *RenderCanvas
	Notes    []RenderNote
	Notices  []string
	Artifact *RenderArtifact
}

// RenderNode is a view node with optional geometry and style.
type RenderNode struct {
	ID              string
	Kind            string
	Name            string
	NameSynthesized bool
	Type            string
	Detail          string
	Text            string
	StandIn         bool
	Parent          string
	Verdict         string
	Ports           []RenderPort
	Origin          *Span
	Geometry        *RenderGeometry
	Style           *RenderStyle
}

// RenderPort is a port shown on a rendered node.
type RenderPort struct {
	ID        string
	Name      string
	Type      string
	Direction string
}

// RenderEdge is a connection between rendered nodes.
type RenderEdge struct {
	From     string
	To       string
	FromPort string
	ToPort   string
	Label    string
	Name     string
	Kind     string
	Origin   *Span
	Route    []RenderPoint
	Style    *RenderStyle
}

// RenderGeometry is a node's position and optional size.
type RenderGeometry struct {
	X         float64
	Y         float64
	Width     float64
	Height    float64
	HasSize   bool
	Collapsed bool
}

// RenderCanvas is the optional canvas bounds for a view.
type RenderCanvas struct {
	Unit    string
	Width   float64
	Height  float64
	HasSize bool
}

// RenderStyle is a rendered node or edge's optional style.
type RenderStyle struct {
	Fill     string
	Line     string
	Text     string
	Font     string
	FontSize float64
	Bold     bool
	Italic   bool
}

// RenderPoint is a point on an edge route.
type RenderPoint struct {
	X float64
	Y float64
}

// RenderRow is one row of a table-style view.
type RenderRow struct {
	Cells  []string
	Origin *Span
}

// RenderNote is an annotation in a rendered view.
type RenderNote struct {
	Text     string
	Anchor   string
	EdgeFrom string
	EdgeTo   string
	X        float64
	Y        float64
	Width    float64
	Height   float64
	HasSize  bool
	Origin   *Span
}

// RenderView renders a named view. Ports are minimal unless WithFullPorts is
// supplied.
func (c *client) RenderView(ctx context.Context, model *Model, viewName string, opts ...RenderViewOption) (*RenderedView, error) {
	hash, err := c.call(model)
	if err != nil {
		return nil, err
	}
	if err := c.requireCapabilities(ctx, CapabilityRenderView); err != nil {
		return nil, err
	}
	var options renderViewOptions
	for _, opt := range opts {
		opt(&options)
	}
	var artifact *pb.RenderArtifactOptions
	if a := options.artifact; a != nil {
		if err := c.requireCapabilities(ctx, CapabilityRenderViewArtifact); err != nil {
			return nil, err
		}
		artifact = &pb.RenderArtifactOptions{Form: a.Form, Palette: a.Palette, DrawingStyle: a.DrawingStyle, Overlay: a.Overlay, Direction: a.Direction, Unplaced: a.Unplaced, Width: a.Width, LinkTemplate: a.LinkTemplate}
	}
	response, err := c.caller.renderView(ctx, &pb.RenderViewRequest{
		ModelHash: hash,
		View:      viewName,
		Ports:     options.ports,
		Artifact:  artifact,
	})
	if err != nil {
		return nil, err
	}
	if artifact != nil && response.Artifact == nil {
		return nil, fmt.Errorf("RenderView response omitted the requested artifact")
	}
	return renderedViewFromProto(response), nil
}

func renderedViewFromProto(response *pb.RenderViewResponse) *RenderedView {
	result := &RenderedView{
		View: response.View, Kind: response.Kind, Stated: response.Stated,
		Columns: append([]string(nil), response.Columns...), Notices: append([]string(nil), response.Notices...),
	}
	if a := response.Artifact; a != nil {
		result.Artifact = &RenderArtifact{Form: a.Form, MediaType: a.MediaType, Text: a.Text}
	}
	for _, node := range response.Nodes {
		converted := RenderNode{
			ID: node.Id, Kind: node.Kind, Name: node.Name, NameSynthesized: node.NameSynthesized,
			Type: node.Type, Detail: node.Detail, Text: node.Text, StandIn: node.StandIn, Parent: node.Parent,
			Origin: renderSpanFromProto(node.Origin), Verdict: node.Verdict,
		}
		for _, port := range node.Ports {
			converted.Ports = append(converted.Ports, RenderPort{
				ID: port.Id, Name: port.Name, Type: port.Type, Direction: port.Direction,
			})
		}
		if node.Geometry != nil {
			converted.Geometry = &RenderGeometry{
				X: node.Geometry.X, Y: node.Geometry.Y, Width: node.Geometry.Width, Height: node.Geometry.Height,
				HasSize: node.Geometry.HasSize, Collapsed: node.Geometry.Collapsed,
			}
		}
		converted.Style = renderStyleFromProto(node.Style)
		result.Nodes = append(result.Nodes, converted)
	}
	for _, edge := range response.Edges {
		converted := RenderEdge{
			From: edge.From, To: edge.To, FromPort: edge.FromPort, ToPort: edge.ToPort,
			Label: edge.Label, Name: edge.Name, Kind: edge.Kind, Origin: renderSpanFromProto(edge.Origin),
			Style: renderStyleFromProto(edge.Style),
		}
		for _, point := range edge.Route {
			converted.Route = append(converted.Route, RenderPoint{X: point.X, Y: point.Y})
		}
		result.Edges = append(result.Edges, converted)
	}
	for _, row := range response.Rows {
		result.Rows = append(result.Rows, RenderRow{Cells: append([]string(nil), row.Cells...), Origin: renderSpanFromProto(row.Origin)})
	}
	for _, note := range response.Notes {
		result.Notes = append(result.Notes, RenderNote{
			Text: note.Text, Anchor: note.Anchor, EdgeFrom: note.EdgeFrom, EdgeTo: note.EdgeTo,
			X: note.X, Y: note.Y, Width: note.Width, Height: note.Height, HasSize: note.HasSize,
			Origin: renderSpanFromProto(note.Origin),
		})
	}
	if response.Canvas != nil {
		result.Canvas = &RenderCanvas{
			Unit: response.Canvas.Unit, Width: response.Canvas.Width, Height: response.Canvas.Height,
			HasSize: response.Canvas.HasSize,
		}
	}
	return result
}

func renderStyleFromProto(style *pb.RenderStyle) *RenderStyle {
	if style == nil {
		return nil
	}
	return &RenderStyle{
		Fill: style.Fill, Line: style.Line, Text: style.Text, Font: style.Font, FontSize: style.FontSize,
		Bold: style.Bold, Italic: style.Italic,
	}
}

func renderSpanFromProto(span *pb.Span) *Span {
	if span == nil {
		return nil
	}
	return &Span{
		File: span.File, StartLine: int(span.StartLine), StartCol: int(span.StartCol),
		EndLine: int(span.EndLine), EndCol: int(span.EndCol),
	}
}

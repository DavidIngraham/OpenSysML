package grpc

import (
	"fmt"
	"slices"

	pb "github.com/Open-MBEE/OpenSysML/api/proto"
	"github.com/Open-MBEE/OpenSysML/internal/ir/view"
)

func renderArtifactOptions(request *pb.RenderArtifactOptions, ports view.Ports) (view.Options, view.Overlay, error) {
	options := view.Options{Ports: ports}
	if request == nil {
		return options, "", nil
	}
	if !slices.Contains(view.Forms(), view.Form(request.Form)) {
		return options, "", fmt.Errorf("unknown rendering form %q; ask for %s", request.Form, view.FormNames(view.Forms()))
	}
	var ok bool
	if request.Palette != "" {
		options.Palette, ok = view.ParsePalette(request.Palette)
		if !ok {
			return options, "", &view.UnknownPaletteError{Name: request.Palette}
		}
	}
	options.Style, ok = view.ParseDrawingStyle(request.DrawingStyle)
	if !ok {
		return options, "", &view.UnknownDrawingStyleError{Name: request.DrawingStyle}
	}
	if request.Unplaced != "" {
		options.Unplaced, ok = view.ParseUnplaced(request.Unplaced)
		if !ok {
			return options, "", &view.UnknownUnplacedError{Name: request.Unplaced}
		}
	}
	if request.Direction != "" {
		options.Direction, ok = view.ParseDirection(request.Direction)
		if !ok {
			return options, "", fmt.Errorf("unknown direction %q; ask for TB, LR, RL or BT", request.Direction)
		}
	}
	if request.Width < 0 {
		return options, "", fmt.Errorf("width must be nonnegative")
	}
	options.Width = int(request.Width)
	if err := view.ParseLinkTemplate(request.LinkTemplate); err != nil {
		return options, "", err
	}
	options.Links.Template = request.LinkTemplate
	overlay, ok := view.ParseOverlay(request.Overlay)
	if !ok {
		return options, "", fmt.Errorf("unknown overlay %q; ask for %s", request.Overlay, view.OverlayNames())
	}
	return options, overlay, nil
}

func renderArtifactMediaType(form string) string {
	switch view.Form(form) {
	case view.FormMarkdown:
		return "text/markdown; charset=utf-8"
	case view.FormCSV:
		return "text/csv; charset=utf-8"
	case view.FormTSV:
		return "text/tab-separated-values; charset=utf-8"
	default:
		return "text/plain; charset=utf-8"
	}
}

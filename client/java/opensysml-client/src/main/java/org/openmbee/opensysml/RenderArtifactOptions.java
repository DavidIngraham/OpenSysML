package org.openmbee.opensysml;

import java.util.Objects;

/** Options for formatted UTF-8 view source; the form is required. */
public record RenderArtifactOptions(
    String form,
    String palette,
    String drawingStyle,
    String overlay,
    String direction,
    String unplaced,
    int width,
    String linkTemplate) {
  public RenderArtifactOptions {
    Objects.requireNonNull(form, "form");
    Objects.requireNonNull(palette, "palette");
    Objects.requireNonNull(drawingStyle, "drawingStyle");
    Objects.requireNonNull(overlay, "overlay");
    Objects.requireNonNull(direction, "direction");
    Objects.requireNonNull(unplaced, "unplaced");
    Objects.requireNonNull(linkTemplate, "linkTemplate");
  }

  /** Requests a form using the renderer defaults. */
  public RenderArtifactOptions(String form) {
    this(form, "", "", "", "", "", 0, "");
  }

  org.openmbee.opensysml.proto.RenderArtifactOptions wire() {
    return org.openmbee.opensysml.proto.RenderArtifactOptions.newBuilder()
        .setForm(form)
        .setPalette(palette)
        .setDrawingStyle(drawingStyle)
        .setOverlay(overlay)
        .setDirection(direction)
        .setUnplaced(unplaced)
        .setWidth(width)
        .setLinkTemplate(linkTemplate)
        .build();
  }
}

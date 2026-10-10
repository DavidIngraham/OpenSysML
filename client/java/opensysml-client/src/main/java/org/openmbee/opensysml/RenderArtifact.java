package org.openmbee.opensysml;

import java.util.Objects;

/** Formatted UTF-8 diagram or table source. */
public record RenderArtifact(String form, String mediaType, String text) {
  public RenderArtifact {
    Objects.requireNonNull(form, "form");
    Objects.requireNonNull(mediaType, "mediaType");
    Objects.requireNonNull(text, "text");
  }
}

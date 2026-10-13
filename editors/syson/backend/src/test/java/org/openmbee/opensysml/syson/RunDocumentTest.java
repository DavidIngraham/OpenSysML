package org.openmbee.opensysml.syson;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.jupiter.api.Test;
import org.openmbee.opensysml.DocumentForm;
import org.openmbee.opensysml.RenderedDocument;
import org.openmbee.opensysml.syson.run.RunDocument;
import org.openmbee.opensysml.syson.run.RunDocumentForm;

class RunDocumentTest {

    @Test
    void takesTheFileNameTheServicePlans() {
        RunDocument document = RunDocument.of(RunDocumentForm.HTML,
                new RenderedDocument(DocumentForm.HTML, "<html></html>", "Reports-Mass.20Report~a1b2.html"),
                "Reports::'Mass Report'");

        assertThat(document.fileName()).isEqualTo("Reports-Mass.20Report~a1b2.html");
    }

    @Test
    void fallsBackToTheDocumentsOwnNameWhenTheServiceNamesNone() {
        RunDocument document = RunDocument.of(RunDocumentForm.MARKDOWN,
                new RenderedDocument(DocumentForm.MARKDOWN, "# Mass"), "Reports::'Mass Report'");

        assertThat(document.fileName()).isEqualTo("Mass Report.md");
        assertThat(RunDocument.of(RunDocumentForm.HTML, new RenderedDocument(DocumentForm.HTML, ""), "'a/b:c'")
                .fileName()).isEqualTo("a_b_c.html");
    }
}

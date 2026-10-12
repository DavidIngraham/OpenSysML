package org.openmbee.opensysml.syson;

import static org.assertj.core.api.Assertions.assertThat;

import java.util.List;
import java.util.Optional;

import org.junit.jupiter.api.Test;
import org.openmbee.opensysml.DocumentValue;
import org.openmbee.opensysml.Quantity;
import org.openmbee.opensysml.syson.run.DocumentValueText;

class DocumentValueTextTest {
    @Test
    void rendersCellsAsPlainText() {
        DocumentValue.ElementRef element = new DocumentValue.ElementRef("P::a", "PartUsage");
        assertThat(DocumentValueText.render(List.of(element, new DocumentValue.StringValue("mount"),
                new DocumentValue.IntegerValue(15), new DocumentValue.RealValue(8.5),
                new DocumentValue.BooleanValue(true), new DocumentValue.InfinityValue())))
                .isEqualTo("P::a, mount, 15, 8.5, true, infinity");
        assertThat(DocumentValueText.render(new DocumentValue.ObjectRef(3, "", Optional.empty()))).isEqualTo("#3");
        assertThat(DocumentValueText.render(new DocumentValue.ObjectRef(3, "car.wheels[2]", Optional.empty())))
                .isEqualTo("car.wheels[2]");
        assertThat(DocumentValueText.render(new DocumentValue.QuantityValue(new Quantity(2.0, Optional.of("kg"), Optional.empty()))))
                .isEqualTo("2.0 kg");
        assertThat(DocumentValueText.render(new DocumentValue.DocumentVerdict(element, "requirement", "text", "path",
                "holds", "", "", List.of()))).isEqualTo("holds");
        assertThat(DocumentValueText.render(List.of())).isEmpty();
    }

    @Test
    void namesTheDocumentFileAsTheServiceDoes() {
        assertThat(org.openmbee.opensysml.syson.run.RunDocument.fileName("Observatory::'Mass Appendix'", ".html"))
                .isEqualTo("Observatory-Mass.20Appendix.html");
        assertThat(org.openmbee.opensysml.syson.run.RunDocument.fileName("Ünïcode::'a::b'", ".md"))
                .isEqualTo(".C3.9Cn.C3.AFcode-a.3A.3Ab.md".replaceFirst("^", "_"));
        assertThat(org.openmbee.opensysml.syson.run.RunDocument.fileName("CON", ".md")).isEqualTo(".43ON.md".replaceFirst("^", "_"));
        String longName = "A".repeat(300);
        String cut = org.openmbee.opensysml.syson.run.RunDocument.fileName(longName, ".md");
        assertThat(cut).hasSize(255).endsWith(".md").contains("~");
        assertThat(cut.substring(0, cut.indexOf('~'))).isEqualTo("A".repeat(255 - 3 - 17));
    }

    @Test
    void namesTheDocumentFileAfterTheDocument() {
        assertThat(org.openmbee.opensysml.syson.run.RunDocument.of(
                org.openmbee.opensysml.syson.run.RunDocumentForm.MARKDOWN, "", "MassReport").fileName())
                .isEqualTo("MassReport.md");
        assertThat(org.openmbee.opensysml.syson.run.RunDocument.of(
                org.openmbee.opensysml.syson.run.RunDocumentForm.HTML, "", "Reports::'Mass Report'").fileName())
                .isEqualTo("Reports-Mass.20Report.html");
    }
}

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
    void namesTheDocumentFileAfterTheDocument() {
        assertThat(org.openmbee.opensysml.syson.run.RunDocument.of(
                org.openmbee.opensysml.syson.run.RunDocumentForm.MARKDOWN, "", "MassReport").fileName())
                .isEqualTo("MassReport.md");
        assertThat(org.openmbee.opensysml.syson.run.RunDocument.of(
                org.openmbee.opensysml.syson.run.RunDocumentForm.HTML, "", "Reports::'Mass Report'").fileName())
                .isEqualTo("Mass Report.html");
    }
}

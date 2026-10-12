package org.openmbee.opensysml.syson.run;

import java.util.List;
import java.util.stream.Collectors;

import org.openmbee.opensysml.DocumentValue;

public final class DocumentValueText {
    private DocumentValueText() {}

    public static String render(List<DocumentValue> cell) {
        return cell.stream().map(DocumentValueText::render).collect(Collectors.joining(", "));
    }

    public static String render(DocumentValue value) {
        if (value instanceof DocumentValue.ElementRef element) return element.id();
        if (value instanceof DocumentValue.ObjectRef object) {
            return object.path().isEmpty() ? "#" + object.id() : object.path();
        }
        if (value instanceof DocumentValue.StringValue string) return string.value();
        if (value instanceof DocumentValue.IntegerValue integer) return Long.toString(integer.value());
        if (value instanceof DocumentValue.BigIntegerValue integer) return integer.value().toString();
        if (value instanceof DocumentValue.RealValue real) return Double.toString(real.value());
        if (value instanceof DocumentValue.RationalValue rational) return rational.value().toString();
        if (value instanceof DocumentValue.BooleanValue bool) return Boolean.toString(bool.value());
        if (value instanceof DocumentValue.InfinityValue) return "infinity";
        if (value instanceof DocumentValue.QuantityValue quantity) {
            return quantity.quantity().magnitude() + quantity.quantity().unit().map(unit -> " " + unit).orElse("");
        }
        if (value instanceof DocumentValue.DocumentVerdict verdict) return verdict.status();
        if (value instanceof DocumentValue.DocumentState state) return state.path();
        if (value instanceof DocumentValue.DocumentEvent event) return event.text();
        return value.toString();
    }
}

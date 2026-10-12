package org.openmbee.opensysml.syson.run;

public record RunDocument(RunDocumentForm form, String content, String fileName) {
    // Names the file after the document's own name, unquoted, with the form's extension.
    public static RunDocument of(RunDocumentForm form, String content, String target) {
        int separator = target.lastIndexOf("::");
        String simpleName = (separator < 0 ? target : target.substring(separator + 2)).replace("'", "");
        return new RunDocument(form, content, simpleName + "." + form.extension());
    }
}

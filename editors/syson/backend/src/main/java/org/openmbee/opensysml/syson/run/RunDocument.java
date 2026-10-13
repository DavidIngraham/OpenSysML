package org.openmbee.opensysml.syson.run;

import org.openmbee.opensysml.RenderedDocument;

public record RunDocument(RunDocumentForm form, String content, String fileName) {

    /**
     * The rendered document under the file name the service plans for it, or, against a service that
     * does not name it, under the document's own name.
     */
    public static RunDocument of(RunDocumentForm form, RenderedDocument rendered, String target) {
        String fileName = rendered.fileName().isEmpty() ? fallbackFileName(target, form) : rendered.fileName();
        return new RunDocument(form, rendered.content(), fileName);
    }

    static String fallbackFileName(String target, RunDocumentForm form) {
        String simple = target.substring(target.lastIndexOf("::") + 2).replace("'", "");
        return simple.replaceAll("[^\\p{L}\\p{N}._ -]", "_") + "." + form.extension();
    }
}

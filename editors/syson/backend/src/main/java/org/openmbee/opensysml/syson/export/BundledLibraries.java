package org.openmbee.opensysml.syson.export;

// Answers whether the OpenSysML service already provides a library package of this name.
@FunctionalInterface
public interface BundledLibraries {
    BundledLibraries NONE = name -> false;

    boolean bundles(String libraryName);
}

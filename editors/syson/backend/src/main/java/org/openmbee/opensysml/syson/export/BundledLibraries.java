package org.openmbee.opensysml.syson.export;

import java.util.Optional;
import java.util.Set;

// Answers whether the OpenSysML service already provides a library package of this name,
// and with which direct members.
@FunctionalInterface
public interface BundledLibraries {
    BundledLibraries NONE = name -> Optional.empty();

    Optional<Set<String>> members(String libraryName);
}

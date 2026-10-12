package org.openmbee.opensysml.syson.export;

import java.util.LinkedHashSet;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;

import org.openmbee.opensysml.Connection;
import org.openmbee.opensysml.Symbol;

// A library package is bundled when an empty model already resolves it: the service's libraries
// are the only names an empty model has. The empty model is parsed afresh for each new name, so
// the service evicting it from its cache costs nothing; answers are kept for the connection's lifetime.
public final class ServiceLibraries implements BundledLibraries {
    private final Connection connection;
    private final Map<String, Optional<Set<String>>> known = new ConcurrentHashMap<>();

    public ServiceLibraries(Connection connection) {
        this.connection = connection;
    }

    @Override
    public Optional<Set<String>> members(String libraryName) {
        if (libraryName == null || libraryName.isBlank()) return Optional.empty();
        return known.computeIfAbsent(libraryName, name -> connection.parse("").find(name).map(ServiceLibraries::memberNames));
    }

    // The simple names of a symbol's direct members, as its child ids end.
    private static Set<String> memberNames(Symbol symbol) {
        Set<String> names = new LinkedHashSet<>();
        for (String id : symbol.childIds()) {
            int separator = id.lastIndexOf("::");
            names.add(separator < 0 ? id : id.substring(separator + 2));
        }
        return names;
    }
}

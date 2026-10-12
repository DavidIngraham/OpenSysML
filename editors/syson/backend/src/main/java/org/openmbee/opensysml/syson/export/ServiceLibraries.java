package org.openmbee.opensysml.syson.export;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

import org.openmbee.opensysml.Connection;
import org.openmbee.opensysml.Model;

// A library package is bundled when an empty model already resolves it: the service's libraries
// are the only names an empty model has. Answers are cached for the connection's lifetime.
public final class ServiceLibraries implements BundledLibraries {
    private final Connection connection;
    private final Map<String, Boolean> known = new ConcurrentHashMap<>();
    private volatile Model empty;

    public ServiceLibraries(Connection connection) {
        this.connection = connection;
    }

    @Override
    public boolean bundles(String libraryName) {
        if (libraryName == null || libraryName.isBlank()) return false;
        return known.computeIfAbsent(libraryName, name -> emptyModel().find(name).isPresent());
    }

    private Model emptyModel() {
        Model model = empty;
        if (model == null) {
            model = connection.parse("");
            empty = model;
        }
        return model;
    }
}

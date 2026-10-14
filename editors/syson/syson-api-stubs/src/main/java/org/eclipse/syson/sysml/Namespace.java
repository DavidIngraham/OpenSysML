package org.eclipse.syson.sysml;

import org.eclipse.emf.common.util.EList;

public interface Namespace extends Element {
    EList<Relationship> getOwnedRelationship();
}

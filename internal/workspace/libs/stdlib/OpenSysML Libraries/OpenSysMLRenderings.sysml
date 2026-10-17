library package OpenSysMLRenderings {
	doc /*
	 * NON-NORMATIVE rendering vocabulary for case, mixed and qualitative fault-tree diagrams.
	 * Models using it are valid SysML v2 and depend on this OpenSysML library.
	*/

	private import Views::*;

	rendering asCaseDiagram : GraphicalRendering[1] :> renderings;
	rendering asMixedDiagram : GraphicalRendering[1] :> renderings;

    rendering asFaultTree : GraphicalRendering[1] :> renderings;
    view def FaultTreeView { render asFaultTree; }

	view def CaseView {
		render asCaseDiagram;
	}

	view def MixedView {
		render asMixedDiagram;
	}
}

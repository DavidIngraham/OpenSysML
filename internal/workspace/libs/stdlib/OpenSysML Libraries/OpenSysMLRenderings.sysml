library package OpenSysMLRenderings {
	doc /*
	 * NON-NORMATIVE rendering vocabulary for case and mixed diagrams.
	 * Models using it are valid SysML v2 and depend on this OpenSysML library.
	*/

	private import Views::*;

	rendering asCaseDiagram : GraphicalRendering[1] :> renderings;
	rendering asMixedDiagram : GraphicalRendering[1] :> renderings;

	view def CaseView {
		render asCaseDiagram;
	}

	view def MixedView {
		render asMixedDiagram;
	}

	/*
	 * A view's depth bounds how many levels of nested elements its rendering
	 * draws below each exposed element: 0, or no depth, draws every level. A
	 * view redefines it (`:>> depth = 2;`) or simply declares `attribute depth = 2;`.
	 */
	view def DepthLimitedView {
		attribute depth : ScalarValues::Natural[0..1];
	}
}

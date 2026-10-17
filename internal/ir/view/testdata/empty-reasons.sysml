package Cart {
	part def Wheel;
	part def Axle;
	part wheel : Wheel;
}

package Hollow;

package EmptyReasons {
	private import Views::*;
	private import SysML::Systems::*;

	// Every element exposed is a part; the view's own filter admits only actions.
	view filteredOut {
		expose Cart::*;
		filter @ActionUsage;
	}

	// The filter is on the expose, and inherited from the view definition.
	view def ActionsOnly {
		expose Cart::**[@ActionUsage];
	}
	view inheritedFilter : ActionsOnly;

	// The expose names a namespace that does not exist.
	view unresolved {
		expose Nowhere::*;
	}

	// The expose names a namespace with no member.
	view hollow {
		expose Hollow::*;
	}

	// A filter with nothing to reject is not the reason.
	view hollowFiltered {
		expose Hollow::*;
		filter @ActionUsage;
	}
}

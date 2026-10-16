metadata def Hidden;

package FilteredWalk {
	private import Views::*;

	port def Signal;

	part rover {
		attribute mass;
		part chassis {
			attribute length;
			part axle;
		}
		#Hidden part telemetry {
			port tx : Signal;
		}
		part radio {
			port rx : Signal;
		}
		connect radio.rx to telemetry.tx;
	}

	// The filters admit the root and prune the walk: no attribute, nothing
	// marked Hidden, is drawn beneath it.
	view structure {
		expose rover;
		filter not @Hidden;
		filter @SysML::PartUsage or @SysML::PortUsage;
		render asTreeDiagram;
	}

	// The connector to telemetry.tx ends at a pruned element: it is reported, not drawn.
	view links {
		expose rover;
		filter not @Hidden;
		render asInterconnectionDiagram;
	}
}

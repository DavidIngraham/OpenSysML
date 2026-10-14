// Ports a Layout places as the source tool drew them: a rectangle sized to carry the port's
// own name, over the border of the part. The rendering keeps the rectangle's position alone and
// draws the notation's port square there.
package LayoutPorts {
	private import Views::*;
	private import DiagramLayout::*;

	port def Signal;
	part def Detector { port tx : Signal { attribute level; } }
	part def Amplifier { port rx : ~Signal; }
	part def Chain {
		part detector : Detector;
		part amplifier : Amplifier;
		connection link connect detector.tx to amplifier.rx;
	}

	// The ports the parts have from their definitions, exposed beside the parts, each drawn
	// over the border between the two parts.
	view typed {
		expose Detector::tx;
		expose Amplifier::rx;
		expose Chain::detector;
		expose Chain::amplifier;
		expose Chain::link;
		render asInterconnectionDiagram;
		metadata Layout about Chain::detector { x = 40; y = 20; width = 400; height = 160; }
		metadata Layout about Chain::amplifier { x = 40; y = 240; width = 400; height = 160; }
		metadata Layout about Detector::tx { x = 200; y = 165; width = 130; height = 36; }
		metadata Layout about Amplifier::rx { x = 200; y = 219; width = 130; height = 36; }
		metadata Route about Chain::link { points = (265, 201, 265, 219); }
	}

	part def Bench {
		part supply { port outlet; }
		part instrument {
			port inlet;
			part sensor { port probe; }
		}
		connection feed connect supply.outlet to instrument.inlet;
	}

	// Declared ports, drawn as members of their parts: the supply's over its right border, the
	// instrument's over its left border, and the sensor's inside the sensor near its bottom.
	view declared {
		expose Bench::supply;
		expose Bench::instrument;
		expose Bench::feed;
		render asInterconnectionDiagram;
		metadata Layout about Bench::supply { x = 40; y = 20; width = 200; height = 100; }
		metadata Layout about Bench::supply::outlet { x = 175; y = 60; width = 130; height = 36; }
		metadata Layout about Bench::instrument { x = 400; y = 20; width = 300; height = 200; }
		metadata Layout about Bench::instrument::inlet { x = 335; y = 42; width = 130; height = 36; }
		metadata Layout about Bench::instrument::sensor { x = 450; y = 80; width = 200; height = 100; }
		metadata Layout about Bench::instrument::sensor::probe { x = 500; y = 150; width = 130; height = 36; }
		metadata Route about Bench::feed { points = (305, 78, 335, 60); }
	}

	// A definition's port in a tree rendering, drawn over the definition's right border; the
	// attribute nested in the port has no place and is left undrawn.
	view tree {
		expose Detector;
		render asTreeDiagram;
		metadata Layout about Detector { x = 40; y = 20; width = 200; height = 100; }
		metadata Layout about Detector::tx { x = 175; y = 100; width = 130; height = 36; }
	}
}

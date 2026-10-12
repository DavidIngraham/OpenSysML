// Use cases whose actor parameters are boundary symbols of a positioned node:
// two on one side meet at one route end, two meet at a corner from different
// sides, two meet on a node with no stated size; a part's nested parts, no
// boundary symbols, meet inside it; and a migration made up one connection's
// name.
package BoundaryParameters {
    part Operator;
    part Administrator;
    part 'Service Technician';
    part Rack;

    use case 'Load Sample' {
        subject;
        actor operator :> Operator;
        actor technician :> 'Service Technician';
        actor administrator :> Administrator;
    }
    part Instrument {
        part detector;
        part filament;
    }
    connection 'Operator to Load Sample' connect Operator to 'Load Sample'.operator;
    connection 'Service Technician to Load Sample' connect 'Service Technician' to 'Load Sample'.technician;
    connection 'Administrator to Load Sample' connect Administrator to 'Load Sample'.administrator;
    connection 'Rack to Detector' connect Rack to Instrument.detector;
    connection 'Rack to Filament' connect Rack to Instrument.filament;

    use case 'Acquire Image' {
        subject;
        actor operator :> Operator;
    }
    use case 'Configure Instrument' {
        subject;
        actor operator :> Operator;
        actor administrator :> Administrator;
    }

    connection 'Operator to Acquire Image' connect Operator to 'Acquire Image'.operator;
    connection 'Operator to Configure Instrument' connect Operator to 'Configure Instrument'.operator;
    connection 'Administrator to Configure Instrument' connect Administrator to 'Configure Instrument'.administrator;

    // The actors left of the use cases, their routes ending on the left border.
    view lab {
        expose Operator;
        expose Administrator;
        expose 'Acquire Image';
        expose 'Configure Instrument';
        expose 'Operator to Acquire Image';
        expose 'Operator to Configure Instrument';
        expose 'Administrator to Configure Instrument';
        render Views::asInterconnectionDiagram;

        metadata DiagramLayout::Layout about Operator { x = 40; y = 60; width = 80; height = 44; }
        metadata DiagramLayout::Layout about Administrator { x = 40; y = 220; width = 80; height = 44; }
        metadata DiagramLayout::Layout about 'Acquire Image' { x = 320; y = 20; width = 160; height = 50; }
        metadata DiagramLayout::Layout about 'Configure Instrument' { x = 320; y = 120; width = 160; height = 50; }
        metadata DiagramLayout::Route about 'Operator to Acquire Image' { points = (120, 82, 320, 45); }
        metadata DiagramLayout::Route about 'Operator to Configure Instrument' { points = (120, 82, 320, 145); }
        metadata DiagramLayout::Route about 'Administrator to Configure Instrument' { points = (120, 242, 320, 145); }
    }

    // The same, mirrored: the actors right of the use cases, the routes ending
    // on the right border, the administrator's coming from above.
    view mirroredLab {
        expose Operator;
        expose Administrator;
        expose 'Configure Instrument';
        expose 'Operator to Configure Instrument';
        expose 'Administrator to Configure Instrument';
        render Views::asInterconnectionDiagram;

        metadata DiagramLayout::Layout about Operator { x = 600; y = 220; width = 80; height = 44; }
        metadata DiagramLayout::Layout about Administrator { x = 600; y = 60; width = 80; height = 44; }
        metadata DiagramLayout::Layout about 'Configure Instrument' { x = 320; y = 120; width = 160; height = 50; }
        metadata DiagramLayout::Route about 'Operator to Configure Instrument' { points = (600, 242, 480, 145); }
        metadata DiagramLayout::Route about 'Administrator to Configure Instrument' { points = (600, 82, 480, 145); }
    }

    // Two routes end at the use case's top-left corner, one from the left and
    // one from above.
    view corner {
        expose Operator;
        expose 'Service Technician';
        expose 'Load Sample';
        expose 'Operator to Load Sample';
        expose 'Service Technician to Load Sample';
        render Views::asInterconnectionDiagram;

        metadata DiagramLayout::Layout about Operator { x = 40; y = 98; width = 80; height = 44; }
        metadata DiagramLayout::Layout about 'Service Technician' { x = 280; y = 20; width = 80; height = 44; }
        metadata DiagramLayout::Layout about 'Load Sample' { x = 320; y = 120; width = 160; height = 50; }
        metadata DiagramLayout::Route about 'Operator to Load Sample' { points = (120, 120, 320, 120); }
        metadata DiagramLayout::Route about 'Service Technician to Load Sample' { points = (320, 64, 320, 120); }
    }

    // Two routes from above and one from the left meet near the top-left
    // corner; the two symbols from above are together wider than the top side
    // leaves beyond the one from the left.
    view crowdedCorner {
        expose Operator;
        expose Administrator;
        expose 'Service Technician';
        expose 'Load Sample';
        expose 'Operator to Load Sample';
        expose 'Service Technician to Load Sample';
        expose 'Administrator to Load Sample';
        render Views::asInterconnectionDiagram;

        metadata DiagramLayout::Layout about Operator { x = 40; y = 98; width = 80; height = 44; }
        metadata DiagramLayout::Layout about 'Service Technician' { x = 280; y = 20; width = 80; height = 44; }
        metadata DiagramLayout::Layout about Administrator { x = 360; y = 20; width = 80; height = 44; }
        metadata DiagramLayout::Layout about 'Load Sample' { x = 320; y = 120; width = 160; height = 50; }
        metadata DiagramLayout::Route about 'Operator to Load Sample' { points = (120, 120, 320, 120); }
        metadata DiagramLayout::Route about 'Service Technician to Load Sample' { points = (320, 64, 320, 120); }
        metadata DiagramLayout::Route about 'Administrator to Load Sample' { points = (400, 64, 400, 120); }
    }

    // The use case has a position but no size, so its box is fitted round its
    // parameters.
    view unsized {
        expose Operator;
        expose Administrator;
        expose 'Configure Instrument';
        expose 'Operator to Configure Instrument';
        expose 'Administrator to Configure Instrument';
        render Views::asInterconnectionDiagram;

        metadata DiagramLayout::Layout about Operator { x = 40; y = 60; width = 80; height = 44; }
        metadata DiagramLayout::Layout about Administrator { x = 40; y = 220; width = 80; height = 44; }
        metadata DiagramLayout::Layout about 'Configure Instrument' { x = 320; y = 120; }
        metadata DiagramLayout::Route about 'Operator to Configure Instrument' { points = (120, 82, 320, 145); }
        metadata DiagramLayout::Route about 'Administrator to Configure Instrument' { points = (120, 242, 320, 145); }
    }

    // The routes to the instrument's nested parts end at one point inside it.
    view interior {
        expose Rack;
        expose Instrument;
        expose 'Rack to Detector';
        expose 'Rack to Filament';
        render Views::asInterconnectionDiagram;

        metadata DiagramLayout::Layout about Rack { x = 40; y = 60; width = 80; height = 44; }
        metadata DiagramLayout::Layout about Instrument { x = 320; y = 120; width = 160; height = 100; }
        metadata DiagramLayout::Route about 'Rack to Detector' { points = (120, 82, 370, 170); }
        metadata DiagramLayout::Route about 'Rack to Filament' { points = (120, 82, 370, 170); }
    }

    metadata MigrationMetadata::SynthesizedName about 'Operator to Acquire Image';
}

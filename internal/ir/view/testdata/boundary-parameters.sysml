// Use cases whose actor parameters are boundary symbols of a positioned node:
// two on one side meet at one route end, and a migration made up one
// connection's name.
package BoundaryParameters {
    part Operator;
    part Administrator;

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

    metadata MigrationMetadata::SynthesizedName about 'Operator to Acquire Image';
}

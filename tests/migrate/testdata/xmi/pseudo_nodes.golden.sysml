package Model {
    package Imaging {
        occurrence def Camera {
            action def Acquire {
                view Acquire : StandardViewDefinitions::ActionFlowView {
                    expose Imaging::Camera::Acquire;
                    @DiagramLayout::Canvas { unit = "px"; width = 500; height = 360; }
                    metadata DiagramLayout::Layout about Imaging::Camera::Acquire::start { x = 85; y = 90; width = 15; height = 15; }
                    metadata DiagramLayout::Layout about position { x = 21; y = 117; width = 142; height = 33; }
                    metadata DiagramLayout::Layout about final { x = 82; y = 200; width = 20; height = 20; }
                    metadata DiagramLayout::Layout about 'Position Not Correct' { x = 300; y = 100; width = 180; height = 160; }
                    metadata DiagramLayout::Layout about 'Position Not Correct'::start { x = 379; y = 127; width = 15; height = 15; }
                    metadata DiagramLayout::Layout about Imaging::Camera::Acquire::'Position Not Correct'::adjust { x = 320; y = 152; width = 142; height = 33; }
                    metadata DiagramLayout::Layout about Imaging::Camera::Acquire::'Position Not Correct'::final { x = 381; y = 220; width = 20; height = 20; }
                    metadata DiagramLayout::Route about 'start to position' { points = (92, 105, 92, 117); }
                    metadata DiagramLayout::Route about 'position to Position Not Correct' { points = (92, 150, 92, 158, 300, 158); }
                    metadata DiagramLayout::Route about 'Position Not Correct to final' { points = (300, 210, 102, 210); }
                    metadata DiagramLayout::Route about Imaging::Camera::Acquire::'Position Not Correct'::'start to adjust' { points = (386, 142, 386, 152); }
                    metadata DiagramLayout::Route about Imaging::Camera::Acquire::'Position Not Correct'::'adjust to final' { points = (391, 185, 391, 220); }
                    render Views::asInterconnectionDiagram;
                }
                succession 'start to position' first start then position;
                action position;
                succession 'position to Position Not Correct' first position then 'Position Not Correct';
                action final terminate;
                action 'Position Not Correct' {
                    succession 'start to adjust' first start then adjust;
                    action adjust;
                    succession 'adjust to final' first adjust then final;
                    action final terminate;
                    metadata MigrationMetadata::SynthesizedName about 'start to adjust', 'adjust to final', final;
                }
                succession 'Position Not Correct to final' first 'Position Not Correct' then final;
                metadata MigrationMetadata::SynthesizedName about 'start to position', 'position to Position Not Correct', final, 'Position Not Correct to final';
            }
            state def Modes {
                view Modes : StandardViewDefinitions::StateTransitionView {
                    expose Imaging::Camera::Modes;
                    @DiagramLayout::Canvas { unit = "px"; width = 140; height = 220; }
                    metadata DiagramLayout::Layout about Imaging::Camera::Modes::start { x = 50; y = 50; width = 20; height = 20; }
                    metadata DiagramLayout::Layout about Imaging::Camera::Modes::Idle { x = 10; y = 100; width = 100; height = 40; }
                    metadata DiagramLayout::Layout about Imaging::Camera::Modes::done { x = 50; y = 190; width = 20; height = 20; }
                    metadata DiagramLayout::Route about 'start then Idle' { points = (60, 70, 60, 100); }
                    metadata DiagramLayout::Route about 'Idle then done' { points = (60, 140, 60, 190); }
                    render Views::asInterconnectionDiagram;
                }
                transition 'start then Idle' first start then Idle;
                state Idle;
                transition 'Idle then done' first Idle then done;
                metadata MigrationMetadata::SynthesizedName about 'start then Idle', 'Idle then done';
            }
        }
    }
}

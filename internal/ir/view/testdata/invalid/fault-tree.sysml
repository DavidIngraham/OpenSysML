package InvalidTree {
    occurrence a : FailureAnalysis::Event;
    occurrence b : FailureAnalysis::Event;
    connection ab : CauseAndEffect::Causation connect a to b;
    connection ba : CauseAndEffect::Causation connect b to a;
    view tree : OpenSysMLRenderings::FaultTreeView { expose InvalidTree::*; }
}

package ExtraCause {
    occurrence a : FailureAnalysis::BasicEvent;
    occurrence b : FailureAnalysis::BasicEvent;
    occurrence loss : FailureAnalysis::OrGate { ref occurrence :>> inputs = a; }
    connection bl : CauseAndEffect::Causation connect b to loss;
    view tree : FailureAnalysisViews::FaultTreeView { expose ExtraCause::*; }
}

package Inherited {
    part def Assembly { occurrence failure : FailureAnalysis::BasicEvent; }
    part assembly : Assembly;
    view tree : FailureAnalysisViews::FaultTreeView { expose assembly; }
}

package Ambiguous {
    occurrence def Both :> FailureAnalysis::AndGate, FailureAnalysis::OrGate;
    occurrence both : Both;
    view tree : FailureAnalysisViews::FaultTreeView { expose both; }
}

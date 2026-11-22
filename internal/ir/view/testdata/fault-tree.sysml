package PowerSupply {
    private import FailureAnalysis::*;
    private import CauseAndEffect::*;
    occurrence def SupplyFault :> BasicEvent;
    occurrence def EitherFailure :> OrGate;
    occurrence a : SupplyFault;
    occurrence b : SupplyFault;
    occurrence common : BasicEvent;
    occurrence left : EitherFailure { ref occurrence :>> inputs = (a,common); }
    occurrence right : OrGate { ref occurrence :>> inputs = (b,common); }
    occurrence loss : AndGate { ref occurrence :>> inputs = (left,right); }
    connection al : Causation connect a to left;
    connection cl : Causation connect common to left;
    connection br : Causation connect b to right;
    connection cr : Causation connect common to right;
    connection lt : Causation connect left to loss;
    connection rt : Causation connect right to loss;
    view def LocalView :> OpenSysMLRenderings::FaultTreeView;
    view tree : LocalView { expose PowerSupply::*; }
    view partial : LocalView { expose a; expose al; }
}


package PartialGate {
    occurrence a : FailureAnalysis::BasicEvent;
    occurrence b : FailureAnalysis::BasicEvent;
    occurrence <'Loss of supply'> loss : FailureAnalysis::AndGate { ref occurrence :>> inputs = (a,b); }
    connection al : CauseAndEffect::Causation connect a to loss;
    view tree : FailureAnalysisViews::FaultTreeView { expose PartialGate::*; }
}

package Nested {
    occurrence assembly : FailureAnalysis::Event {
        occurrence cause : FailureAnalysis::BasicEvent;
        occurrence effect : FailureAnalysis::Event;
        connection ce : CauseAndEffect::Causation connect cause to effect;
    }
    view tree : FailureAnalysisViews::FaultTreeView { expose assembly; }
}



package Reordered {
    occurrence a : FailureAnalysis::BasicEvent;
    occurrence b : FailureAnalysis::Event;
    connection ab : CauseAndEffect::Causation connect (effect :>> theEffect ::> b, cause :>> theCause ::> a);
    view tree : FailureAnalysisViews::FaultTreeView { expose Reordered::*; }
}

package ReferenceCause {
    occurrence a : FailureAnalysis::BasicEvent;
    ref occurrence shared : FailureAnalysis::BasicEvent = a;
    occurrence loss : FailureAnalysis::OrGate { ref occurrence :>> inputs = shared; }
    connection al : CauseAndEffect::Causation connect shared to loss;
    view tree : FailureAnalysisViews::FaultTreeView { expose shared; expose a; expose loss; expose al; }
}

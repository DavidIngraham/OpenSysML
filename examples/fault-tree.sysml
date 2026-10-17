package PowerSupply {
    private import FailureAnalysis::*;
    private import CauseAndEffect::*;
    occurrence def SupplyFault :> BasicEvent;
    occurrence def EitherFailure :> OrGate;
    occurrence a : SupplyFault;
    occurrence b : SupplyFault;
    occurrence common : BasicEvent { attribute :>> description = "Common supply fault"; }
    occurrence left : EitherFailure { ref occurrence :>> inputs = (a,common); }
    occurrence right : OrGate { ref occurrence :>> inputs = (b,common); }
    occurrence loss : AndGate { ref occurrence :>> inputs = (left,right); attribute :>> description = "Both supplies unavailable"; }
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
package PowerReport {
    private import DocumentQueries::*;
    part def Report :> Document {
        attribute :>> title = "Redundant power supply: qualitative fault tree";
        part tree : Diagram { ref :>> source = PowerSupply::tree; }
    }
}

package PowerReview {
    part supply : Parts::Part;
    part review : FailureAnalysis::Review {
        ref part :>> system = supply;
        attribute :>> scopeReviewed = false;
        part outage : FailureAnalysis::ReviewRow subsets modes {
            attribute :>> identifier = "P-01";
            attribute :>> intendedFunction = "Maintain redundant power";
            ref part :>> affected = supply;
            ref occurrence :>> cause = PowerSupply::common;
            ref occurrence :>> failureMode = PowerSupply::loss;
            ref occurrence :>> localEffect = PowerSupply::loss;
            ref occurrence :>> missionEffect = PowerSupply::loss;
            attribute :>> proposedDetection = "Bus monitoring";
            attribute :>> recommendedAction = "Test isolation against a common supply fault";
            attribute :>> ownerRole = "Power";
            attribute :>> safetyCritical = true;
        }
    }
    calc def Rows :> FailureAnalysisViews::ReviewRows {
        in :>> reviewName = "PowerReview::review";
    }
}

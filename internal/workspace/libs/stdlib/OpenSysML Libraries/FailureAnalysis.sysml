library package FailureAnalysis {
    doc /* Non-normative, RAAML-aligned qualitative failure analysis subset.
     * Event/gate vocabulary follows RAAML 1.1 FTA; this is not the UML profile
     * or a claim of RAAML conformance. No probabilities or independence inferred.
     */
    private import ScalarValues::*;
    occurrence def Event {
        attribute description : String;
        attribute present : Boolean;
    }
    occurrence def BasicEvent :> Event;
    occurrence def UndevelopedEvent :> Event;
    occurrence def HouseEvent :> Event;
    abstract occurrence def Gate :> Event {
        ref occurrence inputs : Event [1..*];
    }
    occurrence def OrGate :> Gate {
        attribute :>> present = inputs->ControlFunctions::exists { in x : Event; x.present };
    }
    occurrence def AndGate :> Gate {
        attribute :>> present = SequenceFunctions::size(inputs) > 0 and
            inputs->ControlFunctions::forAll { in x : Event; x.present };
    }
    analysis def Evaluate { subject proposition : Event; out attribute present : Boolean = proposition.present; }
    part def ReviewRow {
        attribute identifier : String;
        ref part affected : Parts::Part;
        attribute intendedFunction : String;
        ref occurrence failureMode : Event;
        ref occurrence cause : Event;
        ref occurrence localEffect : Event;
        ref occurrence missionEffect : Event;
        attribute failureModeText : String = failureMode.description;
        attribute causeText : String = cause.description;
        attribute localEffectText : String = localEffect.description;
        attribute missionEffectText : String = missionEffect.description;
        attribute proposedDetection : String;
        attribute recommendedAction : String;
        attribute ownerRole : String;
        attribute safetyCritical : Boolean;
        attribute dispositionAccepted : Boolean default false;
        attribute actionEvidenceAccepted : Boolean default false;
        attribute evidenceReference : String default "Not yet available";
    }
    part def Review {
        ref part system : Parts::Part;
        part modes : ReviewRow [1..*];
        attribute scopeReviewed : Boolean;
    }
    analysis def DesignFailureReview {
        subject review : Review;
        out attribute modeCount : Integer = SequenceFunctions::size(review.modes);
        out attribute safetyCriticalCount : Integer = SequenceFunctions::size(review.modes->ControlFunctions::select { in mode : ReviewRow; mode.safetyCritical });
        out attribute unresolvedCriticalCount : Integer = SequenceFunctions::size(review.modes->ControlFunctions::select {
            in mode : ReviewRow; mode.safetyCritical and not (mode.dispositionAccepted and mode.actionEvidenceAccepted and mode.evidenceReference != "" and mode.evidenceReference != "Not yet available") });
        out attribute openActionCount : Integer = SequenceFunctions::size(review.modes->ControlFunctions::select { in mode : ReviewRow; not mode.actionEvidenceAccepted });
        return reviewReady : Boolean = modeCount > 0 and review.scopeReviewed and unresolvedCriticalCount == 0;
        objective inventory { require constraint { modeCount > 0 } }
    }
}

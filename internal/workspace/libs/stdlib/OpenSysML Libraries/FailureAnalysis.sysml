library package FailureAnalysis {
    doc /* NON-NORMATIVE, RAAML-aligned qualitative failure analysis subset.
     * Event/gate vocabulary follows RAAML 1.1 FTA; this is not the UML profile
     * or a claim of RAAML conformance. No probabilities or independence inferred.
     */
    private import ScalarValues::*;
    occurrence def Event {
        doc /* A proposition about an explicitly chosen exposure interval; present is unknown until bound. */
        attribute description : String;
        attribute present : Boolean;
    }
    occurrence def BasicEvent :> Event { doc /* An event treated as a leaf within this analysis scope, without an assigned probability. */ }
    occurrence def UndevelopedEvent :> Event { doc /* An event whose contributing causes remain to be developed. */ }
    occurrence def HouseEvent :> Event { doc /* A scenario condition supplied as an explicit Boolean, not an inferred failure rate. */ }
    abstract occurrence def Gate :> Event {
        doc /* A proposition computed from referenced input events. */
        ref occurrence inputs : Event [1..*];
    }
    occurrence def OrGate :> Gate {
        doc /* True when at least one input proposition holds. */
        attribute :>> present = inputs->ControlFunctions::exists { in x : Event; x.present };
    }
    occurrence def AndGate :> Gate {
        doc /* True when every input holds in the same interval; no temporal ordering or independence is implied. */
        attribute :>> present = SequenceFunctions::size(inputs) > 0 and
            inputs->ControlFunctions::forAll { in x : Event; x.present };
    }
    analysis def Evaluate { doc /* Read an event proposition through the native analysis interface. */ subject proposition : Event; out attribute present : Boolean = proposition.present; }
    part def ReviewRow {
        doc /* A failure review and action record referencing events; not itself a failure occurrence. */
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
        doc /* An explicitly scoped review of a system, with one or more owned failure rows. */
        ref part system : Parts::Part;
        part modes : ReviewRow [1..*];
        attribute scopeReviewed : Boolean;
    }
    analysis def DesignFailureReview {
        doc /* Inventory and disposition checks only; declared evidence references do not establish physical adequacy. */
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

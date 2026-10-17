package test {
    part review : FailureAnalysis::Review {
        attribute :>> scopeReviewed = true;
        part row : FailureAnalysis::ReviewRow subsets modes {
            attribute :>> safetyCritical = true;
            attribute :>> dispositionAccepted = true;
            attribute :>> actionEvidenceAccepted = true;
            attribute :>> evidenceReference = "";
        }
    }
    analysis check : FailureAnalysis::DesignFailureReview { subject :>> review = test::review; }
}

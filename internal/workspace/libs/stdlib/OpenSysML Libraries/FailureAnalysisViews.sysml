library package FailureAnalysisViews {
    doc /* NON-NORMATIVE views for the RAAML-aligned FailureAnalysis subset.
     * Supply review scope as a qualified name; rows are selected by semantic type.
     * The fault-tree rendering uses native types and standard Causation connections.
     */
    private import DocumentQueries::*;
    private import ScalarValues::*;
    view def FaultTreeView :> OpenSysMLRenderings::FaultTreeView { doc /* A qualitative causal view with gate symbols; rendering requires a supporting engine. */ }
    calc def ReviewRows :> Query {
        doc /* Select owned descendant review rows by semantic type and project their referenced descriptions and actions. */
        in reviewName : String;
        Project(source = WhereType(source = Descendants(source = Named(qualifiedName = reviewName)),
            type = "FailureAnalysis::ReviewRow"), columns = (
            Column(name = "ID", path = "identifier"),
            Column(name = "Function / scope", path = "intendedFunction"),
            Column(name = "Cause", path = "causeText"),
            Column(name = "Failure", path = "failureModeText"),
            Column(name = "Action needed", path = "recommendedAction")))
    }
}

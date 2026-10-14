- The SysML v1 migrator places the symbols of an initial node, a flow final node, an initial
  pseudostate and a final state, which it writes as the inherited `start` and `done`, with a
  `DiagramLayout::Layout` named through the body they are drawn in (`about Run::start`,
  `about 'Position Not Correct'::start`), where it used to drop them as not exposed; and it
  writes the `Route` of a control flow, object flow or transition from the edge's source to its
  target however the tool drew the line, orienting it by the symbols the stream names at its
  ends or, in an export naming none, by the boxes the record places. A diagram whose tool drew
  an initial node's flow from the action back to the node no longer renders the start marker
  on top of that action.

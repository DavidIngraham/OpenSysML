- The SysML v1 migrator now reads the symbols Cameo lists under a part shape's `parts` and a
  state shape's `regions` — the nested parts, constraint usages, connectors and comments drawn
  inside a part on an internal block or parametric diagram — as symbols of their own, parented
  to the enclosing shape. They were dropped before, which re-parented the ports and connector
  ends drawn inside them to the outer part and recorded the dropped connectors' ends on it, so
  the outer part counted as a line and lost its placement too. Such diagrams now place the
  nested shapes inside their part, route the connectors between them and keep the outer part's
  `Layout`; the migration report counts them as positioned instead of lying on connections.

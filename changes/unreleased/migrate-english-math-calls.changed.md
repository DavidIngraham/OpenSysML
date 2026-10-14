- The SysML v1 migrator translates a call of a mathematical function by its bare name in an
  English constraint or guard body (`sine_wave = sin(t)` is `sine_wave == TrigFunctions::sin(t)`):
  `sin`, `cos`, `tan`, `cot`, `arcsin`/`asin`, `arccos`/`acos`, `arctan`/`atan` through
  `TrigFunctions`; `sqrt`, `abs`, `min`, `max`, `floor`, `ceil`/`ceiling`, `round` and `pow`
  as the script dialects write them; `exp`, `ln`, `log` and `atan2` through
  `OpenSysMLMathFunctions`, refused in a strict migration. Each takes the arguments the library
  function takes and yields its result type. In an English or a script body alike, a `Complex`
  argument is refused where the parameter is a `Real` and from `min` and `max`, which order
  their arguments; `abs` of a `Complex` is `ComplexFunctions::abs`, and `abs`, `min` and `max`
  of a `Number` go through `NumericalFunctions` rather than `RealFunctions`. A dotted name, a name outside the table, a name of
  several words before the `(` or the wrong number of arguments stays refused, and the reason
  now lists the names an English body does read.

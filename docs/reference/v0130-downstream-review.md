# Downstream compatibility on v0.13.0

The compatibility stack is rebased onto official OpenSysML v0.13.0 commit
`df402e904b2d227aef4718b908adcf3630241706`. The previous tested compiler pin,
`5578dbc8b56b4f892d9cbf3fd137a1a0cb52cf9c`, remains a rollback reference.
The extension-library pin remains
`38097bbc0198767590c58f89796a436678319924`, based on official
`3faa2bf1ddda4f35c3ea1f0e2600b55a10f0e0f1`.

## Replay and source behavior

All production compatibility changes are retained. The old baseline-recording
commit was replaced by new measurements against v0.13.0; its 20-library install
coverage is retained. The General View conflict preserves upstream's new
requirement-party expansion for unbounded views and keeps bounded views within
their declared context. A regression checks both policies. Located parallel
relationship declarations retain distinct identities.

The source-save compatibility fix is separate: new atomic saves honor the
process creation mask, while existing files retain their mode. The stock
v0.13.0 writer unconditionally chmods new files after creating them, overriding
a strict mask. Two permission fixtures now explicitly establish the existing
0640 mode they test. The affected tests pass under mask 0077.

The stable client digest table includes v0.13.0 after hashing the actual ten
published service/kernel binaries and comparing their sidecars. No release
checksum is inferred from a source build or an unpublished artifact.

## Baseline adjudication

The specification-example corpus is regenerated with v0.13.0 from the same
pinned specification PDFs. Its 511-file digest is
`59f89a89c067c8231e86c8fab37798943ab5a5f12d9db1d48b30ef533e14fa26`,
matching official v0.13.0. Compared with the saved November extraction, only
SysML 7.17.12 examples 4 and 5 gain the generated `ISQMechanics::PowerValue`
import. The old digest is not accepted as a changed reference baseline.

The complete API comparison covers 951 files in nine roots. It records 115131
matched elements, 17474 ours-only and 26825 pilot-only. Property buckets are
2341623 present-equal, 168720 present-different, 186464 ours-missing, 89054
pilot-missing, 2519203 absent-both, 70042 indeterminate and 565
pilot-nondeterministic. Refusals, collisions and type mismatches do not increase.

The only changed existing semantic fragment is the self-model surfaces file,
whose authored frontend members describe the compatibility layer. Other
existing fragment movements are reference duplicate/library-element counts
from the two additional libraries. The fault-tree example adds one fragment.
The structure-reader differential changes only its examples root: 70 models,
54390 matched elements and 22 unmatched, versus 69/54061/22 at the official base.
These additional observations include existing symbol-versus-graph gaps; they
do not establish full metamodel agreement.

The diagnostic comparison covers 940 files: 542 fully agree, with 1119 agreed
diagnostics, 127 ours-only and 2455 pilot-only. The extra authored fault-tree
fixture accounts for the additional file and reference diagnostic. Both native
Turtle and API-JSON three-hop conversions of that fixture are byte-stable;
source-text-stripped structural equivalence is a separate check. The corpus
agreement ratchet retains its existing clause keys and disagreement classes;
new fault-tree exposures increase the existing MembershipExpose and NamespaceExpose
observations, not the list of accepted semantic behaviors. A draft update run
failed against the old per-part maxima before writing the complete new totals;
the final check runs against the adjudicated totals.

## Downstream acceptance boundaries

The unchanged eight Blue Dog inputs validate successfully with exactly the
same seven existing name-conflict warnings as the rollback compiler. These
warnings remain visible. FailureAnalysis semantics, causal traversal and the
current fault-tree renderer remain in the compatibility layer. SciPy/SLSQP and
the study equations are unchanged.

Native/WASM checks retain Flashlight's depth-one requirement context and the
authored Gorge depth-zero view's four requirement identities and three directed
derivations. The complete Blue Dog declaration projection still contains
15949 declarations and discloses 169413 omitted objects. Effective requirement
checks retain the native typing, specialization and containment path and its
relationship identity rather than drawing a new collapsed relationship.

The downstream review branches record their exact final compiler pin,
artifact hashes and application test receipts. Earlier preliminary runs and
interrupted runs are not substituted for final artifact checks.

## Separate FailureAnalysis architecture experiment

Stock v0.13.0 executes the external FailureAnalysis library and its isolated
review-row query. Its supported external-engine/tool composition and
declarative self-check rules are distinct from renderer registration. The
fork's internal ModelExtensions hook composes export and identity callbacks;
it is neither a public stock extension API nor a custom renderer registry.

Extracting the library does not remove the browser identity, projection,
lifecycle, memory, bounded-context or standalone Python snapshot requirements.
The separate architecture review preserves this distinction and does not
migrate the active compiler or model architecture.

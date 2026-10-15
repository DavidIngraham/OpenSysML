"""Standalone dataclass emission for resolved scalar SysML records."""
from typing import Optional, Sequence

from opensysml.generate import (
    Definition, Feature, _identifier, _literal, _hash_literal, class_names,
)


class UnsupportedDataclassError(ValueError):
    """A definition cannot be represented faithfully by a scalar snapshot."""


# Included in generated modules so snapshots need no OpenSysML installation.
_RUNTIME = r'''from __future__ import annotations
from dataclasses import dataclass, field, fields
from fractions import Fraction
from math import isfinite
from numbers import Integral, Real, Complex
from typing import ClassVar, Mapping, TypeVar

_T = TypeVar("_T", bound="_Record")


def _convert(value, primitive, lower, upper, collection, name):
    if collection:
        if not isinstance(value, (list, tuple)):
            raise TypeError(f"{name}: expected a sequence")
        if len(value) < lower or (upper is not None and len(value) > upper):
            raise ValueError(f"{name}: multiplicity violation")
        return tuple(_convert(v, primitive, 1, 1, False, name) for v in value)
    if value is None and lower == 0:
        return None
    if primitive == 'Boolean' and type(value) is bool:
        return value
    if primitive == 'String' and isinstance(value, str):
        return value
    if primitive in ('Integer', 'Natural') and isinstance(value, Integral) and not isinstance(value, bool):
        if primitive == 'Natural' and value < 0:
            raise ValueError(f"{name}: expected a nonnegative integer")
        return int(value)
    if primitive == 'Real' and isinstance(value, Real) and not isinstance(value, bool):
        result = float(value)
        if not isfinite(result):
            raise ValueError(f"{name}: expected a finite Real")
        return result
    if primitive == 'Rational' and isinstance(value, (Integral, Fraction)) and not isinstance(value, bool):
        return Fraction(int(value) if isinstance(value, Integral) else value)
    if primitive == 'Complex' and isinstance(value, Complex) and not isinstance(value, bool):
        result = complex(value)
        if not isfinite(result.real) or not isfinite(result.imag):
            raise ValueError(f"{name}: expected a finite Complex")
        return result
    raise TypeError(f"{name}: expected {primitive}")


@dataclass(frozen=True, slots=True, kw_only=True)
class _Record:
    sysml_id: ClassVar[str]
    sysml_types: ClassVar[tuple[str, ...]]

    def __post_init__(self):
        for f in fields(self):
            m = f.metadata
            object.__setattr__(self, f.name, _convert(getattr(self, f.name),
                m['primitive'], m['lower'], m['upper'], m['collection'], m['sysml_name']))

    @classmethod
    def from_mapping(cls: type[_T], values: Mapping[str, object]) -> _T:
        names = {f.metadata['sysml_name']: f.name for f in fields(cls)}
        if set(values) != set(names):
            raise ValueError(f"{cls.sysml_id}: missing or unexpected fields")
        return cls(**{names[k]: v for k, v in values.items()})

    @classmethod
    def from_instance(cls: type[_T], instance) -> _T:
        if getattr(instance, 'type_symbol_id', None) not in cls.sysml_types:
            raise TypeError(f"Expected an evaluated instance of {cls.sysml_id}")
        return cls.from_mapping({f.metadata['sysml_name']: instance[f.metadata['sysml_name']]
                                 for f in fields(cls)})

    def to_mapping(self) -> dict[str, object]:
        return {f.metadata['sysml_name']: getattr(self, f.name) for f in fields(self)}
'''


def render_dataclasses(definitions: Sequence[Definition], stamp: Optional[str] = None) -> str:
    """Emit frozen scalar snapshots, rejecting unsupported semantics explicitly.

    Inheritance is flattened into fields; snapshots are values, not model
    classifiers. Default expressions are evaluated by the model, never copied.
    """
    by_id = {d.id: d for d in definitions}
    names = class_names(definitions)
    if set(names.values()) & {'Fraction', 'Integral', 'Real', 'Complex', 'ClassVar', 'Mapping', 'TypeVar', 'field', 'fields', 'dataclass', 'isfinite', '_Record', '_T', '_convert', 'SYSML_MODEL_HASH', 'SYSML_GENERATOR_VERSION'}:
        raise UnsupportedDataclassError('Generated class shadows snapshot support code')
    if len(set(names.values())) != len(names):
        raise UnsupportedDataclassError('Generated class name collision')
    cache: dict[str, dict[str, Feature]] = {}

    def features(definition, visiting=()):
        if definition.id in visiting:
            raise UnsupportedDataclassError('Cyclic generalization: ' + definition.id)
        if definition.id in cache:
            return cache[definition.id]
        inherited = {}
        own = {f.name: f for f in definition.features}
        for base in definition.base_ids:
            if base not in by_id:
                raise UnsupportedDataclassError('Missing base definition: ' + base)
            for name, f in features(by_id[base], visiting + (definition.id,)).items():
                if name in inherited and inherited[name] != f and name not in own:
                    raise UnsupportedDataclassError('Ambiguous inherited feature: ' + name)
                inherited[name] = f
        inherited.update(own)
        cache[definition.id] = inherited
        return inherited

    lines = ['# Generated by opensysml.generate --format dataclass. Do not edit.',
             _RUNTIME, 'SYSML_GENERATOR_VERSION = "dataclass-2"',
             f'SYSML_MODEL_HASH = {_literal(_hash_literal(stamp))}']
    scalar_types = dict(Real='float', Integer='int', Natural='int', Boolean='bool',
                        String='str', Rational='Fraction', Complex='complex')
    reserved = {'sysml_id', 'sysml_types', 'from_mapping', 'to_mapping', 'from_instance', 'sysml_withheld_library_attributes'}
    for definition in sorted(definitions, key=lambda d: d.id):
        members = features(definition)
        accepted = []
        def specializes(d, target, seen=()):
            return d.id == target or any(b not in seen and specializes(by_id[b], target, seen+(d.id,))
                                        for b in d.base_ids if b in by_id)
        for candidate in definitions:
            if specializes(candidate, definition.id):
                accepted.append(candidate.id)
        lines += ['', '', '@dataclass(frozen=True, slots=True, kw_only=True)',
                  f'class {names[definition.id]}(_Record):',
                  f'    sysml_id: ClassVar[str] = {_literal(definition.id)}',
                  f'    sysml_types: ClassVar[tuple[str, ...]] = {tuple(sorted(accepted))!r}',
                  f'    sysml_withheld_library_attributes: ClassVar[int] = {definition.facts.withheld_library_attributes}']
        identifiers = set()
        for feature in members.values():
            t, m = feature.facts.type, feature.facts.multiplicity
            if not t or t.quantity or t.primitive not in scalar_types:
                raise UnsupportedDataclassError('Expected a resolved non-quantity scalar: ' + feature.facts.id)
            lower = 1
            upper: Optional[int] = 1
            if m and (m.lower or m.upper):
                try:
                    lower = int(m.lower)
                    upper = None if m.upper == '*' else int(m.upper)
                except ValueError as exc:
                    raise UnsupportedDataclassError('Unresolved multiplicity: ' + feature.facts.id) from exc
                if lower < 0 or (upper is not None and (upper < lower or upper < 1)):
                    raise UnsupportedDataclassError('Unsupported multiplicity: ' + feature.facts.id)
            collection = upper is None or upper > 1
            identifier = _identifier(feature.name)
            if identifier in reserved or identifier.startswith('_'):
                identifier = 'value_' + identifier
            if identifier in identifiers:
                raise UnsupportedDataclassError('Generated field name collision: ' + feature.facts.id)
            identifiers.add(identifier)
            annotation = scalar_types[t.primitive]
            if collection:
                annotation = f'tuple[{annotation}, ...]'
            elif lower == 0:
                annotation += ' | None'
            metadata = dict(sysml_name=feature.name, sysml_id=feature.facts.id,
                            declared_type=t.declared, unit=t.unit, primitive=t.primitive,
                            lower=lower, upper=upper, collection=collection)
            lines.append(f'    {identifier}: {annotation} = field(metadata={metadata!r})')
    return '\n'.join(lines) + '\n'

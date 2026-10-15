"""Scalar dataclass snapshots preserve the model contract without a service."""
import dataclasses
import sys
import types

import pytest
from opensysml.generate_records import render_dataclasses, UnsupportedDataclassError
from opensysml.typefacts import Multiplicity, Specialization, TypeFacts
from tests.test_generate import definition, feature, FakeInstance


def load(definitions):
    source = render_dataclasses(definitions)
    module = types.ModuleType('generated_records_test')
    sys.modules[module.__name__] = module
    exec(compile(source, '<generated>', 'exec'), module.__dict__)
    return module


def scalar(name, primitive='Real', **kwargs):
    return feature(name, type_facts=TypeFacts(primitive=primitive), **kwargs)


def test_frozen_independent_round_trip_and_validation():
    cls = load([definition('Demo::Design', features=[scalar('length_m'), scalar('enabled', 'Boolean')])]).Design
    original = {'length_m': 2, 'enabled': True}
    value = cls.from_mapping(original)
    assert dataclasses.is_dataclass(value)
    assert value.length_m == 2.0
    assert type(value.length_m) is float
    original['length_m'] = 3
    assert value.length_m == 2
    assert value.to_mapping() == {'length_m': 2.0, 'enabled': True}
    assert dataclasses.replace(value, length_m=1.4).length_m == 1.4
    with pytest.raises(dataclasses.FrozenInstanceError):
        value.length_m = 3
    with pytest.raises(TypeError):
        cls(length_m=True, enabled=True)
    with pytest.raises(ValueError):
        cls(length_m=float('nan'), enabled=True)
    with pytest.raises(ValueError):
        cls.from_mapping({'length_m': 2})
    with pytest.raises(ValueError):
        cls.from_mapping({'length_m': 2, 'enabled': True, 'unexpected': 1})


def test_optional_and_collection_are_detached_and_enforce_cardinality():
    cls = load([definition('Demo::Design', features=[
        scalar('optional', multiplicity=Multiplicity('0', '1')),
        scalar('samples', multiplicity=Multiplicity('2', '3'))])]).Design
    samples = [1, 2]
    value = cls(optional=None, samples=samples)
    samples.append(3)
    assert value.samples == (1., 2.)
    with pytest.raises(ValueError):
        cls(optional=None, samples=[1])
    with pytest.raises(TypeError):
        cls(optional=None, samples=[True, 2])


def test_inheritance_redefinition_and_evaluated_instance():
    base = definition('Demo::Base', features=[scalar('length_m')])
    child = definition('Demo::Child', features=[scalar('length_m', 'Integer'), scalar('mass')],
                       specializations=[Specialization('specializes', target_id=base.id)])
    m = load([child, base])
    value = m.Child.from_instance(FakeInstance({'length_m': 2, 'mass': 3}, child.id))
    assert value.mass == 3 and type(value.length_m) is int
    assert m.Base.from_instance(FakeInstance({'length_m': 2}, child.id)).length_m == 2
    with pytest.raises(TypeError):
        m.Child.from_instance(FakeInstance({'length_m': 2, 'mass': 3}, 'Unrelated'))
    assert dataclasses.fields(m.Child)[0].metadata['sysml_name'] == 'length_m'


@pytest.mark.parametrize('facts', [None, TypeFacts(declared='Unknown'), TypeFacts(primitive='Real', quantity=True), TypeFacts(resolved_id='Nested')])
def test_unsupported_types_are_explicit(facts):
    with pytest.raises(UnsupportedDataclassError):
        render_dataclasses([definition('Demo::D', features=[feature('x', type_facts=facts)])])


def test_collisions_cycles_and_missing_bases_are_explicit():
    with pytest.raises(UnsupportedDataclassError):
        render_dataclasses([definition('Demo::D', features=[scalar('a-b'), scalar('a_b')])])
    for target in ['Demo::D', 'Missing']:
        with pytest.raises(UnsupportedDataclassError):
            render_dataclasses([definition('Demo::D', specializations=[Specialization('specializes', target_id=target)])])


def test_deterministic_and_standard_library_only():
    defs = [definition('Z', features=[scalar('x')]), definition('A')]
    assert render_dataclasses(defs) == render_dataclasses(defs[::-1])
    assert 'from opensysml' not in render_dataclasses(defs)


def test_declaration_order_and_numeric_protocols():
    from fractions import Fraction
    cls = load([definition('Demo::Design', features=[scalar('width'), scalar('height')])]).Design
    assert [f.name for f in dataclasses.fields(cls)] == ['width', 'height']
    assert cls(width=Fraction(3, 2), height=2).width == 1.5


@pytest.mark.integration
def test_live_declaration_order_defaults_and_multifile_inheritance(tmp_path):
    from opensysml import Connection
    from opensysml.generate import generate_source
    base = 'package DemoBase { private import ScalarValues::*; part def Parent { attribute z : Real default = 2 * 3; } }'
    child = 'package Demo { private import ScalarValues::*; part def Child :> DemoBase::Parent { attribute a : Real = 4; } }'
    with Connection() as connection:
        model = connection.parse_sources([('base.sysml', base), ('child.sysml', child)], strict=True)
        source = generate_source(model, base + child, format='dataclass', selected=['Demo::Child'])
        module = types.ModuleType('ordered_snapshot_test')
        sys.modules[module.__name__] = module
        exec(compile(source, '<generated>', 'exec'), module.__dict__)
        assert [f.name for f in dataclasses.fields(module.Child)] == ['z', 'a']
        value = module.Child.from_instance(model.instantiate('Demo::Child'))
        assert value.to_mapping() == {'z': 6., 'a': 4.}
        with pytest.raises(TypeError):
            module.Child(a=4)


@pytest.mark.integration
def test_live_cli_generation_check_and_snapshot(tmp_path):
    from opensysml import Connection
    from opensysml.generate import main
    source = tmp_path / 'record.sysml'
    source.write_text('package Demo { private import ScalarValues::*; '
                      'part def Design { attribute length_m : Real = 1.4; } }')
    output = tmp_path / 'record.py'
    args = [str(source), '--format', 'dataclass', '--definition', 'Demo::Design', '-o', str(output)]
    assert main(args) == 0
    assert main(args + ['--check']) == 0
    module = types.ModuleType('live_snapshot_test')
    sys.modules[module.__name__] = module
    exec(compile(output.read_text(), str(output), 'exec'), module.__dict__)
    with Connection() as connection:
        model = connection.load(str(source))
        value = module.Design.from_instance(model.instantiate('Demo::Design'))
    assert value.length_m == 1.4
    assert value.to_mapping() == {'length_m': 1.4}
    output.write_text(output.read_text() + '# stale\n')
    assert main(args + ['--check']) == 1

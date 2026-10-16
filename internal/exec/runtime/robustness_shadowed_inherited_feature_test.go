package runtime

import (
	"errors"
	"testing"
)

// TestRuntimeRobustnessShadowedInheritedFeature exercises a specialization whose
// feature duplicates an inherited name without redefining it: the two are held
// apart, so an unvalued one is read as unset rather than as its namesake.
func TestRuntimeRobustnessShadowedInheritedFeature(t *testing.T) {
	t.Run("unvalued_shadowing_feature_does_not_hide_the_inherited_one", testShadowedUnvaluedShadowingFeature)
	t.Run("unvalued_inherited_feature_is_not_read_from_the_shadowing_one", testShadowedUnvaluedInheritedFeature)
}

func testShadowedUnvaluedShadowingFeature(t *testing.T) {
	results, err := executeActionSource(t, "Derived", `package test {
		private import ScalarValues::*;
		action def Base {
			attribute go : Boolean = true;
			attribute c : Integer := 0;
			first start;
			then action a;
			first a if go then b;
			action b { assign c := 1; }
		}
		action def Derived :> Base {
			attribute go : Boolean;
		}
	}`)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got := results["c"]; got.Const.Int != 1 {
		t.Fatalf("c = %v, want 1 from the inherited guard reading Base::go", got)
	}
	if got, ok := results["Base::go"]; !ok || !got.isBool() || !got.Const.Bool {
		t.Fatalf("Base::go = %v (held %v), want true", got, ok)
	}
	if _, ok := results["go"]; ok {
		t.Fatalf("go = %v, want the specialization's unvalued feature left unset", results["go"])
	}
}

func testShadowedUnvaluedInheritedFeature(t *testing.T) {
	_, err := executeActionSource(t, "Derived", `package test {
		private import ScalarValues::*;
		action def Base {
			attribute go : Boolean;
			attribute c : Integer := 0;
			first start;
			then action a;
			first a if go then b;
			action b { assign c := 1; }
		}
		action def Derived :> Base {
			attribute go : Boolean = false;
		}
	}`)
	var noValue *NoValueError
	if !errors.As(err, &noValue) {
		t.Fatalf("error = %v, want NoValueError for the general's unvalued go, not a read of the specialization's", err)
	}
	if noValue.Feature != "Base::go" {
		t.Fatalf("NoValueError.Feature = %q, want Base::go", noValue.Feature)
	}
}

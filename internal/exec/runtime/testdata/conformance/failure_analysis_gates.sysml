package test { private import FailureAnalysis::*;
occurrence a0 : BasicEvent { attribute :>> present = false; }
occurrence b0 : BasicEvent { attribute :>> present = false; }
occurrence or0 : OrGate { ref occurrence :>> inputs = (a0,b0); }
occurrence and0 : AndGate { ref occurrence :>> inputs = (a0,b0); }
occurrence a1 : BasicEvent { attribute :>> present = false; }
occurrence b1 : BasicEvent { attribute :>> present = true; }
occurrence or1 : OrGate { ref occurrence :>> inputs = (a1,b1); }
occurrence and1 : AndGate { ref occurrence :>> inputs = (a1,b1); }
occurrence a2 : BasicEvent { attribute :>> present = true; }
occurrence b2 : BasicEvent { attribute :>> present = false; }
occurrence or2 : OrGate { ref occurrence :>> inputs = (a2,b2); }
occurrence and2 : AndGate { ref occurrence :>> inputs = (a2,b2); }
occurrence a3 : BasicEvent { attribute :>> present = true; }
occurrence b3 : BasicEvent { attribute :>> present = true; }
occurrence or3 : OrGate { ref occurrence :>> inputs = (a3,b3); }
occurrence and3 : AndGate { ref occurrence :>> inputs = (a3,b3); }
analysis check {
out attribute or0 : ScalarValues::Boolean = test::or0.present;
out attribute and0 : ScalarValues::Boolean = test::and0.present;
out attribute or1 : ScalarValues::Boolean = test::or1.present;
out attribute and1 : ScalarValues::Boolean = test::and1.present;
out attribute or2 : ScalarValues::Boolean = test::or2.present;
out attribute and2 : ScalarValues::Boolean = test::and2.present;
out attribute or3 : ScalarValues::Boolean = test::or3.present;
out attribute and3 : ScalarValues::Boolean = test::and3.present;
}
}
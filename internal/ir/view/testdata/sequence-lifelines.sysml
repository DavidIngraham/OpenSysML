// Messages ordered by the successions on the lifelines their events are on:
// the send of one message and the receive of another on one lifeline order the
// two messages, whichever lifeline each message is received on.
package Signals {
	item def A;
	item def B;
	item def C;
}

package Lifelines {
	private import Signals::*;

	// q receives a, then sends b, then sends c; the receive events of b and c
	// are on two lifelines nothing orders against each other.
	occurrence def Relay {
		part p {
			event occurrence s1;
			then event occurrence r3;
		}

		part q {
			event occurrence r1;
			then event occurrence s2;
			then event occurrence s3;
		}

		part r {
			event occurrence r2;
		}

		message a of A from p.s1 to q.r1;
		message b of B from q.s2 to r.r2;
		message c of C from q.s3 to p.r3;
	}

	// The same interaction with its lifelines declared in the other order.
	occurrence def RelayReversed {
		part r {
			event occurrence r2;
		}

		part q {
			event occurrence r1;
			then event occurrence s2;
			then event occurrence s3;
		}

		part p {
			event occurrence s1;
			then event occurrence r3;
		}

		message a of A from p.s1 to q.r1;
		message b of B from q.s2 to r.r2;
		message c of C from q.s3 to p.r3;
	}

	// The only order runs through `settle`, an event no message attaches to.
	occurrence def Settled {
		part client {
			event occurrence askSent;
			event occurrence replyTaken;
		}

		part server {
			event occurrence askTaken;
			then event occurrence settle;
			then event occurrence replySent;
		}

		message reply of B from server.replySent to client.replyTaken;
		message ask of A from client.askSent to server.askTaken;
	}

	// Two messages nothing orders: `pong` is declared first and stays first.
	occurrence def Unordered {
		part left {
			event occurrence pongSent;
			event occurrence pingSent;
		}

		part right {
			event occurrence pongTaken;
			event occurrence pingTaken;
		}

		message pong of B from left.pongSent to right.pongTaken;
		message ping of A from left.pingSent to right.pingTaken;
	}
}

package LifelineViews {
	private import StandardViewDefinitions::*;

	view relayView : SequenceView {
		expose Lifelines::Relay;
	}

	view relayReversedView : SequenceView {
		expose Lifelines::RelayReversed;
	}

	view settledView : SequenceView {
		expose Lifelines::Settled;
	}

	view unorderedView : SequenceView {
		expose Lifelines::Unordered;
	}
}

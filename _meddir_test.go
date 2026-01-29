package sdp

import "testing"

func TestNegotiateAnswerMode_StateMachineAndIntersection(t *testing.T) {
	type tc struct {
		name    string
		local   string
		remote  string
		putHold bool
		want    string
	}

	tests := []tc{
		// Defaults
		{
			name:    "defaults-empty-local-and-remote-to-sendrecv",
			local:   "",
			remote:  "",
			putHold: false,
			want:    SendRecv,
		},

		// --- HOLD behavior (your state machine rule) ---
		{
			name:    "hold-when-already-recvonly-goes-inactive",
			local:   RecvOnly,
			remote:  SendOnly, // peer is sending; we were receiving
			putHold: true,
			want:    Inactive,
		},
		{
			name:    "hold-when-sendrecv-prefer-sendonly",
			local:   SendRecv,
			remote:  SendRecv,
			putHold: true,
			want:    SendOnly,
		},
		{
			name:    "hold-when-sendonly-stays-sendonly",
			local:   SendOnly,
			remote:  SendRecv,
			putHold: true,
			want:    SendOnly,
		},
		{
			name:    "hold-when-inactive-stays-inactive",
			local:   Inactive,
			remote:  SendRecv,
			putHold: true,
			want:    Inactive,
		},

		// --- UNHOLD/RESUME preference ---
		{
			name:    "resume-from-sendonly-prefers-sendrecv-if-offer-allows",
			local:   SendOnly,
			remote:  SendRecv,
			putHold: false,
			want:    SendRecv,
		},
		{
			name:    "resume-from-inactive-prefers-sendrecv-if-offer-allows",
			local:   Inactive,
			remote:  SendRecv,
			putHold: false,
			want:    SendRecv,
		},

		// --- Offer/Answer intersection sanity checks ---
		// Offer recvonly => offerer only receives => we can only send (if we choose to)
		{
			name:    "offer-recvonly-answer-sendonly-when-local-sendrecv",
			local:   SendRecv,
			remote:  RecvOnly,
			putHold: false,
			want:    SendOnly,
		},
		{
			name:    "offer-recvonly-answer-inactive-when-local-recvonly",
			local:   RecvOnly,
			remote:  RecvOnly,
			putHold: false,
			want:    Inactive, // both want to receive => no media possible
		},

		// Offer sendonly => offerer only sends => we can only recv (if we choose to)
		{
			name:    "offer-sendonly-answer-recvonly-when-local-sendrecv",
			local:   SendRecv,
			remote:  SendOnly,
			putHold: false,
			want:    RecvOnly,
		},
		{
			name:    "offer-sendonly-answer-inactive-when-local-sendonly",
			local:   SendOnly,
			remote:  SendOnly,
			putHold: false,
			want:    RecvOnly, // both want to send => no receiver
		},

		// Offer inactive => always inactive
		{
			name:    "offer-inactive-answer-inactive",
			local:   SendRecv,
			remote:  Inactive,
			putHold: false,
			want:    Inactive,
		},

		// --- Interplay: hold preference vs offer constraints ---
		{
			name:    "hold-prefers-sendonly-but-offer-sendonly-forces-inactive",
			local:   SendRecv,
			remote:  SendOnly, // offerer sends only; they will not receive
			putHold: true,     // we would like SendOnly hold, but can't send to someone not receiving
			want:    Inactive,
		},
		{
			name:    "hold-when-local-recvonly-is-inactive-regardless-of-offer",
			local:   RecvOnly,
			remote:  SendRecv,
			putHold: true,
			want:    Inactive,
		},

		// --- Missing remote direction defaults to sendrecv ---
		{
			name:    "empty-remote-treated-as-sendrecv",
			local:   SendOnly,
			remote:  "",
			putHold: false,
			want:    SendRecv, // resume from sendonly -> prefer sendrecv; offer allows it
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NegotiateAnswerMode(tt.local, tt.remote, tt.putHold)
			if got != tt.want {
				t.Fatalf("NegotiateAnswerMode(local=%q, remote=%q, putHold=%v) = %q; want %q",
					tt.local, tt.remote, tt.putHold, got, tt.want)
			}
		})
	}
}

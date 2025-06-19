package sdp_test

import (
	"testing"

	"github.com/Moatassem/sdp"
)

func TestSetConnection(t *testing.T) {
	t.Run("Check Effective Socket", func(t *testing.T) {
		ses, err := sdp.ParseString(`v=0
o=- 206 1 IN IP4 192.168.1.101
s=session
c=IN IP4 192.168.1.5
b=CT:1000
t=0 0
m=audio 51624 RTP/AVP 97 101 13 0 8
c=IN IP4 192.168.1.101
a=rtcp:51625
a=label:Audio
a=sendrecv
a=rtpmap:97 RED/8000
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=rtpmap:13 CN/8000
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=ptime:20
`)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		conn := ses.GetEffectiveMediaSocket(ses.GetMediaFlow(sdp.Audio))

		if conn != "192.168.1.101:51624" {
			t.Errorf("expected connection to be %s, got %s", "192.168.1.101:51624", conn)
		}
	})

	t.Run("Check Effective Socket with Media Flow with IPv4 = 0.0.0.0", func(t *testing.T) {
		ses, err := sdp.ParseString(`v=0
o=- 206 1 IN IP4 192.168.1.101
s=session
c=IN IP4 192.168.1.5
b=CT:1000
t=0 0
m=audio 51624 RTP/AVP 97 101 13 0 8
c=IN IP4 0.0.0.0
a=rtcp:51625
a=label:Audio
a=sendrecv
a=rtpmap:97 RED/8000
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=rtpmap:13 CN/8000
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=ptime:20
`)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		conn := ses.GetEffectiveMediaSocket(ses.GetMediaFlow(sdp.Audio))

		if conn != "192.168.1.5:51624" {
			t.Errorf("expected connection to be %s, got %s", "192.168.1.5:51624", conn)
		}
	})

	t.Run("Check Effective Socket with Media Flow with no IPv4", func(t *testing.T) {
		ses, err := sdp.ParseString(`v=0
o=- 206 1 IN IP4 192.168.1.101
s=session
c=IN IP4 192.168.1.5
b=CT:1000
t=0 0
m=audio 51624 RTP/AVP 97 101 13 0 8
a=rtcp:51625
a=label:Audio
a=sendrecv
a=rtpmap:97 RED/8000
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=rtpmap:13 CN/8000
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=ptime:20
`)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		conn := ses.GetEffectiveMediaSocket(ses.GetMediaFlow(sdp.Audio))

		if conn != "192.168.1.5:51624" {
			t.Errorf("expected connection to be %s, got %s", "192.168.1.5:51624", conn)
		}
	})

	t.Run("Set and Check Effective Socket", func(t *testing.T) {
		ses, err := sdp.ParseString(`v=0
o=- 2508 1 IN IP4 192.168.1.2
s=sipclientgo/1.0
c=IN IP4 192.168.1.2
t=0 0
m=audio 51191 RTP/AVP 9 8 0 101
a=rtpmap:9 G722/8000
a=rtpmap:8 PCMA/8000
a=rtpmap:0 PCMU/8000
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=sendonly
a=ssrc:7345055
`)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		ses.SetConnection(sdp.Audio, "10.0.0.1", 55841)

		conn := ses.GetEffectiveMediaSocket(ses.GetMediaFlow(sdp.Audio))

		if conn != "10.0.0.1:55841" {
			t.Errorf("expected connection to be %s, got %s", "10.0.0.1:55841", conn)
		}
	})
}

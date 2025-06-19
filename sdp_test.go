package sdp_test

import (
	"fmt"
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

		ses.SetConnection(sdp.Audio, "10.0.0.1", 55841, false)

		conn := ses.GetEffectiveMediaSocket(ses.GetMediaFlow(sdp.Audio))

		if conn != "10.0.0.1:55841" {
			t.Errorf("expected connection to be %s, got %s", "10.0.0.1:55841", conn)
		}
	})

	t.Run("Set and Check Effective Socket globalized", func(t *testing.T) {
		ses, err := sdp.ParseString(`v=0
o=- 3849203748 3849203748 IN IP4 192.0.2.1
s=Multimedia Session Example
t=0 0
a=group:BUNDLE audio video data
a=msid-semantic: WMS myStream
c=IN IP4 203.0.113.1
m=audio 49170 RTP/AVP 0 96
c=IN IP4 203.0.113.2
a=rtpmap:0 PCMU/8000
a=rtpmap:96 opus/48000/2
a=fmtp:96 minptime=10;useinbandfec=1
a=sendrecv
a=mid:audio
a=ssrc:1001 cname:audioCname
m=video 51372 RTP/AVP 97 98
c=IN IP4 203.0.113.3
a=rtpmap:97 H264/90000
a=rtpmap:98 VP8/90000
a=fmtp:97 profile-level-id=42e01f;packetization-mode=1
a=sendrecv
a=mid:video
a=ssrc:1002 cname:videoCname
m=application 50000 DTLS/SCTP 5000
c=IN IP4 203.0.113.4
a=mid:data
a=sctpmap:5000 webrtc-datachannel 1024
`)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		ses.SetConnection(sdp.Audio, "192.168.1.2", 55841, true)

		fmt.Println(ses.String())

		conn := ses.GetEffectiveMediaSocket(ses.GetMediaFlow(sdp.Audio))

		if conn != "192.168.1.2:55841" {
			t.Errorf("expected connection to be %s, got %s", "192.168.1.2:55841", conn)
		}
	})

}

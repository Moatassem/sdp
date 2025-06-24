package sdp_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/Moatassem/sdp"
)

func TestParseSDP(t *testing.T) {
	t.Run("Parse SDP", func(t *testing.T) {
		sdpString := `v=0
o=- 2508 1 IN IP4 192.168.1.2
s=sipclientgo/1.0
c=IN IP4 192.168.1.2
t=0 0
m=audio 51191 RTP/AVP 9 8 0 101
a=rtpmap:9 G722/8000/3
a=rtpmap:8 PCMA/8000/2
a=rtpmap:0 PCMU/8000/1
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=sendonly
a=ssrc:7345055
`
		ses, err := sdp.ParseString(sdpString)

		t.Run("No Error", func(t *testing.T) {
			if err != nil {
				t.Fatalf("failed to parse SDP: %v", err)
			}
		})

		t.Run("G722 Channels", func(t *testing.T) {
			g722channels := ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.G722).Channels
			if g722channels != 3 {
				t.Errorf("expected G722 channels to be 3, got %d", g722channels)
			}
		})

		t.Run("PCMA Channels", func(t *testing.T) {
			pcmaChannels := ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.PCMA).Channels
			if pcmaChannels != 2 {
				t.Errorf("expected PCMA channels to be 2, got %d", pcmaChannels)
			}
		})
		t.Run("PCMU Channels", func(t *testing.T) {
			pcmuChannels := ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.PCMU).Channels
			if pcmuChannels != 1 {
				t.Errorf("expected PCMU channels to be 1, got %d", pcmuChannels)
			}
		})

		t.Run("Telephone Event Channels", func(t *testing.T) {
			telephoneEventChannels := ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.RFC4733).Channels
			if telephoneEventChannels != 1 {
				t.Errorf("expected Telephone Event channels to be 1, got %d", telephoneEventChannels)
			}
		})

	})
}

func TestDropSDPByName(t *testing.T) {
	t.Run("Parse SDP", func(t *testing.T) {
		sdpString := `v=0
o=- 2508 1 IN IP4 192.168.1.2
s=sipclientgo/1.0
c=IN IP4 192.168.1.2
t=0 0
m=audio 51191 RTP/AVP 9 8 0 101
a=rtpmap:9 G722/8000/3
a=rtpmap:8 PCMA/8000/2
a=rtpmap:0 PCMU/8000/1
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=sendonly
a=ssrc:7345055
`

		t.Run("No Error", func(t *testing.T) {
			if _, err := sdp.ParseString(sdpString); err != nil {
				t.Fatalf("failed to parse SDP: %v", err)
			}
		})

		t.Run("Drop G722 only", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).DropFormatsByName(sdp.G722)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.G722) != nil {
				t.Errorf("expected G722 format to be dropped, but it still exists")
			}
		})

		t.Run("Drop G722 & PCMA", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).DropFormatsByName(sdp.G722, sdp.PCMA)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.G722) != nil {
				t.Errorf("expected G722 format to be dropped, but it still exists")
			}
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.PCMA) != nil {
				t.Errorf("expected PCMA format to be dropped, but it still exists")
			}
		})

		t.Run("Drop PCMA & RFC4733", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).DropFormatsByName(sdp.PCMA, sdp.RFC4733)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.PCMA) != nil {
				t.Errorf("expected PCMA format to be dropped, but it still exists")
			}
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.RFC4733) != nil {
				t.Errorf("expected RFC4733 format to be dropped, but it still exists")
			}
		})

		t.Run("Drop RFC4733 only", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).DropFormatsByName(sdp.RFC4733)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.RFC4733) != nil {
				t.Errorf("expected RFC4733 format to be dropped, but it still exists")
			}
		})

	})
}

func TestFilterSDPByName(t *testing.T) {
	t.Run("Parse SDP", func(t *testing.T) {
		sdpString := `v=0
o=- 2508 1 IN IP4 192.168.1.2
s=sipclientgo/1.0
c=IN IP4 192.168.1.2
t=0 0
m=audio 51191 RTP/AVP 9 8 0 101
a=rtpmap:9 G722/8000/3
a=rtpmap:8 PCMA/8000/2
a=rtpmap:0 PCMU/8000/1
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=sendonly
a=ssrc:7345055
`

		t.Run("No Error", func(t *testing.T) {
			if _, err := sdp.ParseString(sdpString); err != nil {
				t.Fatalf("failed to parse SDP: %v", err)
			}
		})

		t.Run("Filter G722 only", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).FilterFormatsByName(sdp.G722)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.G722) == nil || len(ses.GetMediaFlow(sdp.Audio).Format) != 1 {
				t.Errorf("expected G722 format to be filtered, but was dropped")
			}
		})

		t.Run("Filter G722 & PCMA", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).FilterFormatsByName(sdp.G722, sdp.PCMA)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.G722) == nil || len(ses.GetMediaFlow(sdp.Audio).Format) != 2 {
				t.Errorf("expected G722 format to be filtered, but was dropped")
			}
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.PCMA) == nil || len(ses.GetMediaFlow(sdp.Audio).Format) != 2 {
				t.Errorf("expected PCMA format to be filtered, but was dropped")
			}
		})

		t.Run("Filter PCMA & RFC4733", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).FilterFormatsByName(sdp.PCMA, sdp.RFC4733)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.PCMA) == nil || len(ses.GetMediaFlow(sdp.Audio).Format) != 2 {
				t.Errorf("expected PCMA & RFC4733 format to be filtered, but was dropped")
			}
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.RFC4733) == nil || len(ses.GetMediaFlow(sdp.Audio).Format) != 2 {
				t.Errorf("expected RFC4733 format to be filtered, but was dropped")
			}
		})

		t.Run("Filter RFC4733 only", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).FilterFormatsByName(sdp.RFC4733)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.RFC4733) == nil || len(ses.GetMediaFlow(sdp.Audio).Format) != 1 {
				t.Errorf("expected RFC4733 format to be filtered, but was dropped")
			}
		})

	})
}

func TestDropSDPByPayload(t *testing.T) {
	t.Run("Parse SDP", func(t *testing.T) {
		sdpString := `v=0
o=- 2508 1 IN IP4 192.168.1.2
s=sipclientgo/1.0
c=IN IP4 192.168.1.2
t=0 0
m=audio 51191 RTP/AVP 9 8 0 101
a=rtpmap:9 G722/8000/3
a=rtpmap:8 PCMA/8000/2
a=rtpmap:0 PCMU/8000/1
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=sendonly
a=ssrc:7345055
`

		t.Run("No Error", func(t *testing.T) {
			if _, err := sdp.ParseString(sdpString); err != nil {
				t.Fatalf("failed to parse SDP: %v", err)
			}
		})

		t.Run("Drop G722 only", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).DropFormatsByPayload(sdp.G722)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.G722) != nil {
				t.Errorf("expected G722 format to be dropped, but it still exists")
			}
		})

		t.Run("Drop G722 & PCMA", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).DropFormatsByPayload(sdp.G722, sdp.PCMA)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.G722) != nil {
				t.Errorf("expected G722 format to be dropped, but it still exists")
			}
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.PCMA) != nil {
				t.Errorf("expected PCMA format to be dropped, but it still exists")
			}
		})

		t.Run("Drop PCMA & RFC4733", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).DropFormatsByPayload(sdp.PCMA, sdp.RFC4733)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.PCMA) != nil {
				t.Errorf("expected PCMA format to be dropped, but it still exists")
			}
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.RFC4733) != nil {
				t.Errorf("expected RFC4733 format to be dropped, but it still exists")
			}
		})

		t.Run("Drop RFC4733 only", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).DropFormatsByPayload(sdp.RFC4733)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.RFC4733) != nil {
				t.Errorf("expected RFC4733 format to be dropped, but it still exists")
			}
		})

	})
}

func TestFilterSDPByPayload(t *testing.T) {
	t.Run("Parse SDP", func(t *testing.T) {
		sdpString := `v=0
o=- 2508 1 IN IP4 192.168.1.2
s=sipclientgo/1.0
c=IN IP4 192.168.1.2
t=0 0
m=audio 51191 RTP/AVP 9 8 0 101
a=rtpmap:9 G722/8000/3
a=rtpmap:8 PCMA/8000/2
a=rtpmap:0 PCMU/8000/1
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=sendonly
a=ssrc:7345055
`

		t.Run("No Error", func(t *testing.T) {
			if _, err := sdp.ParseString(sdpString); err != nil {
				t.Fatalf("failed to parse SDP: %v", err)
			}
		})

		t.Run("Filter G722 only", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).FilterFormatsByPayload(sdp.G722)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.G722) == nil || len(ses.GetMediaFlow(sdp.Audio).Format) != 1 {
				t.Errorf("expected G722 format to be filtered, but was dropped")
			}
		})

		t.Run("Filter G722 & PCMA", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).FilterFormatsByPayload(sdp.G722, sdp.PCMA)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.G722) == nil || len(ses.GetMediaFlow(sdp.Audio).Format) != 2 {
				t.Errorf("expected G722 format to be filtered, but was dropped")
			}
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.PCMA) == nil || len(ses.GetMediaFlow(sdp.Audio).Format) != 2 {
				t.Errorf("expected PCMA format to be filtered, but was dropped")
			}
		})

		t.Run("Filter PCMA & RFC4733", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).FilterFormatsByPayload(sdp.PCMA, sdp.RFC4733)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.PCMA) == nil || len(ses.GetMediaFlow(sdp.Audio).Format) != 2 {
				t.Errorf("expected PCMA & RFC4733 format to be filtered, but was dropped")
			}
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.RFC4733) == nil || len(ses.GetMediaFlow(sdp.Audio).Format) != 2 {
				t.Errorf("expected RFC4733 format to be filtered, but was dropped")
			}
		})

		t.Run("Filter RFC4733 only", func(t *testing.T) {
			ses, _ := sdp.ParseString(sdpString)
			ses.GetMediaFlow(sdp.Audio).FilterFormatsByPayload(sdp.RFC4733)
			if ses.GetMediaFlow(sdp.Audio).FormatByName(sdp.RFC4733) == nil || len(ses.GetMediaFlow(sdp.Audio).Format) != 1 {
				t.Errorf("expected RFC4733 format to be filtered, but was dropped")
			}
		})

	})

}

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

		addr := ses.GetEffectiveMediaUdpAddr(sdp.Audio)
		if addr.String() != "10.0.0.1:55841" {
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

		addr := ses.GetEffectiveMediaUdpAddr(sdp.Audio)
		if addr.String() != "192.168.1.2:55841" {
			t.Errorf("expected connection to be %s, got %s", "192.168.1.2:55841", conn)
		}
	})
}

func BenchmarkEqualSDP(b *testing.B) {
	sdp1 := `v=0
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
`
	sdp2 := `v=0
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
`

	b.Run("Hash 256 SHA", func(b *testing.B) {
		hash1 := hashSDPBytes([]byte(sdp1))
		for b.Loop() {
			bytes2 := []byte(sdp2)
			_ = hash1 == hashSDPBytes(bytes2)
			_ = string(bytes2)
		}
	})

	// b.Run("Equal SDP - One computed", func(b *testing.B) {
	// 	ses1, _ := sdp.ParseString(sdp1)
	// 	for b.Loop() {
	// 		ses2, _ := sdp.ParseString(sdp2)
	// 		_ = ses1.Equals(ses2)
	// 	}
	// })

	b.Run("Equal SDP - Precomputed", func(b *testing.B) {
		ses1, _ := sdp.ParseString(sdp1)
		ses2, _ := sdp.ParseString(sdp2)
		for b.Loop() {
			_ = ses1.Equals(ses2)
			_ = ses2.Bytes()
		}
	})
}

func bytesToHexString(data []byte) string {
	return hex.EncodeToString(data)
}

func hashSDPBytes(bytes []byte) string {
	hash := sha256.Sum256(bytes)
	return bytesToHexString(hash[:])
}

package sdp_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
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
		ses, _, err := sdp.ParseString(sdpString, false)

		t.Run("No Error", func(t *testing.T) {
			if err != nil {
				t.Fatalf("failed to parse SDP: %v", err)
			}
		})

		t.Run("Exact Clone", func(t *testing.T) {
			if !ses.Equals(ses.Clone()) {
				t.Fatalf("expected clone to be equal to original session, but they are not")
			}
		})

		t.Run("G722 Channels", func(t *testing.T) {
			ses1 := ses.Clone()
			g722channels := ses1.GetMediaFlow(sdp.Audio).FormatByName("G722").Channels
			if g722channels != 3 {
				t.Errorf("expected G722 channels to be 3, got %d", g722channels)
			}
		})

		t.Run("PCMA Channels", func(t *testing.T) {
			ses1 := ses.Clone()
			pcmaChannels := ses1.GetMediaFlow(sdp.Audio).FormatByName("PCMA").Channels
			if pcmaChannels != 2 {
				t.Errorf("expected PCMA channels to be 2, got %d", pcmaChannels)
			}
		})
		t.Run("PCMU Channels", func(t *testing.T) {
			ses1 := ses.Clone()
			pcmuChannels := ses1.GetMediaFlow(sdp.Audio).FormatByName("PCMU").Channels
			if pcmuChannels != 1 {
				t.Errorf("expected PCMU channels to be 1, got %d", pcmuChannels)
			}
		})

		t.Run("Telephone Event Channels", func(t *testing.T) {
			ses1 := ses.Clone()
			telephoneEventChannels := ses1.GetMediaFlow(sdp.Audio).FormatByName(sdp.GetCodecName(sdp.RFC4733PT)).Channels
			if telephoneEventChannels != 1 {
				t.Errorf("expected Telephone Event channels to be 1, got %d", telephoneEventChannels)
			}
		})

	})
}

func TestRestoreMissingStaticPTs(t *testing.T) {
	ses1, _, err := sdp.ParseString(`v=0
o=- 3849203748 3849203748 IN IP4 192.0.2.1
s=Multimedia Session Example
t=0 0
a=group:BUNDLE audio video data
a=msid-semantic: WMS myStream
c=IN IP4 203.0.113.1
m=audio 49170 RTP/AVP 0 8 9 18 96 97 98 99 100 101 102 103 104
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=rtpmap:9 G722/8000
a=rtpmap:18 G729/8000
a=rtpmap:96 opus/48000/2
a=fmtp:96 minptime=10;useinbandfec=1
a=rtpmap:97 AMR/8000
a=rtpmap:98 AMR-WB/16000
a=rtpmap:99 speex/8000
a=rtpmap:100 iLBC/8000
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=rtpmap:102 GSM/8000
a=rtpmap:103 LPC/8000
a=rtpmap:104 SILK/24000
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
`, false)

	if err != nil {
		t.Fatalf("failed to parse SDP: %v", err)
	}

	ses2, _, err := sdp.ParseString(`v=0
o=- 3849203748 3849203748 IN IP4 192.0.2.1
s=Multimedia Session Example
t=0 0
a=group:BUNDLE audio video data
a=msid-semantic: WMS myStream
c=IN IP4 203.0.113.1
m=audio 49170 RTP/AVP 0 8 9 18 96 97 98 99 100 101 102 103 104
a=rtpmap:96 opus/48000/2
a=fmtp:96 minptime=10;useinbandfec=1
a=rtpmap:97 AMR/8000
a=rtpmap:98 AMR-WB/16000
a=rtpmap:99 speex/8000
a=rtpmap:100 iLBC/8000
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=rtpmap:102 GSM/8000
a=rtpmap:103 LPC/8000
a=rtpmap:104 SILK/24000
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
`, false)

	if err != nil {
		t.Fatalf("failed to parse SDP: %v", err)
	}

	fmt.Println(ses2.String())

	ses2.RestoreMissingRtpmaps()

	fmt.Println(ses2.String())

	if !ses1.Equals(ses2) {
		t.Error("ses1 not equal to ses2")
	}

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
		ses, _, err := sdp.ParseString(sdpString, false)

		t.Run("No Error", func(t *testing.T) {
			if err != nil {
				t.Fatalf("failed to parse SDP: %v", err)
			}
		})

		t.Run("Drop G722 only", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).DropFormatsByName("g722")
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("g722") != nil {
				t.Errorf("expected G722 format to be dropped, but it still exists")
			}
		})

		t.Run("Drop G722 & PCMA", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).DropFormatsByName("G722", "PCMA")
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("G722") != nil {
				t.Errorf("expected G722 format to be dropped, but it still exists")
			}
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("PCMA") != nil {
				t.Errorf("expected PCMA format to be dropped, but it still exists")
			}
		})

		t.Run("Drop PCMA & RFC4733", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).DropFormatsByName("PCMA", "RFC4733")
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("PCMA") != nil {
				t.Errorf("expected PCMA format to be dropped, but it still exists")
			}
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("RFC4733") != nil {
				t.Errorf("expected RFC4733 format to be dropped, but it still exists")
			}
		})

		t.Run("Drop RFC4733 only", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).DropFormatsByName("RFC4733")
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("RFC4733") != nil {
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
		ses, _, err := sdp.ParseString(sdpString, false)

		t.Run("No Error", func(t *testing.T) {
			if err != nil {
				t.Fatalf("failed to parse SDP: %v", err)
			}
		})

		t.Run("Filter G722 only", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).FilterFormatsByName("g722")
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("g722") == nil || len(ses1.GetMediaFlow(sdp.Audio).Formats) != 1 {
				t.Errorf("expected G722 format to be filtered, but was dropped")
			}
		})

		t.Run("Filter G722 & PCMA", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).FilterFormatsByName("G722", "PCMA")
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("G722") == nil || len(ses1.GetMediaFlow(sdp.Audio).Formats) != 2 {
				t.Errorf("expected G722 format to be filtered, but was dropped")
			}
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("PCMA") == nil || len(ses1.GetMediaFlow(sdp.Audio).Formats) != 2 {
				t.Errorf("expected PCMA format to be filtered, but was dropped")
			}
		})

		t.Run("Filter PCMA & RFC4733", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).FilterFormatsByName("PCMa", sdp.RFC4733) //sdp.GetCodecName(sdp.RFC4733PT)
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("PCMA") == nil || len(ses1.GetMediaFlow(sdp.Audio).Formats) != 2 {
				t.Errorf("expected PCMA & RFC4733 format to be filtered, but was dropped")
			}
			if ses1.GetMediaFlow(sdp.Audio).FormatByName(sdp.GetCodecName(sdp.RFC4733PT)) == nil || len(ses1.GetMediaFlow(sdp.Audio).Formats) != 2 {
				t.Errorf("expected RFC4733 format to be filtered, but was dropped")
			}
		})

		t.Run("Filter RFC4733 only", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).FilterFormatsByName(sdp.GetCodecName(sdp.RFC4733PT))
			if ses1.GetMediaFlow(sdp.Audio).FormatByName(sdp.GetCodecName(sdp.RFC4733PT)) == nil || len(ses1.GetMediaFlow(sdp.Audio).Formats) != 1 {
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
		ses, _, err := sdp.ParseString(sdpString, false)

		t.Run("No Error", func(t *testing.T) {
			if err != nil {
				t.Fatalf("failed to parse SDP: %v", err)
			}
		})

		t.Run("Drop G722 only", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).DropFormatsByPayload(sdp.G722)
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("G722") != nil {
				t.Errorf("expected G722 format to be dropped, but it still exists")
			}
		})

		t.Run("Drop G722 & PCMA", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).DropFormatsByPayload(sdp.G722, sdp.PCMA)
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("G722") != nil {
				t.Errorf("expected G722 format to be dropped, but it still exists")
			}
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("PCMA") != nil {
				t.Errorf("expected PCMA format to be dropped, but it still exists")
			}
		})

		t.Run("Drop PCMA & RFC4733", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).DropFormatsByPayload(sdp.PCMA, sdp.RFC4733PT)
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("PCMA") != nil {
				t.Errorf("expected PCMA format to be dropped, but it still exists")
			}
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("RFC4733") != nil {
				t.Errorf("expected RFC4733 format to be dropped, but it still exists")
			}
		})

		t.Run("Drop RFC4733 only", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).DropFormatsByPayload(sdp.RFC4733PT)
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("RFC4733") != nil {
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

		ses, _, err := sdp.ParseString(sdpString, false)
		t.Run("No Error", func(t *testing.T) {
			if err != nil {
				t.Fatalf("failed to parse SDP: %v", err)
			}
		})

		t.Run("Filter G722 only", func(t *testing.T) {
			ses1 := ses.Clone()
			audioFlow := ses1.GetMediaFlow(sdp.Audio)
			audioFlow.FilterFormatsByPayload(sdp.G722)
			if audioFlow.FormatByName("G722") == nil || len(audioFlow.Formats) != 1 {
				t.Errorf("expected G722 format to be filtered, but was dropped")
			}
		})

		t.Run("Filter G722 & PCMA", func(t *testing.T) {
			ses1 := ses.Clone()
			audioFlow := ses1.GetMediaFlow(sdp.Audio)
			audioFlow.FilterFormatsByPayload(sdp.G722, sdp.PCMA)
			if audioFlow.FormatByName("G722") == nil || len(audioFlow.Formats) != 2 {
				t.Errorf("expected G722 format to be filtered, but was dropped")
			}
			if audioFlow.FormatByName("PCMA") == nil || len(audioFlow.Formats) != 2 {
				t.Errorf("expected PCMA format to be filtered, but was dropped")
			}
		})

		t.Run("Filter PCMA & RFC4733", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).FilterFormatsByPayload(sdp.PCMA, sdp.RFC4733PT)
			if ses1.GetMediaFlow(sdp.Audio).FormatByName("PCMA") == nil || len(ses1.GetMediaFlow(sdp.Audio).Formats) != 2 {
				t.Errorf("expected PCMA & RFC4733 format to be filtered, but was dropped")
			}
			if ses1.GetMediaFlow(sdp.Audio).FormatByName(sdp.GetCodecName(sdp.RFC4733PT)) == nil || len(ses1.GetMediaFlow(sdp.Audio).Formats) != 2 {
				t.Errorf("expected RFC4733 format to be filtered, but was dropped")
			}
		})

		t.Run("Filter RFC4733 only", func(t *testing.T) {
			ses1 := ses.Clone()
			ses1.GetMediaFlow(sdp.Audio).FilterFormatsByPayload(sdp.RFC4733PT)
			if ses1.GetMediaFlow(sdp.Audio).FormatByName(sdp.GetCodecName(sdp.RFC4733PT)) == nil || len(ses1.GetMediaFlow(sdp.Audio).Formats) != 1 {
				t.Errorf("expected RFC4733 format to be filtered, but was dropped")
			}
		})

	})

}

func TestSetConnection(t *testing.T) {
	t.Run("Check Effective Socket", func(t *testing.T) {
		ses, _, err := sdp.ParseString(`v=0
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		conn := ses.GetEffectiveMediaSocket(ses.GetMediaFlow(sdp.Audio))

		if conn != "192.168.1.101:51624" {
			t.Errorf("expected connection to be %s, got %s", "192.168.1.101:51624", conn)
		}
	})

	t.Run("Check Effective Socket with Media Flow with IPv4 = 0.0.0.0", func(t *testing.T) {
		ses, _, err := sdp.ParseString(`v=0
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		conn := ses.GetEffectiveMediaSocket(ses.GetMediaFlow(sdp.Audio))

		if conn != "192.168.1.5:51624" {
			t.Errorf("expected connection to be %s, got %s", "192.168.1.5:51624", conn)
		}
	})

	t.Run("Check Effective Socket with Media Flow with no IPv4", func(t *testing.T) {
		ses, _, err := sdp.ParseString(`v=0
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		conn := ses.GetEffectiveMediaSocket(ses.GetMediaFlow(sdp.Audio))

		if conn != "192.168.1.5:51624" {
			t.Errorf("expected connection to be %s, got %s", "192.168.1.5:51624", conn)
		}
	})

	t.Run("Set and Check Effective Socket", func(t *testing.T) {
		ses, _, err := sdp.ParseString(`v=0
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
`, false)

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
		ses, _, err := sdp.ParseString(`v=0
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		ses.SetConnection(sdp.Audio, "192.168.1.2", 55841, true)

		// fmt.Println(ses.String())

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

func TestGenerateSDPAnswer(t *testing.T) {

	t.Run("Extended SDP Offer", func(t *testing.T) {
		ses1, _, err := sdp.ParseString(`v=0
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		ses1.DisableFlows(sdp.Application, sdp.Video)
		if ses1.GetMediaFlow(sdp.Application).Port != 0 {
			t.Errorf("expected Application media port to be 0, got %d", ses1.GetMediaFlow(sdp.Application).Port)
		}
		ses1.GetMediaFlow(sdp.Application).Port = 23222
		if ses1.GetMediaFlow(sdp.Video).Port != 0 {
			t.Errorf("expected Video media port to be 0, got %d", ses1.GetMediaFlow(sdp.Video).Port)
		}
		ses1.GetMediaFlow(sdp.Video).Port = 23233

		ses1.DisableFlowsExcept(sdp.Audio)
		for _, media := range ses1.Media {
			if media.Type != sdp.Audio && media.Port != 0 {
				t.Errorf("expected media port to be 0 for %s, got %d", media.Type, media.Port)
			}
		}

		if ses1.AreAllFlowsDroppedOrDisabled() {
			t.Errorf("expected not all flows to be dropped or disabled, but they are")
		}

		ses2, _, err := sdp.ParseString(`v=0
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		if err := ses2.AlignMediaFlows(ses1); err != nil {
			t.Errorf("expected error to be nil, got %s", err)
		}

		for i, media := range ses2.Media {
			if media.Type != ses1.Media[i].Type {
				t.Errorf("expected media type %s, got %s", ses1.Media[i].Type, media.Type)
			}
			if media.Type != sdp.Audio && media.Port != 0 {
				t.Errorf("expected media port to be 0 for %s, got %d", media.Type, media.Port)
			}
		}
	})

	t.Run("Drop Some flows in Extended SDP Offer", func(t *testing.T) {
		ses1, _, err := sdp.ParseString(`v=0
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		ses1.DropFlowsExcept(sdp.Audio)

		if len(ses1.Media) != 1 || ses1.GetAudioMediaFlow() == nil {
			t.Errorf("expected only Audio media flow to remain, got %d media flows", len(ses1.Media))
		}

		ses2, _, err := sdp.ParseString(`v=0
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		if err := ses2.AlignMediaFlows(ses1); err != nil {
			t.Errorf("expected error to be nil, got %s", err)
		}

		for i, media := range ses2.Media {
			if media.Type != ses1.Media[i].Type {
				t.Errorf("expected media type %s, got %s", ses1.Media[i].Type, media.Type)
			}
			if media.Type != sdp.Audio && media.Port != 0 {
				t.Errorf("expected media port to be 0 for %s, got %d", media.Type, media.Port)
			}
		}
	})

	t.Run("Drop All flows in Extended SDP Offer", func(t *testing.T) {
		ses1, _, err := sdp.ParseString(`v=0
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		ses1.DropFlows(sdp.Audio, sdp.Video)

		if len(ses1.Media) != 1 && ses1.GetMediaFlow(sdp.Application) != nil {
			t.Errorf("expected all flows to be dropped or disabled, but they are not")
		}
	})

	t.Run("Disable Some flows in Extended SDP Offer", func(t *testing.T) {
		ses1, _, err := sdp.ParseString(`v=0
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		ses1.DisableFlowsExcept(sdp.Application)

		if ses1.GetMediaFlow(sdp.Application).Port == 0 {
			t.Error("expected Application media port to be non-zero, got 0")
		}

		ses2, _, err := sdp.ParseString(`v=0
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		if err := ses2.AlignMediaFlows(ses1); err != nil {
			t.Errorf("expected error to be nil, got %s", err)
		}

		for i, media := range ses2.Media {
			if media.Type != ses1.Media[i].Type {
				t.Errorf("expected media type %s, got %s", ses1.Media[i].Type, media.Type)
			}
			if media.Type != sdp.Audio && media.Port != 0 {
				t.Errorf("expected media port to be 0 for %s, got %d", media.Type, media.Port)
			}
		}
	})

	t.Run("Disable All flows in Extended SDP Offer", func(t *testing.T) {
		ses1, _, err := sdp.ParseString(`v=0
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		ses1.DisableFlows(sdp.Video, sdp.Application)

		if ses1.GetMediaFlow(sdp.Video).Port != 0 || ses1.GetMediaFlow(sdp.Application).Port != 0 {
			t.Errorf("expected Video and Application media ports to be 0, got %d and %d", ses1.GetMediaFlow(sdp.Video).Port, ses1.GetMediaFlow(sdp.Application).Port)
		}

		ses2, _, err := sdp.ParseString(`v=0
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		if err := ses2.AlignMediaFlows(ses1); err != nil {
			t.Errorf("expected error to be nil, got %s", err)
		}

		for i, media := range ses2.Media {
			if media.Type != ses1.Media[i].Type {
				t.Errorf("expected media type %s, got %s", ses1.Media[i].Type, media.Type)
			}
			if media.Type != sdp.Audio && media.Port != 0 {
				t.Errorf("expected media port to be 0 for %s, got %d", media.Type, media.Port)
			}
		}
	})
}

func TestOrderSDPOffer(t *testing.T) {

	t.Run("Order and Filter SDP Offer", func(t *testing.T) {
		ses, _, err := sdp.ParseString(`v=0
o=- 3849203748 3849203748 IN IP4 192.0.2.1
s=Multimedia Session Example
t=0 0
a=group:BUNDLE audio video data
a=msid-semantic: WMS myStream
c=IN IP4 203.0.113.1
m=audio 51191 RTP/AVP 9 8 0 96 101
a=rtpmap:9 G722/8000/3
a=rtpmap:8 PCMA/8000/2
a=rtpmap:0 PCMU/8000/1
a=rtpmap:96 opus/48000/2
a=fmtp:96 minptime=10;useinbandfec=1
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		mf := ses.GetMediaFlow(sdp.Audio)
		mf.OrderFormatsByName("opus", "pCMu", "telephone-event")

		if len(mf.Formats) != 3 {
			t.Errorf("expected 3 formats, got %d", len(mf.Formats))
		}

		if mf.FormatByName("opus") == nil {
			t.Errorf("expected Opus format to be present, but it was not found")
		}
		if mf.FormatByName("PcmU") == nil {
			t.Errorf("expected PCMU format to be present, but it was not found")
		}
		if mf.FormatByName("telephone-event") == nil {
			t.Errorf("expected Telephone Event format to be absent, but it was found")
		}
	})

	t.Run("Order and Filter SDP Offer with *", func(t *testing.T) {
		ses, _, err := sdp.ParseString(`v=0
o=- 3849203748 3849203748 IN IP4 192.0.2.1
s=Multimedia Session Example
t=0 0
a=group:BUNDLE audio video data
a=msid-semantic: WMS myStream
c=IN IP4 203.0.113.1
m=audio 51191 RTP/AVP 9 8 0 96 101
a=rtpmap:9 G722/8000/3
a=rtpmap:8 PCMA/8000/2
a=rtpmap:0 PCMU/8000/1
a=rtpmap:96 opus/48000/2
a=fmtp:96 minptime=10;useinbandfec=1
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		mf := ses.GetMediaFlow(sdp.Audio)
		mf.OrderFormatsByName("*")

		if len(mf.Formats) != 5 {
			t.Errorf("expected 5 formats, got %d", len(mf.Formats))
		}
	})

	t.Run("Order and Filter SDP Offer with *, G729, PCMA", func(t *testing.T) {
		ses, _, err := sdp.ParseString(`v=0
o=- 3849203748 3849203748 IN IP4 192.0.2.1
s=Multimedia Session Example
t=0 0
a=group:BUNDLE audio video data
a=msid-semantic: WMS myStream
c=IN IP4 203.0.113.1
m=audio 51191 RTP/AVP 9 8 0 96 101
a=rtpmap:9 G722/8000/3
a=rtpmap:8 PCMA/8000/2
a=rtpmap:0 PCMU/8000/1
a=rtpmap:96 opus/48000/2
a=fmtp:96 minptime=10;useinbandfec=1
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		mf := ses.GetMediaFlow(sdp.Audio)
		mf.OrderFormatsByName("*", "G729", "PCMA")

		if len(mf.Formats) != 5 {
			t.Errorf("expected 5 formats, got %d", len(mf.Formats))
		}

		if mf.Formats[4].Name != "PCMA" {
			t.Errorf("expected first format to be PCMA, got %s", mf.Formats[0].Name)
		}
	})

	t.Run("Order and Filter SDP Offer with PCMU, *", func(t *testing.T) {
		ses, _, err := sdp.ParseString(`v=0
o=- 3849203748 3849203748 IN IP4 192.0.2.1
s=Multimedia Session Example
t=0 0
a=group:BUNDLE audio video data
a=msid-semantic: WMS myStream
c=IN IP4 203.0.113.1
m=audio 51191 RTP/AVP 9 8 0 96 101
a=rtpmap:9 G722/8000/3
a=rtpmap:8 PCMA/8000/2
a=rtpmap:0 PCMU/8000/1
a=rtpmap:96 opus/48000/2
a=fmtp:96 minptime=10;useinbandfec=1
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		mf := ses.GetMediaFlow(sdp.Audio)
		mf.OrderFormatsByName("PCMU", "*")

		if len(mf.Formats) != 5 {
			t.Errorf("expected 5 formats, got %d", len(mf.Formats))
		}

		if mf.Formats[0].Name != "PCMU" {
			t.Errorf("expected first format to be PCMU, got %s", mf.Formats[0].Name)
		}
	})

	t.Run("Order and Filter SDP Offer with G722, *, G729", func(t *testing.T) {
		ses, _, err := sdp.ParseString(`v=0
o=- 3849203748 3849203748 IN IP4 192.0.2.1
s=Multimedia Session Example
t=0 0
a=group:BUNDLE audio video data
a=msid-semantic: WMS myStream
c=IN IP4 203.0.113.1
m=audio 49170 RTP/AVP 0 8 9 18 96 97 98 99 100 101 102 103 104
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=rtpmap:9 G722/8000
a=rtpmap:18 G729/8000
a=rtpmap:96 opus/48000/2
a=fmtp:96 minptime=10;useinbandfec=1
a=rtpmap:97 AMR/8000
a=rtpmap:98 AMR-WB/16000
a=rtpmap:99 speex/8000
a=rtpmap:100 iLBC/8000
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=rtpmap:102 GSM/8000
a=rtpmap:103 LPC/8000
a=rtpmap:104 SILK/24000
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
`, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		mf := ses.GetMediaFlow(sdp.Audio)
		mf.OrderFormatsByName("G722", "*", "G729")

		if len(mf.Formats) != 13 {
			t.Errorf("expected 13 formats, got %d", len(mf.Formats))
		}

		if mf.Formats[0].Name != "G722" {
			t.Errorf("expected first format to be G722, got %s", mf.Formats[0].Name)
		}
		if mf.Formats[12].Name != "G729" {
			t.Errorf("expected last format to be G729, got %s", mf.Formats[12].Name)
		}
	})
}

func TestBuildEchoResponderAnswer(t *testing.T) {

	t.Run("Generate SDP Answer for SendRecv", func(t *testing.T) {
		sdpString := `v=0
o=- 3849203748 3849203748 IN IP4 192.0.2.1
s=Multimedia Session Example
c=IN IP4 203.0.113.1
t=0 0
a=group:BUNDLE audio video data
a=msid-semantic: WMS myStream
m=audio 51191 RTP/AVP 9 8 0 96 101
a=rtpmap:9 G722/8000/3
a=rtpmap:8 PCMA/8000/2
a=rtpmap:0 PCMU/8000/1
a=rtpmap:96 opus/48000/2
a=fmtp:96 minptime=10;useinbandfec=1
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=sendrecv
a=mid:audio
a=ssrc:1001 cname:audioCname
m=video 51372 RTP/AVP 97 98
c=IN IP4 203.0.113.3
a=rtpmap:97 H264/90000
a=fmtp:97 profile-level-id=42e01f;packetization-mode=1
a=rtpmap:98 VP8/90000
a=sendrecv
a=mid:video
a=ssrc:1002 cname:videoCname
m=application 50000 DTLS/SCTP 5000
c=IN IP4 203.0.113.4
a=mid:data
a=sctpmap:5000 webrtc-datachannel 1024
`
		ses, _, err := sdp.ParseString(sdpString, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		if answer, _, err := ses.BuildEchoResponderAnswer("PCMA"); err != nil {
			t.Errorf("expected error to be nil, got %s", err)
		} else {
			mf := answer.GetAudioMediaFlow()
			if len(mf.Formats) != 2 {
				t.Errorf("expected 2 audio formats, got %d", len(mf.Formats))
			}
			if mf.Mode != sdp.SendRecv {
				t.Errorf("expected audio mode to be sendrecv, got %s", mf.Mode)
			}
		}
	})

	t.Run("Generate SDP Answer for SendOnly", func(t *testing.T) {
		sdpString := `v=0
o=- 3849203748 3849203748 IN IP4 192.0.2.1
s=Multimedia Session Example
c=IN IP4 203.0.113.1
t=0 0
a=group:BUNDLE audio video data
a=msid-semantic: WMS myStream
m=audio 51191 RTP/AVP 9 8 0 96 101
a=rtpmap:9 G722/8000/3
a=rtpmap:8 PCMA/8000/2
a=rtpmap:0 PCMU/8000/1
a=rtpmap:96 opus/48000/2
a=fmtp:96 minptime=10;useinbandfec=1
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=sendonly
a=mid:audio
a=ssrc:1001 cname:audioCname
m=video 51372 RTP/AVP 97 98
c=IN IP4 203.0.113.3
a=rtpmap:97 H264/90000
a=fmtp:97 profile-level-id=42e01f;packetization-mode=1
a=rtpmap:98 VP8/90000
a=sendrecv
a=mid:video
a=ssrc:1002 cname:videoCname
m=application 50000 DTLS/SCTP 5000
c=IN IP4 203.0.113.4
a=mid:data
a=sctpmap:5000 webrtc-datachannel 1024
`
		ses, _, err := sdp.ParseString(sdpString, false)
		if err != nil {
			t.Fatalf("failed to parse SDP: %v", err)
		}

		if answer, _, err := ses.BuildEchoResponderAnswer("opus"); err != nil {
			t.Errorf("expected error to be nil, got %s", err)
		} else {
			mf := answer.GetAudioMediaFlow()
			if len(mf.Formats) != 2 {
				t.Errorf("expected 2 audio formats, got %d", len(mf.Formats))
			}
			if mf.Mode != sdp.RecvOnly {
				t.Errorf("expected audio mode to be sendrecv, got %s", mf.Mode)
			}
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
		b.ReportAllocs()
		hash1 := hashSDPBytes([]byte(sdp1))
		for b.Loop() {
			bytes2 := []byte(sdp2)
			_ = hash1 == hashSDPBytes(bytes2)
			_ = string(bytes2)
		}
	})

	// b.Run("Equal SDP - One computed", func(b *testing.B) {
	// 	ses1, _, _ := sdp.ParseString(sdp1)
	// 	for b.Loop() {
	// 		ses2, _, _ := sdp.ParseString(sdp2)
	// 		_ = ses1.Equals(ses2)
	// 	}
	// })

	b.Run("Equal SDP - Precomputed", func(b *testing.B) {
		b.ReportAllocs()
		ses1, _, _ := sdp.ParseString(sdp1, false)
		ses2, _, _ := sdp.ParseString(sdp2, false)
		for b.Loop() {
			_ = ses1.Equals(ses2)
			_ = ses2.Bytes()
		}
	})
}

func BenchmarkFormatByName(b *testing.B) {
	sdpstring := `v=0
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

	ses, _, _ := sdp.ParseString(sdpstring, false)
	mf := ses.GetMediaFlow(sdp.Audio)

	b.Run("FormatByToLower", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = mf.FormatByName("PCMA")
		}
	})

	// b.Run("FormatByEqualFold", func(b *testing.B) {
	// 	b.ReportAllocs()
	// 	for b.Loop() {
	// 		_ = mf.FormatByName2("PCMA")
	// 	}
	// })
}

func BenchmarkSDPParseVsClone(b *testing.B) {
	sdpString := `v=0
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

	ses1, _, _ := sdp.ParseString(sdpString, false)
	b.Run("WithParse", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sdp.ParseString(sdpString, false)
		}
	})

	b.Run("WithClone", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = ses1.Clone()
		}
	})
}

func TestParseWebRTCSDP(t *testing.T) {
	sdpString := "v=0\r\no=- 4399166264069674367 2 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0\r\na=extmap-allow-mixed\r\na=msid-semantic: WMS 6573e9d8-9f2e-4feb-b064-13d4650251cf\r\nm=audio 9 UDP/TLS/RTP/SAVPF 111 63 9 0 8 13 110 126\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=candidate:4061950107 1 udp 2113937151 7b0d7f1b-c5b5-49d1-9ac9-44b835a58971.local 64679 typ host generation 0 network-cost 999\r\na=ice-ufrag:Vznr\r\na=ice-pwd:Tt/AbCdcHggQF8RipEHfZg10\r\na=ice-options:trickle\r\na=fingerprint:sha-256 3E:AD:44:E7:0C:B7:25:DE:4F:7E:21:AF:90:CA:BC:5E:66:AB:61:56:FA:BB:16:95:D4:61:CB:4B:F1:BD:4C:8E\r\na=setup:actpass\r\na=mid:0\r\na=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level\r\na=extmap:2 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\r\na=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r\na=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid\r\na=sendrecv\r\na=msid:6573e9d8-9f2e-4feb-b064-13d4650251cf 4dcd226c-03f6-442a-a40b-ba207e835a35\r\na=rtcp-mux\r\na=rtcp-rsize\r\na=rtpmap:111 opus/48000/2\r\na=rtcp-fb:111 transport-cc\r\na=fmtp:111 minptime=10;useinbandfec=1\r\na=rtpmap:63 red/48000/2\r\na=fmtp:63 111/111\r\na=rtpmap:9 G722/8000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:13 CN/8000\r\na=rtpmap:110 telephone-event/48000\r\na=rtpmap:126 telephone-event/8000\r\na=ssrc:397513585 cname:88eQYfCvDAGnLJ+q\r\na=ssrc:397513585 msid:6573e9d8-9f2e-4feb-b064-13d4650251cf 4dcd226c-03f6-442a-a40b-ba207e835a35\r\n"
	ses, _, err := sdp.ParseString(sdpString, false)
	if err != nil {
		t.Fatalf("failed to parse SDP: %v", err)
	}
	fmt.Println(ses.String())
}
func TestParseVoIPSDP(t *testing.T) {
	sdpString := "v=0\r\no=- 4399167 2 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0\r\na=extmap-allow-mixed\r\na=msid-semantic: WMS 6573e9d8-9f2e-4feb-b064-13d4650251cf\r\nm=audio 9 UDP/TLS/RTP/SAVPF 111 63 9 0 8 13 110 126\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=candidate:4061950107 1 udp 2113937151 7b0d7f1b-c5b5-49d1-9ac9-44b835a58971.local 64679 typ host generation 0 network-cost 999\r\na=ice-ufrag:Vznr\r\na=ice-pwd:Tt/AbCdcHggQF8RipEHfZg10\r\na=ice-options:trickle\r\na=fingerprint:sha-256 3E:AD:44:E7:0C:B7:25:DE:4F:7E:21:AF:90:CA:BC:5E:66:AB:61:56:FA:BB:16:95:D4:61:CB:4B:F1:BD:4C:8E\r\na=setup:actpass\r\na=mid:0\r\na=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level\r\na=extmap:2 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\r\na=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r\na=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid\r\na=sendrecv\r\na=msid:6573e9d8-9f2e-4feb-b064-13d4650251cf 4dcd226c-03f6-442a-a40b-ba207e835a35\r\na=rtcp-mux\r\na=rtcp-rsize\r\na=rtpmap:111 opus/48000/2\r\na=rtcp-fb:111 transport-cc\r\na=fmtp:111 minptime=10;useinbandfec=1\r\na=rtpmap:63 red/48000/2\r\na=fmtp:63 111/111\r\na=rtpmap:9 G722/8000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:13 CN/8000\r\na=rtpmap:110 telephone-event/48000\r\na=rtpmap:126 telephone-event/8000\r\na=ssrc:397513585 cname:88eQYfCvDAGnLJ+q\r\na=ssrc:397513585 msid:6573e9d8-9f2e-4feb-b064-13d4650251cf 4dcd226c-03f6-442a-a40b-ba207e835a35\r\n"
	ses, _, err := sdp.ParseString(sdpString, false)
	if err != nil {
		t.Fatalf("failed to parse SDP: %v", err)
	}
	fmt.Println(ses.String())
}
func BenchmarkOrderSDP(b *testing.B) {

	b.Run("My Func", func(b *testing.B) {
		ses, _, _ := sdp.ParseString(`v=0
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
`, false)
		flow := ses.GetMediaFlow(sdp.Audio)
		formats := make([]*sdp.Format, len(flow.Formats))
		copy(formats, flow.Formats)

		for b.Loop() {
			flow := ses.GetMediaFlow(sdp.Audio)
			flow.Formats = make([]*sdp.Format, len(flow.Formats))
			copy(flow.Formats, formats)
			flow.OrderFormatsByName("RED", "PCMU", "telephone-event", "CN", "PCMA")
		}
	})

	// 	b.Run("DS Func", func(b *testing.B) {
	// 		ses, _, _ := sdp.ParseString(`v=0
	// o=- 206 1 IN IP4 192.168.1.101
	// s=session
	// c=IN IP4 192.168.1.5
	// b=CT:1000
	// t=0 0
	// m=audio 51624 RTP/AVP 97 101 13 0 8
	// c=IN IP4 0.0.0.0
	// a=rtcp:51625
	// a=label:Audio
	// a=sendrecv
	// a=rtpmap:97 RED/8000
	// a=rtpmap:101 telephone-event/8000
	// a=fmtp:101 0-16
	// a=rtpmap:13 CN/8000
	// a=rtpmap:0 PCMU/8000
	// a=rtpmap:8 PCMA/8000
	// a=ptime:20
	// `, false)
	// 		flow := ses1.GetMediaFlow(sdp.Audio)
	// 		formats := make([]*sdp.Format, len(flow.Format))
	// 		copy(formats, flow.Format)

	// 		for b.Loop() {
	// 			flow := ses1.GetMediaFlow(sdp.Audio)
	// 			flow.Format = make([]*sdp.Format, len(flow.Format))
	// 			copy(flow.Format, formats)
	// 			flow.OrderFormatsByNameDS("RED", "PCMU", "telephone-event", "CN", "PCMA")
	// 		}
	// 	})

}

func BenchmarkLoweringString(b *testing.B) {
	b.Run("My AsciiToLower", func(b *testing.B) {
		line := `jdsEhdk-FFkEEWXksld`

		for b.Loop() {
			_ = sdp.AsciiToLower(line)
		}
	})

	b.Run("Standard ToLower", func(b *testing.B) {
		line := `jdsEhdk-FFkEEWXksld`

		for b.Loop() {
			_ = strings.ToLower(line)
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

package sdp

import (
	"fmt"
	"strings"
)

const (
	ContentType       = "application/sdp"
	RFC4733           = "telephone-event"
	ComfortNoise      = "CN"
	ComfortNoiseLower = "cn"

	DynamicPayloadStart = 96
)

const (
	PCMU uint8 = 0
	GSM  uint8 = 3
	G723 uint8 = 4
	DVI4 uint8 = 5
	// DVI4   = 6
	LPC   uint8 = 7
	PCMA  uint8 = 8
	G722  uint8 = 9
	L16   uint8 = 11
	QCELP uint8 = 12
	CN    uint8 = 13
	MPA   uint8 = 14
	G728  uint8 = 15
	// DVI4   = 16
	// DVI4   = 17
	G729      uint8 = 18
	Opus      uint8 = 107
	RFC4733PT uint8 = 101

//	0,PCMU,8000,1
//
// 3,GSM,8000,1
// 4,G723,8000,1
// 5,DVI4,8000,1
// 6,DVI4,16000,1
// 7,LPC,8000,1
// 8,PCMA,8000,1
// 9,G722,8000,1
// 11,L16,44100,1
// 12,QCELP,8000,1
// 13,CN,8000,1
// 14,MPA,90000
// 15,G728,8000,1
// 16,DVI4,11025,1
// 17,DVI4,22050,1
// 18,G729,8000,1
)

var (
	SupportedCodecsStringList = []string{"PCMA", "PCMU", "G722", "opus"} // "G729"
)

// var mapCodecs = map[uint8]string{
// 	PCMU:      "PCMU",
// 	GSM:       "GSM",
// 	G723:      "G723",
// 	DVI4:      "DVI4",
// 	LPC:       "LPC",
// 	PCMA:      "PCMA",
// 	G722:      "G722",
// 	L16:       "L16",
// 	QCELP:     "QCELP",
// 	CN:        "CN",
// 	MPA:       "MPA",
// 	G728:      "G728",
// 	G729:      "G729",
// 	Opus:      "opus",
// 	RFC4733PT: "telephone-event",
// }

var mapCodecs = map[uint8]string{
	0:   "PCMU",
	3:   "GSM",
	4:   "G723",
	5:   "DVI4",
	6:   "DVI4",
	7:   "LPC",
	8:   "PCMA",
	9:   "G722",
	10:  "L16",
	11:  "L16",
	12:  "QCELP",
	13:  "CN",
	14:  "MPA",
	15:  "G728",
	16:  "DVI4",
	17:  "DVI4",
	18:  "G729",
	25:  "CelB",
	26:  "JPEG",
	31:  "H261",
	32:  "MPV",
	33:  "MP2T",
	34:  "H263",
	101: "telephone-event",
	107: "opus",
}

type CodecFamily string
type MediaUse string

const (
	// Codec families
	FamilyWaveform  CodecFamily = "waveform"  // PCM / ADPCM waveform (PCMU, PCMA, L16, G.722, G.726)
	FamilyCELP      CodecFamily = "celp"      // G.729, G.723.1, iLBC, AMR, AMR-WB
	FamilyVocoder   CodecFamily = "vocoder"   // LPC, MELP, GSM-FR (legacy speech model)
	FamilyTransform CodecFamily = "transform" // MP3/AAC/CELT; video hybrids (H264/VP8) are listed as transform here
	FamilyLossless  CodecFamily = "lossless"
	FamilyHybrid    CodecFamily = "hybrid" // Mixed (e.g., Opus = SILK+CELT)
	FamilyOther     CodecFamily = "other"

	// Media uses
	UseAudio MediaUse = "audio"
	UseVideo MediaUse = "video"
	UseDTMF  MediaUse = "dtmf"
	UseCN    MediaUse = "comfort-noise"
)

type CodecInfo struct {
	PayloadType uint8
	Name        string
	ClockRate   int // Hz (audio) or RTP clock (video; usually 90000)
	Channels    int // 0 for video / not-applicable
	Use         MediaUse
	Family      CodecFamily
}

// --- Static RTP payload types (RFC 3551) ---
var codecsInfoMap = map[uint8]CodecInfo{
	0:  {0, "PCMU", 8000, 1, UseAudio, FamilyWaveform},
	1:  {1, "Reserved", 8000, 1, UseAudio, FamilyOther},
	2:  {2, "G726-32", 8000, 1, UseAudio, FamilyWaveform},
	3:  {3, "GSM", 8000, 1, UseAudio, FamilyVocoder},
	4:  {4, "G723", 8000, 1, UseAudio, FamilyCELP},
	5:  {5, "DVI4", 8000, 1, UseAudio, FamilyWaveform},
	6:  {6, "DVI4", 16000, 1, UseAudio, FamilyWaveform},
	7:  {7, "LPC", 8000, 1, UseAudio, FamilyVocoder},
	8:  {8, "PCMA", 8000, 1, UseAudio, FamilyWaveform},
	9:  {9, "G722", 8000, 1, UseAudio, FamilyWaveform}, // RTP clock remains 8 kHz, but it is actually 16 kHz
	10: {10, "L16", 44100, 2, UseAudio, FamilyWaveform},
	11: {11, "L16", 44100, 1, UseAudio, FamilyWaveform},
	12: {12, "QCELP", 8000, 1, UseAudio, FamilyCELP},
	13: {13, "CN", 8000, 1, UseCN, FamilyOther},
	14: {14, "MPA", 90000, 2, UseAudio, FamilyTransform}, // MPEG audio
	15: {15, "G728", 8000, 1, UseAudio, FamilyCELP},
	16: {16, "DVI4", 11025, 1, UseAudio, FamilyWaveform},
	17: {17, "DVI4", 22050, 1, UseAudio, FamilyWaveform},
	18: {18, "G729", 8000, 1, UseAudio, FamilyCELP},
	25: {25, "CelB", 90000, 0, UseVideo, FamilyOther},
	26: {26, "JPEG", 90000, 0, UseVideo, FamilyOther},
	28: {28, "NV", 90000, 0, UseVideo, FamilyOther},
	31: {31, "H261", 90000, 0, UseVideo, FamilyOther},
	32: {32, "MPV", 90000, 0, UseVideo, FamilyTransform},  // MPEG-1/2 Video
	33: {33, "MP2T", 90000, 0, UseVideo, FamilyTransform}, // MPEG-2 TS
	34: {34, "H263", 90000, 0, UseVideo, FamilyOther},

	// --- Common dynamic payloads (typical assignments; negotiated in SDP) ---
	97:  {97, "AMR", 8000, 1, UseAudio, FamilyCELP},
	98:  {98, "AMR-WB", 16000, 1, UseAudio, FamilyCELP},
	99:  {99, "iLBC", 8000, 1, UseAudio, FamilyCELP},
	100: {100, "VP8", 90000, 0, UseVideo, FamilyTransform},
	101: {101, "VP9", 90000, 0, UseVideo, FamilyTransform},
	102: {102, "H264", 90000, 0, UseVideo, FamilyTransform},
	103: {103, "H265", 90000, 0, UseVideo, FamilyTransform},
	104: {104, "AV1", 90000, 0, UseVideo, FamilyTransform},
	105: {105, "telephone-event", 8000, 1, UseDTMF, FamilyOther}, // RFC 4733 (often PT=101 as well)
	// 106: {106, "CN", 16000, 1, UseCN, FamilyOther},               // WB CN
	// 107: {107, "CN", 32000, 1, UseCN, FamilyOther},               // SWB CN
	// 108: {108, "CN", 48000, 1, UseCN, FamilyOther},               // FB CN
	107: {107, "opus", 48000, 2, UseAudio, FamilyHybrid}, // SILK (CELP) + CELT (transform)
}

// --- AMR / AMR-WB frame size tables (bytes per 20 ms frame, speech class A/B/C only; no SID) ---
var amrFrameSizes = map[float64]int{
	4.75: 12,
	5.15: 13,
	5.90: 15,
	6.70: 17,
	7.40: 19,
	7.95: 20,
	10.2: 27,
	12.2: 31,
}

var amrWBFrameSizes = map[float64]int{
	6.60:  17,
	8.85:  23,
	12.65: 32,
	14.25: 36,
	15.85: 40,
	18.25: 46,
	19.85: 50,
	23.05: 58,
	23.85: 60,
}

// FrameSize returns the RTP payload bytes for one packet of codec c,
// given a frame duration in milliseconds (frameDurationMs).
// For multi-mode codecs (AMR/AMR-WB/G.723.1), supply modeKbps to get exact sizes.
// For variable-size codecs (e.g., Opus) or video, it returns 0 with no error (unknown).
//
// modeKbps usage:
//   - G.723: 6.3 or 5.3 (kbps)
//   - AMR: one of {4.75, 5.15, 5.90, 6.70, 7.40, 7.95, 10.2, 12.2}
//   - AMR-WB: one of {6.60, 8.85, 12.65, 14.25, 15.85, 18.25, 19.85, 23.05, 23.85}
func FrameSize(c CodecInfo, frameDurationMs int, modeKbps float64) (int, error) {
	switch c.Name {
	// --- Waveform codecs ---
	case "PCMU", "PCMA": // 8-bit log PCM @ 8 kHz
		samples := (c.ClockRate * frameDurationMs / 1000) * c.Channels
		return samples, nil // 1 byte per sample

	case "L16": // 16-bit linear PCM
		samples := (c.ClockRate * frameDurationMs / 1000) * c.Channels
		return samples * 2, nil

	case "G722": // 64 kbps ADPCM; RTP clock is 8 kHz; 1 byte per (8 kHz) sample per channel
		samples := (8000 * frameDurationMs / 1000) * c.Channels
		return samples, nil

	case "G726-32":
		// Classic ADPCM @ 32 kbps → 4 bits/sample (0.5 bytes)
		// Samples at 8 kHz → bytes = 8000 * dur * 0.5 * channels
		samples := (c.ClockRate * frameDurationMs / 1000) * c.Channels
		return samples / 2, nil

	// --- Fixed-size CELP codecs ---
	case "G729":
		// 10 bytes per 10 ms speech frame @ 8 kHz
		if frameDurationMs%10 != 0 {
			return 0, fmt.Errorf("G.729 requires 10 ms multiples")
		}
		return (frameDurationMs / 10) * 10, nil

	case "G728":
		// 16 kbps LD-CELP. Base frame 2.5 ms = 5 bytes. So:
		// bytes = (frameDurationMs / 2.5) * 5.
		// Integer math: require frameDurationMs * 2 % 5 == 0
		if (frameDurationMs*2)%5 != 0 {
			return 0, fmt.Errorf("G.728 requires multiples of 2.5 ms")
		}
		frames := (frameDurationMs * 2) / 5 // number of 2.5 ms frames
		return frames * 5, nil              // 5 bytes per 2.5 ms → 20 bytes/10 ms

	case "iLBC":
		// 20 ms = 38 bytes; 30 ms = 50 bytes
		switch frameDurationMs {
		case 20:
			return 38, nil
		case 30:
			return 50, nil
		default:
			return 0, fmt.Errorf("iLBC supports 20 or 30 ms only")
		}

	case "G723":
		// Two modes: 6.3 kbps = 24 bytes/30 ms, 5.3 kbps = 20 bytes/30 ms
		if frameDurationMs%30 != 0 {
			return 0, fmt.Errorf("G.723 requires 30 ms multiples")
		}
		switch modeKbps {
		case 6.3:
			return (frameDurationMs / 30) * 24, nil
		case 5.3:
			return (frameDurationMs / 30) * 20, nil
		default:
			return 0, fmt.Errorf("G.723 mode must be 6.3 or 5.3 kbps")
		}

	// --- AMR family (sizes per mode, 20 ms only) ---
	case "AMR":
		if frameDurationMs != 20 {
			return 0, fmt.Errorf("AMR supports 20 ms frames only")
		}
		if sz, ok := amrFrameSizes[modeKbps]; ok {
			return sz, nil
		}
		return 0, fmt.Errorf("unsupported AMR mode %.2f kbps", modeKbps)

	case "AMR-WB":
		if frameDurationMs != 20 {
			return 0, fmt.Errorf("AMR-WB supports 20 ms frames only")
		}
		if sz, ok := amrWBFrameSizes[modeKbps]; ok {
			return sz, nil
		}
		return 0, fmt.Errorf("unsupported AMR-WB mode %.2f kbps", modeKbps)

	case "Opus":
		// Valid durations: 2.5, 5, 10, 20, 40, 60 ms
		switch frameDurationMs {
		case 2, 5, 10, 20, 40, 60:
			// Opus payload size varies with target bitrate/FEC/complexity.
			return 0, nil
		default:
			return 0, fmt.Errorf("OPUS supports 2, 5, 10, 20, 40, or 60 ms") // using 2 instead of 2.5 due to int ms granularity
		}

	// --- Comfort Noise (RFC 3389) ---
	case "CN":
		// Typical SID payload sizes (approximate/common):
		switch c.ClockRate {
		case 8000:
			return 2, nil // narrowband
		case 16000:
			return 3, nil // wideband
		case 32000:
			return 6, nil // super-wideband
		case 48000:
			return 6, nil // fullband
		default:
			return 0, fmt.Errorf("unsupported CN clock rate %d", c.ClockRate)
		}

	// --- DTMF (telephone-event, RFC 4733) ---
	case "telephone-event":
		// A single DTMF event report is 4 bytes; senders may repeat/extend,
		// but per packet payload is typically 4 bytes.
		return 4, nil

	// Video & others: cannot know without encoding
	case "H261", "H263", "H264", "H265", "VP8", "VP9", "AV1", "JPEG", "CelB", "MPV", "MP2T", "NV":
		return 0, nil

	default:
		return 0, fmt.Errorf("frame size not implemented for codec %s", c.Name)
	}
}

// get canonical codec name and its type
func IdentifyPayloadTypeByName(name string) (CodecInfo, bool) {
	var codecInfo CodecInfo
	for _, ci := range codecsInfoMap {
		if strings.EqualFold(ci.Name, name) {
			return ci, true
		}
	}
	return codecInfo, false
}

func GetCodecNames(payloadType ...uint8) []string {
	if len(payloadType) == 0 {
		return nil
	}
	codecs := make([]string, len(payloadType))
	for i, pt := range payloadType {
		if name, ok := mapCodecs[pt]; ok {
			codecs[i] = name
		} else {
			codecs[i] = "Unknown"
		}
	}
	return codecs
}

func GetCodecName(pt uint8) string {
	if name, ok := mapCodecs[pt]; ok {
		return name
	}
	return "Unknown"
}

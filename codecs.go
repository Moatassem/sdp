package sdp

import "strings"

const (
	ContentType = "application/sdp"
	RFC4733     = "telephone-event"
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
	Opus      uint8 = 96
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

type CodecType = string

const (
	CodecTypeAudio CodecType = "audio"
	CodecTypeVideo CodecType = "video"
	CodecTypeDTMF  CodecType = "dtmf"
)

var (
	SupportedCodecsUint8List  = []uint8{PCMA, PCMU, G722, G729, Opus}
	SupportedCodecsStringList = []string{"PCMA", "PCMU", "G722", "G729", "opus"}
)

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

type CodecInfo struct {
	Payload      uint8
	Name         string
	Channels     int
	SamplingRate int
	Type         string // "audio" or "video" or "dtmf"
}

var (
	mapCodecs = map[uint8]string{
		PCMU:      "PCMU",
		GSM:       "GSM",
		G723:      "G723",
		DVI4:      "DVI4",
		LPC:       "LPC",
		PCMA:      "PCMA",
		G722:      "G722",
		L16:       "L16",
		QCELP:     "QCELP",
		CN:        "CN",
		MPA:       "MPA",
		G728:      "G728",
		G729:      "G729",
		Opus:      "opus",
		RFC4733PT: "telephone-event",
	}

	codecsInfoMap = map[uint8]CodecInfo{
		// Static audio codecs
		0:  {0, "PCMU", 1, 8000, "audio"},
		3:  {3, "GSM", 1, 8000, "audio"},
		4:  {4, "G723", 1, 8000, "audio"},
		5:  {5, "DVI4", 1, 8000, "audio"},
		6:  {6, "DVI4", 1, 16000, "audio"},
		7:  {7, "LPC", 1, 8000, "audio"},
		8:  {8, "PCMA", 1, 8000, "audio"},
		9:  {9, "G722", 1, 8000, "audio"},
		10: {10, "L16", 2, 44100, "audio"},
		11: {11, "L16", 1, 44100, "audio"},
		12: {12, "QCELP", 1, 8000, "audio"},
		13: {13, "CN", 1, 8000, "audio"},
		14: {14, "MPA", 2, 90000, "audio"},
		15: {15, "G728", 1, 8000, "audio"},
		16: {16, "DVI4", 1, 11025, "audio"},
		17: {17, "DVI4", 1, 22050, "audio"},
		18: {18, "G729", 1, 8000, "audio"},

		// Static video codecs
		25: {25, "CelB", 0, 90000, "video"},
		26: {26, "JPEG", 0, 90000, "video"},
		28: {28, "nv", 0, 90000, "video"},
		31: {31, "H261", 0, 90000, "video"},
		32: {32, "MPV", 0, 90000, "video"},
		33: {33, "MP2T", 0, 90000, "video"},
		34: {34, "H263", 0, 90000, "video"},

		// Dynamic audio codecs (common assignments)
		96:  {96, "Opus", 2, 48000, "audio"},
		97:  {97, "AMR", 1, 8000, "audio"},
		98:  {98, "iLBC", 1, 8000, "audio"},
		99:  {99, "G726-32", 1, 8000, "audio"},
		100: {100, "AAC", 2, 48000, "audio"},

		// Dynamic video codecs (common assignments)
		101: {101, "VP8", 0, 90000, "video"},
		102: {102, "VP9", 0, 90000, "video"},
		103: {103, "H264", 0, 90000, "video"},
		104: {104, "H265", 0, 90000, "video"},
		105: {105, "AV1", 0, 90000, "video"},
	}
)

// get canonical codec name and its type
func IdentifyPayloadTypeByName(name string) (uint8, string, bool) {
	if AsciiToLower(name) == RFC4733 {
		return 101, CodecTypeDTMF, true
	}
	for payload, ci := range codecsInfoMap {
		if strings.EqualFold(ci.Name, name) {
			return payload, ci.Type, true
		}
	}
	return 0, "", false
}

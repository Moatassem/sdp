package sdp

type Codec struct {
	Payload  uint8
	Name     string
	Channels int
	Bitrate  int
}

var (
	SupportedCodecsUint8List  = []uint8{PCMA, PCMU, G722, G729}
	SupportedCodecsStringList = []string{"PCMA", "PCMU", "G722", "G729"}
	mapCodecs                 = map[uint8]string{
		PCMU:  "PCMU",
		GSM:   "GSM",
		G723:  "G723",
		DVI4:  "DVI4",
		LPC:   "LPC",
		PCMA:  "PCMA",
		G722:  "G722",
		L16:   "L16",
		QCELP: "QCELP",
		CN:    "CN",
		MPA:   "MPA",
		G728:  "G728",
		G729:  "G729",
	}
)

const (
	RFC4733     = "telephone-event"
	ContentType = "application/sdp"
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

func GetCodecName(payloadType ...uint8) []string {
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

package sdp

import (
	"strings"
	"time"
)

func init() {
	mapCodecs = map[uint8]string{
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
	}
}

// Session represents an SDP session description.
type Session struct {
	Version     int          // Protocol Version ("v=")
	Origin      *Origin      // Origin ("o=")
	Name        string       // Session Name ("s=")
	Information string       // Session Information ("i=")
	URI         string       // URI ("u=")
	Email       []string     // Email Address ("e=")
	Phone       []string     // Phone Number ("p=")
	Connection  *Connection  // Connection Data ("c=")
	Bandwidth   []*Bandwidth // Bandwidth ("b=")
	TimeZone    []*TimeZone  // TimeZone ("z=")
	Key         []*Key       // Encryption Keys ("k=")
	Timing      *Timing      // Timing ("t=")
	Repeat      []*Repeat    // Repeat Times ("r=")
	Attributes  Attributes   // Session Attributes ("a=")
	Mode        string       // Streaming mode ("sendrecv", "recvonly", "sendonly", or "inactive")
	Media       []*Media     // Media Descriptions ("m=")
}

// String returns the encoded session description as string.
func (s *Session) String() string {
	return string(s.Bytes())
}

// Bytes returns the encoded session description as buffer.
func (s *Session) Bytes() []byte {
	e := NewEncoder(nil)
	e.Encode(s)
	return e.Bytes()
}

func NewSessionSDP(sesID, sesVer int64, ipv4, nm, ssrc, mdir string, port int, codecs []uint8) *Session {
	formats := make([]*Format, 0, len(codecs))
	for _, codec := range codecs {
		formats = append(formats, getFormat(uint8(codec)))
	}

	return &Session{
		Version: 0,
		Origin: &Origin{
			Username:       "-",
			SessionID:      sesID,
			SessionVersion: sesVer,
			Network:        NetworkInternet,
			Type:           TypeIPv4,
			Address:        ipv4,
		},
		Name: nm,
		Connection: &Connection{
			Network: NetworkInternet,
			Type:    TypeIPv4,
			Address: ipv4,
		},
		Media: []*Media{
			{
				Chosen:     true,
				Type:       Audio,
				Port:       port,
				Proto:      RtpAvp,
				Attributes: Attributes{{Name: "ssrc", Value: ssrc}},
				Mode:       mdir,
				Format:     formats,
			},
		},
	}

}

func getFormat(codec uint8) *Format {
	frmt := &Format{
		Payload:   codec,
		Name:      mapCodecs[codec],
		ClockRate: 8000,
		Channels:  1,
	}
	if codec == Telephone_Event {
		frmt.Params = append(frmt.Params, "0-16")
	}
	return frmt
}

// Get Chosen Media Description
func (s *Session) GetChosenMedia() *Media {
	for _, m := range s.Media {
		if m.Chosen {
			return m
		}
	}
	return nil
}

func (s *Session) GetEffectivePTime() string {
	media := s.GetChosenMedia()
	attrbnm := "ptime"
	ptime := media.Attributes.Get(attrbnm)
	if ptime != "" {
		return ptime
	}
	ptime = s.Attributes.Get(attrbnm)
	if ptime != "" {
		return ptime
	}
	return "20"
}

// func (attrbs Attributes) GetAttributeValue(nm string) string {
// 	for i := 0; i < len(attrbs); i++ {
// 		attrb := attrbs[i]
// 		if attrb.Name == "ptime" {
// 			return attrb.Value
// 		}
// 	}
// 	return ""
// }

func (s *Session) GetEffectiveConnection(media *Media) string {
	var ipv4 string
	for i := range media.Connection {
		ipv4 = media.Connection[i].Address
		if ipv4 != "" {
			return ipv4
		}
	}
	return s.Connection.Address
}

func (s *Session) IsT38Image() bool {
	for i := range s.Media {
		media := s.Media[i]
		if media.Type == Image && media.Port > 0 && media.Proto == Udptl && media.FormatDescr == "t38" {
			return true
		}
	}
	return false
}

func (s *Session) GetEffectiveMediaDirective() string {
	media := s.GetChosenMedia()
	if media.Mode != "" {
		return media.Mode
	}
	if s.Mode != "" {
		return s.Mode
	}
	return SendRecv
}

func (s *Session) IsCallHeld() bool {
	media := s.GetChosenMedia()
	var mode string
	if media.Mode != "" {
		mode = media.Mode
	} else if s.Mode != "" {
		mode = s.Mode
	} else {
		mode = SendRecv
	}
	if mode == SendOnly || mode == Inactive {
		return true
	}
	if ipv4 := s.GetEffectiveConnection(media); ipv4 == "" || ipv4 == "0.0.0.0" {
		return true
	}
	return false
}

// Origin represents an originator of the session.
type Origin struct {
	Username       string
	SessionID      int64
	SessionVersion int64
	Network        string
	Type           string
	Address        string
}

const (
	NetworkInternet = "IN"
)

const (
	TypeIPv4 = "IP4"
	TypeIPv6 = "IP6"
)

// Connection contains connection data.
type Connection struct {
	Network    string
	Type       string
	Address    string
	TTL        int
	AddressNum int
}

// Bandwidth contains session or media bandwidth information.
type Bandwidth struct {
	Type  string
	Value int
}

// TimeZone represents a time zones change information for a repeated session.
type TimeZone struct {
	Time   time.Time
	Offset time.Duration
}

// Key contains a key exchange information.
// Deprecated. Use for backwards compatibility only.
type Key struct {
	Method, Value string
}

// Timing specifies start and stop times for a session.
type Timing struct {
	Start time.Time
	Stop  time.Time
}

// Repeat specifies repeat times for a session.
type Repeat struct {
	Interval time.Duration
	Duration time.Duration
	Offsets  []time.Duration
}

// Media contains media description.
type Media struct {
	Chosen      bool
	Type        string
	Port        int
	PortNum     int
	Proto       string
	Information string        // Media Information ("i=")
	Connection  []*Connection // Connection Data ("c=")
	Bandwidth   []*Bandwidth  // Bandwidth ("b=")
	Key         []*Key        // Encryption Keys ("k=")
	Attributes                // Attributes ("a=")
	Mode        string        // Streaming mode ("sendrecv", "recvonly", "sendonly", or "inactive")
	Format      []*Format     // Media Format for RTP/AVP or RTP/SAVP protocols ("rtpmap", "fmtp", "rtcp-fb")
	FormatDescr string        // Media Format for other protocols
}

// Streaming modes.
const (
	SendRecv = "sendrecv"
	SendOnly = "sendonly"
	RecvOnly = "recvonly"
	Inactive = "inactive"

	Audio       = "audio"       //[RFC8866]
	Video       = "video"       //[RFC8866]
	Text        = "text"        //[RFC8866]
	Application = "application" //[RFC8866]
	Message     = "message"     //[RFC8866]
	Image       = "image"       //[RFC6466]

	RtpAvp            = "RTP/AVP"               // [RFC8866]
	Udp               = "udp"                   // [RFC8866]
	Vat               = "vat"                   // [1]
	Rtp               = "rtp"                   // [1]
	Udptl             = "udptl"                 // [ITU-T Recommendation T.38, 'Procedures for real-time Group 3 facsimile communication over IP networks', June 1998. (Section 9)]
	Tcp               = "TCP"                   // [RFC4145]
	RtpAvpf           = "RTP/AVPF"              // [RFC4585]
	TcpRtpAvp         = "TCP/RTP/AVP"           // [RFC4571]
	RtpSavp           = "RTP/SAVP"              // [RFC3711]
	TcpBfcp           = "TCP/BFCP"              // [RFC8856]
	TcpTlsBfcp        = "TCP/TLS/BFCP"          // [RFC8856]
	TcpTls            = "TCP/TLS"               // [RFC8122]
	FluteUdp          = "FLUTE/UDP"             // [RFC-mehta-rmt-flute-sdp-05]
	TcpMsrp           = "TCP/MSRP"              // [RFC4975]
	TcpTlsMsrp        = "TCP/TLS/MSRP"          // [RFC4975]
	Dccp              = "DCCP"                  // [RFC5762]
	DccpRtpAvp        = "DCCP/RTP/AVP"          // [RFC5762]
	DccpRtpSavp       = "DCCP/RTP/SAVP"         // [RFC5762]
	DccpRtpAvpf       = "DCCP/RTP/AVPF"         // [RFC5762]
	DccpRtpSavpf      = "DCCP/RTP/SAVPF"        // [RFC5762]
	RtpSavpf          = "RTP/SAVPF"             // [RFC5124]
	UdpTlsRtpSavp     = "UDP/TLS/RTP/SAVP"      // [RFC5764]
	DccpTlsRtpSavp    = "DCCP/TLS/RTP/SAVP"     // [RFC5764]
	UdpTlsRtpSavpf    = "UDP/TLS/RTP/SAVPF"     // [RFC5764]
	DccpTlsRtpSavpf   = "DCCP/TLS/RTP/SAVPF"    // [RFC5764]
	UdpMbmsFecRtpAvp  = "UDP/MBMS-FEC/RTP/AVP"  // [RFC6064]
	UdpMbmsFecRtpSavp = "UDP/MBMS-FEC/RTP/SAVP" // [RFC6064]
	UdpMbmsRepair     = "UDP/MBMS-REPAIR"       // [RFC6064]
	FecUdp            = "FEC/UDP"               // [RFC6364]
	UdpFec            = "UDP/FEC"               // [RFC6364]
	TcpMrcpv2         = "TCP/MRCPv2"            // [RFC6787]
	TcpTlsMrcpv2      = "TCP/TLS/MRCPv2"        // [RFC6787]
	Pstn              = "PSTN"                  // [RFC7195]
	UdpTlsUdptl       = "UDP/TLS/UDPTL"         // [RFC7345]
	Voice             = "voice"                 // [RFC2848]
	Fax               = "fax"                   // [RFC2848]
	Pager             = "pager"                 // [RFC2848]
	TcpRtpAvpf        = "TCP/RTP/AVPF"          // [RFC7850]
	TcpRtpSavp        = "TCP/RTP/SAVP"          // [RFC7850]
	TcpRtpSavpf       = "TCP/RTP/SAVPF"         // [RFC7850]
	TcpDtlsRtpSavp    = "TCP/DTLS/RTP/SAVP"     // [RFC7850]
	TcpDtlsRtpSavpf   = "TCP/DTLS/RTP/SAVPF"    // [RFC7850]
	TcpTlsRtpAvp      = "TCP/TLS/RTP/AVP"       // [RFC7850]
	TcpTlsRtpAvpf     = "TCP/TLS/RTP/AVPF"      // [RFC7850]
	TcpWsBfcp         = "TCP/WS/BFCP"           // [RFC8857]
	TcpWssBfcp        = "TCP/WSS/BFCP"          // [RFC8857]
	UdpDtlsSctp       = "UDP/DTLS/SCTP"         // [RFC8841]
	TcpDtlsSctp       = "TCP/DTLS/SCTP"         // [RFC8841]
	TcpDtlsBfcp       = "TCP/DTLS/BFCP"         // [RFC8856]
	UdpBfcp           = "UDP/BFCP"              // [RFC8856]
	UdpTlsBfcp        = "UDP/TLS/BFCP"          // [RFC8856]

)

func GetOriginatingMode(local string, putCallonHold bool) (string, bool) {
	if putCallonHold {
		switch local {
		case "", SendRecv:
			local = SendOnly
		case RecvOnly:
			local = Inactive
		default:
			return "", false
		}
	} else {
		switch local {
		case SendOnly:
			local = SendRecv
		case Inactive:
			local = RecvOnly
		case "":
			local = SendRecv
		default:
			return "", false
		}
	}
	return local, true
}

// NegotiateAnswerMode negotiates streaming mode.
func NegotiateAnswerMode(local, remote string) string {
	switch local {
	case "", SendRecv:
		switch remote {
		case RecvOnly:
			return SendOnly
		case SendOnly:
			return RecvOnly
		default:
			return remote
		}
	case Inactive, SendOnly:
		switch remote {
		case SendRecv, RecvOnly:
			return SendOnly
		}
	case RecvOnly:
		switch remote {
		case SendOnly:
			return RecvOnly
		default:
			return remote
		}
	}
	return Inactive
}

func IsMedDirHolding(md string) bool {
	return md == SendOnly || md == Inactive
}

// DeleteAttr removes all elements with name from attrs.
func DeleteAttr(attrs Attributes, name ...string) Attributes {
	n := 0
loop:
	for _, it := range attrs {
		for _, v := range name {
			if it.Name == v {
				continue loop
			}
		}
		attrs[n] = it
		n++
	}
	return attrs[:n]
}

// FormatByPayload returns format description by payload type.
func (m *Media) FormatByPayload(payload uint8) *Format {
	for _, f := range m.Format {
		if f.Payload == payload {
			return f
		}
	}
	return nil
}

// Format is a media format description represented by "rtpmap" attributes.
type Format struct {
	Payload   uint8
	Name      string
	ClockRate int
	Channels  int
	Feedback  []string // "rtcp-fb" attributes
	Params    []string // "fmtp" attributes
}

func (f *Format) String() string {
	return f.Name
}

var epoch = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)

func isRTP(media, proto string) bool {
	switch media {
	case "audio", "video":
		return strings.Contains(proto, "RTP/AVP") || strings.Contains(proto, "RTP/SAVP")
	default:
		return false
	}
}

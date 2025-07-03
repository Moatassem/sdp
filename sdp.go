package sdp

import (
	"cmp"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"
)

var epoch = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)

const DeltaRune rune = 'a' - 'A'

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

func (s *Session) Clone() *Session {
	if s == nil {
		return nil
	}

	clone := &Session{
		Version:     s.Version,
		Name:        s.Name,
		Information: s.Information,
		URI:         s.URI,
		Mode:        s.Mode,
	}

	// Clone Origin
	if s.Origin != nil {
		clone.Origin = &Origin{
			Username:       s.Origin.Username,
			SessionID:      s.Origin.SessionID,
			SessionVersion: s.Origin.SessionVersion,
			Network:        s.Origin.Network,
			Type:           s.Origin.Type,
			Address:        s.Origin.Address,
		}
	}

	// Clone Email slice
	if s.Email != nil {
		clone.Email = make([]string, len(s.Email))
		copy(clone.Email, s.Email)
	}

	// Clone Phone slice
	if s.Phone != nil {
		clone.Phone = make([]string, len(s.Phone))
		copy(clone.Phone, s.Phone)
	}

	// Clone Connection
	if s.Connection != nil {
		clone.Connection = &Connection{
			Network:    s.Connection.Network,
			Type:       s.Connection.Type,
			Address:    s.Connection.Address,
			TTL:        s.Connection.TTL,
			AddressNum: s.Connection.AddressNum,
		}
	}

	// Clone Bandwidth slice
	if s.Bandwidth != nil {
		clone.Bandwidth = make([]*Bandwidth, len(s.Bandwidth))
		for i, bw := range s.Bandwidth {
			if bw != nil {
				clone.Bandwidth[i] = &Bandwidth{
					Type:  bw.Type,
					Value: bw.Value,
				}
			}
		}
	}

	// Clone TimeZone slice
	if s.TimeZone != nil {
		clone.TimeZone = make([]*TimeZone, len(s.TimeZone))
		for i, tz := range s.TimeZone {
			if tz != nil {
				clone.TimeZone[i] = &TimeZone{
					Time:   tz.Time,
					Offset: tz.Offset,
				}
			}
		}
	}

	// Clone Key slice
	if s.Key != nil {
		clone.Key = make([]*Key, len(s.Key))
		for i, k := range s.Key {
			if k != nil {
				clone.Key[i] = &Key{
					Method: k.Method,
					Value:  k.Value,
				}
			}
		}
	}

	// Clone Timing
	if s.Timing != nil {
		clone.Timing = &Timing{
			Start: s.Timing.Start,
			Stop:  s.Timing.Stop,
		}
	}

	// Clone Repeat slice
	if s.Repeat != nil {
		clone.Repeat = make([]*Repeat, len(s.Repeat))
		for i, r := range s.Repeat {
			if r != nil {
				offsets := make([]time.Duration, len(r.Offsets))
				copy(offsets, r.Offsets)
				clone.Repeat[i] = &Repeat{
					Interval: r.Interval,
					Duration: r.Duration,
					Offsets:  offsets,
				}
			}
		}
	}

	// Clone Attributes
	clone.Attributes = s.Attributes.clone()

	// Clone Media slice
	if s.Media != nil {
		clone.Media = make([]*Media, len(s.Media))
		for i, m := range s.Media {
			if m == nil {
				continue
			}
			clone.Media[i] = m.clone(-1)
		}
	}

	return clone
}

// String returns the encoded session description as string.
func (ses *Session) String() string {
	return string(ses.Bytes())
}

// Bytes returns the encoded session description as bytes.
func (ses *Session) Bytes() []byte {
	e := NewEncoder(nil)
	e.Encode(ses)
	return e.Bytes()
}

func (ses *Session) BuildEchoResponderAnswer(audiofrmts ...string) (*Session, bool, error) {
	if len(audiofrmts) == 0 {
		return nil, false, fmt.Errorf("Cannot build EchoResponder answer: no audio formats provided")
	}

	if mf := ses.GetAudioMediaFlow(); mf == nil {
		return nil, false, fmt.Errorf("Cannot build EchoResponder answer: no audio media flow found")
	} else {
		if mf.Port == 0 {
			return nil, false, fmt.Errorf("Cannot build EchoResponder answer: audio media flow is disabled")
		}
	}
	answer := ses.Clone()
	answer.Name = "EchoResponder"
	answer.Mode = ""
	answer.Origin.SessionVersion = 1

	mf := answer.DisableFlowsExcept(Audio).GetAudioMediaFlow()

	audiofound, dtmffound := mf.KeepOnlyFirstAudioCodecAlongRFC4733(audiofrmts...)

	if !audiofound {
		return nil, false, fmt.Errorf("Cannot build EchoResponder answer: no common audio formats found in audio media flow")
	}

	mf.Mode = NegotiateAnswerMode(SendRecv, ses.GetEffectiveMediaDirective())

	return answer, dtmffound, nil
}

func NewSessionSDP(sesID, sesVer int64, ipv4, nm, ssrc, mdir string, port int, codecs []uint8) (*Session, error) {
	formats := make([]*Format, 0, len(codecs))
	for _, codec := range codecs {
		frmt, err := buildFormat(codec)
		if err != nil {
			return nil, err
		}
		formats = append(formats, frmt)
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
				Type:  Audio,
				Port:  port,
				Proto: RtpAvp,
				Attributes: func() Attributes {
					if ssrc != "" {
						return Attributes{{Name: "ssrc", Value: ssrc}}
					}
					return nil
				}(),
				Mode:    mdir,
				Formats: formats,
			},
		},
	}, nil
}

func buildFormat(codec uint8) (*Format, error) {
	name, ok := mapCodecs[codec]
	if !ok {
		return nil, fmt.Errorf("unknown codec with payload %d", codec)
	}
	frmt := &Format{
		Payload:   codec,
		Name:      name,
		ClockRate: 8000,
		Channels:  1,
	}
	if codec == RFC4733PT {
		frmt.Params = append(frmt.Params, "0-16")
	}
	return frmt, nil
}

func (ses *Session) GetAudioMediaFlow() *Media {
	return ses.GetMediaFlow(Audio)
}

func (ses *Session) GetEffectivePTime() string {
	media := ses.GetAudioMediaFlow()
	attrbnm := "ptime"
	ptime := media.Attributes.Get(attrbnm)
	if ptime != "" {
		return ptime
	}
	ptime = ses.Attributes.Get(attrbnm)
	if ptime != "" {
		return ptime
	}
	return "20"
}

func (ses *Session) GetEffectiveMediaIPv4(media *Media) string {
	var ipv4 string
	for i := range media.Connection {
		ipv4 = media.Connection[i].Address
		if ipv4 != "" && ipv4 != "0.0.0.0" {
			return ipv4
		}
	}
	return ses.Connection.Address
}

func (ses *Session) GetEffectiveMediaSocket(media *Media) string {
	var ipv4 string

	for i := range media.Connection {
		ip := media.Connection[i].Address
		if ip != "" && ip != "0.0.0.0" {
			ipv4 = ip
			break
		}
	}

	if ipv4 = cmp.Or(ipv4, ses.Connection.Address); ipv4 == "" || media.Port <= 0 {
		return ""
	}

	return fmt.Sprintf("%s:%d", ipv4, media.Port)
}

func (ses *Session) GetEffectiveConnectionForMedia(medType string) string {
	for _, media := range ses.Media {
		if media.Type == medType {
			return ses.GetEffectiveMediaIPv4(media)
		}
	}
	return ""
}

func (ses *Session) GetEffectiveMediaUdpAddr(medType string) *net.UDPAddr {
	media := ses.GetMediaFlow(medType)
	if media == nil {
		return nil
	}
	skt := ses.GetEffectiveMediaSocket(media)
	addr, _ := net.ResolveUDPAddr("udp", skt)
	return addr
}

func (ses *Session) GetMediaFlow(medType string) *Media {
	for _, media := range ses.Media {
		if media.Type == medType {
			return media
		}
	}
	return nil
}

func (ses *Session) AlignMediaFlows(offer *Session) error {
	srcMediaMap := make(map[string]*Media, len(ses.Media))
	for _, media := range ses.Media {
		if _, ok := srcMediaMap[media.Type]; ok {
			return fmt.Errorf("duplicate media type [%s] in source session - Not supported", media.Type)
		}
		srcMediaMap[media.Type] = media
	}
	// need to check if SDP of src contains distinct media types
	ses.Media = make([]*Media, 0, len(offer.Media))
	for _, media := range offer.Media {
		if flow, ok := srcMediaMap[media.Type]; ok {
			ses.Media = append(ses.Media, flow)
		} else {
			ses.Media = append(ses.Media, media.clone(0))
		}
	}
	return nil
}

func (ses *Session) SetConnection(medType, IPv4 string, Port int, setGlobal bool) *Session {
	if medType == "" {
		return ses
	}

	conn := &Connection{
		Network: NetworkInternet,
		Type:    TypeIPv4,
		Address: IPv4,
	}

	if setGlobal {
		ses.Connection = conn
		for _, media := range ses.Media {
			media.Connection = nil
			if media.Type == medType {
				media.Connection = nil
				media.Port = Port
			}
		}
		return ses
	}

	for _, media := range ses.Media {
		if media.Type == medType {
			media.Connection = nil
			media.Connection = append(media.Connection, conn)
			media.Port = Port
			return ses
		}
	}

	return ses
}

func (ses *Session) DropFlowsExcept(medTypes ...string) *Session {
	return ses.dropFlows(true, medTypes...)
}

func (ses *Session) DropFlows(medTypes ...string) *Session {
	return ses.dropFlows(false, medTypes...)
}

func (ses *Session) dropFlows(except bool, medTypes ...string) *Session {
	medTypesMap, ok := sliceToSet(medTypes...)
	if !ok {
		return ses
	}
	for i := 0; i < len(ses.Media); {
		if _, ok := medTypesMap[ses.Media[i].Type]; ok != except {
			ses.Media = append(ses.Media[:i], ses.Media[i+1:]...)
		} else {
			i++
		}
	}
	return ses
}

func (ses *Session) DisableFlowsExcept(medTypes ...string) *Session {
	return ses.disableFlows(true, medTypes...)
}

func (ses *Session) DisableFlows(medTypes ...string) *Session {
	return ses.disableFlows(false, medTypes...)
}

func (ses *Session) disableFlows(except bool, medTypes ...string) *Session {
	medTypesMap, ok := sliceToSet(medTypes...)
	if !ok {
		return ses
	}
	for _, mf := range ses.Media {
		if _, ok := medTypesMap[mf.Type]; ok != except {
			mf.Port = 0
		}
	}
	return ses
}

func sliceToSet[T comparable](items ...T) (map[T]struct{}, bool) {
	if len(items) == 0 {
		return nil, false
	}
	set := make(map[T]struct{}, len(items))
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set, true
}

func (ses *Session) AreAllFlowsDroppedOrDisabled() bool {
	if len(ses.Media) == 0 {
		return true
	}
	for _, media := range ses.Media {
		if media.Port > 0 {
			return false
		}
	}
	return true
}

func (ses *Session) IsT38Image() bool {
	for _, media := range ses.Media {
		if media.Type == Image && media.Port > 0 && media.Proto == Udptl && media.FormatDescr == "t38" {
			return true
		}
	}
	return false
}

func (ses *Session) GetEffectiveMediaDirective() string {
	media := ses.GetAudioMediaFlow()
	if media.Mode != "" {
		return media.Mode
	}
	if ses.Mode != "" {
		return ses.Mode
	}
	return SendRecv
}

func (ses *Session) IsCallHolding() bool {
	return IsMedDirHolding(ses.GetEffectiveMediaDirective())
}

func (ses *Session) IsCallHeld() bool {
	media := ses.GetAudioMediaFlow()
	var mode string
	if media.Mode != "" {
		mode = media.Mode
	} else if ses.Mode != "" {
		mode = ses.Mode
	} else {
		mode = SendRecv
	}
	if mode == SendOnly || mode == Inactive {
		return true
	}
	if ipv4 := ses.GetEffectiveMediaIPv4(media); ipv4 == "" || ipv4 == "0.0.0.0" {
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
	Formats     []*Format     // Media Format for RTP/AVP or RTP/SAVP protocols ("rtpmap", "fmtp", "rtcp-fb")
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

func (m *Media) clone(prt int) *Media {
	newPort, newMode := func() (int, string) {
		if prt == 0 {
			return 0, ""
		}
		return m.Port, m.Mode
	}()

	mediaClone := &Media{
		Type:        m.Type,
		Port:        newPort,
		PortNum:     m.PortNum,
		Proto:       m.Proto,
		Information: m.Information,
		Mode:        newMode,
		FormatDescr: m.FormatDescr,
	}

	// Clone Connection slice for Media
	if m.Connection != nil {
		mediaClone.Connection = make([]*Connection, len(m.Connection))
		for j, conn := range m.Connection {
			if conn != nil {
				mediaClone.Connection[j] = &Connection{
					Network:    conn.Network,
					Type:       conn.Type,
					Address:    conn.Address,
					TTL:        conn.TTL,
					AddressNum: conn.AddressNum,
				}
			}
		}
	}

	// Clone Bandwidth slice for Media
	if m.Bandwidth != nil {
		mediaClone.Bandwidth = make([]*Bandwidth, len(m.Bandwidth))
		for j, bw := range m.Bandwidth {
			if bw != nil {
				mediaClone.Bandwidth[j] = &Bandwidth{
					Type:  bw.Type,
					Value: bw.Value,
				}
			}
		}
	}

	// Clone Key slice for Media
	if m.Key != nil {
		mediaClone.Key = make([]*Key, len(m.Key))
		for j, k := range m.Key {
			if k != nil {
				mediaClone.Key[j] = &Key{
					Method: k.Method,
					Value:  k.Value,
				}
			}
		}
	}

	// Clone Attributes for Media
	mediaClone.Attributes = m.Attributes.clone()

	// Clone Format slice for Media
	if m.Formats != nil {
		mediaClone.Formats = make([]*Format, len(m.Formats))
		for j, f := range m.Formats {
			if f != nil {
				mediaClone.Formats[j] = f.clone()
			}
		}
	}

	return mediaClone
}

// FormatByPayload returns format description by payload type.
func (m *Media) FormatByPayload(payload uint8) *Format {
	for _, f := range m.Formats {
		if f.Payload == payload {
			return f
		}
	}
	return nil
}

// FormatByName returns format by codec name - way faster (10x) than using ToLower() on each format name and input format name.
func (m *Media) FormatByName(frmt string) *Format {
	for _, f := range m.Formats {
		if strings.EqualFold(f.Name, frmt) {
			return f
		}
	}
	return nil
}

func (m *Media) KeepOnlyFirstAudioCodecAlongRFC4733(audiofrmts ...string) (bool, bool) {
	if len(m.Formats) == 0 {
		return false, false
	}

	for i, f := range audiofrmts {
		audiofrmts[i] = AsciiToLower(f)
	}

	var dtmfFormat, audioFormat *Format

	for _, f := range m.Formats {
		if audioFormat != nil && dtmfFormat != nil {
			break
		}
		if f.Name == RFC4733 {
			if dtmfFormat == nil {
				dtmfFormat = f
			}
			continue
		}
		if audioFormat == nil {
			if slices.Contains(audiofrmts, AsciiToLower(f.Name)) {
				audioFormat = f
			}
		}
	}
	if audioFormat == nil {
		return false, false
	}

	m.Formats = make([]*Format, 0, 2)

	m.Formats = append(m.Formats, audioFormat)

	if dtmfFormat != nil {
		m.Formats = append(m.Formats, dtmfFormat)
		return true, true
	}

	return true, false
}

func (m *Media) OrderFormatsByName(filterformats ...string) {
	if len(filterformats) == 0 || len(filterformats) == 1 && filterformats[0] == "*" {
		return
	}

	for i, ff := range filterformats {
		filterformats[i] = AsciiToLower(ff)
	}

	filterformatsmap := make(map[string]struct{}, len(filterformats))
	for _, ff := range filterformats {
		filterformatsmap[ff] = struct{}{}
	}

	formatsmap := make(map[string]*Format, len(m.Formats))
	formats := make([]string, len(m.Formats))
	for i, frmt := range m.Formats {
		ff := AsciiToLower(frmt.Name)
		formatsmap[ff] = frmt
		formats[i] = ff
	}

	m.Formats = make([]*Format, 0, len(m.Formats))
	for _, ff := range filterformats {
		if ff == "*" {
			for _, f := range formats {
				if _, ok := filterformatsmap[f]; !ok {
					m.Formats = append(m.Formats, formatsmap[f])
				}
			}
			continue
		}
		if f, ok := formatsmap[ff]; ok {
			m.Formats = append(m.Formats, f)
		}
	}
}

func (m *Media) FormatNames() []string {
	names := make([]string, len(m.Formats))
	for i, f := range m.Formats {
		names[i] = f.Name
	}

	return names
}

func (m *Media) FilterFormatsByName(frmts ...string) {
	m.findFormatByName(false, frmts...)
}

func (m *Media) DropFormatsByName(frmts ...string) {
	m.findFormatByName(true, frmts...)
}

func (m *Media) FilterFormatsByPayload(frmts ...uint8) {
	m.findFormatByPayload(false, frmts...)
}

func (m *Media) DropFormatsByPayload(frmts ...uint8) {
	m.findFormatByPayload(true, frmts...)
}

func (m *Media) findFormatByName(drop bool, frmts ...string) {
	if len(frmts) == 0 {
		return
	}
	formatNames := make(map[string]struct{}, len(frmts))
	for _, f := range frmts {
		ff := AsciiToLower(f)
		formatNames[ff] = struct{}{}
	}
	m.filterFormats(drop, func(f *Format) bool {
		ff := AsciiToLower(f.Name)
		_, ok := formatNames[ff]
		return ok
	})
}

func (m *Media) findFormatByPayload(drop bool, frmts ...uint8) {
	if len(frmts) == 0 {
		return
	}
	formatPayloads := make(map[uint8]struct{}, len(frmts))
	for _, v := range frmts {
		formatPayloads[v] = struct{}{}
	}
	m.filterFormats(drop, func(f *Format) bool {
		_, ok := formatPayloads[f.Payload]
		return ok
	})
}

func (m *Media) filterFormats(drop bool, matchFunc func(f *Format) bool) {
	for i := 0; i < len(m.Formats); {
		if matchFunc(m.Formats[i]) == drop {
			m.Formats = append(m.Formats[:i], m.Formats[i+1:]...)
		} else {
			i++
		}
	}
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

func (f *Format) clone() *Format {
	clone := &Format{
		Payload:   f.Payload,
		Name:      f.Name,
		ClockRate: f.ClockRate,
		Channels:  f.Channels,
		Feedback:  make([]string, len(f.Feedback)),
		Params:    make([]string, len(f.Params)),
	}

	copy(clone.Feedback, f.Feedback)
	copy(clone.Params, f.Params)

	return clone
}

func isRTP(media, proto string) bool {
	switch media {
	case "audio", "video":
		return strings.Contains(proto, "RTP/AVP") || strings.Contains(proto, "RTP/SAVP")
	default:
		return false
	}
}

func AsciiToLower(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := range len(s) {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += byte(DeltaRune)
		}
		b.WriteByte(c)
	}
	return b.String()
}

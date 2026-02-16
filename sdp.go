package sdp

import (
	"cmp"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"
)

var epoch = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)

const deltaRune rune = 'a' - 'A'

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
	PTime       string       // ptime:20 20ms
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
		PTime:       s.PTime,
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

// Deletes attribute by name
func (s *Session) DeleteAttribute(name string) {
	s.Attributes = slices.DeleteFunc(s.Attributes, func(at *Attr) bool { return at.Name == name })
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

func (ses *Session) BuildSelfAnswer(currentLocalMediaDirective string, audiofrmts ...string) (*Session, bool, error) {
	if len(audiofrmts) == 0 {
		return nil, false, fmt.Errorf("cannot build EchoResponder answer: no audio formats provided")
	}

	if mf := ses.GetAudioMediaFlow(); mf == nil {
		return nil, false, fmt.Errorf("cannot build EchoResponder answer: no audio media flow found")
	} else {
		if mf.Port == 0 {
			return nil, false, fmt.Errorf("cannot build EchoResponder answer: audio media flow is disabled")
		}
	}
	answer := ses.Clone()
	answer.Mode = ""
	answer.Origin.SessionVersion = 1

	mf := answer.DisableFlowsExcept(Audio).GetAudioMediaFlow()

	audiofound, dtmffound := mf.KeepOnlyFirstAudioCodecAlongRFC4733(audiofrmts...)

	if !audiofound {
		return nil, false, fmt.Errorf("cannot build EchoResponder answer: no common audio formats found in audio media flow")
	}

	mf.Mode = NegotiateAnswerMode(currentLocalMediaDirective, ses.GetEffectiveMediaDirective())

	return answer, dtmffound, nil
}

func NewSessionSDP(sesID, sesVer int64, ipv4, nm, ssrc, mdir string, port int, codecs []uint8) (*Session, error) {
	formats := make([]*Format, 0, len(codecs))
	for _, codec := range codecs {
		frmt, err := BuildFormat(codec)
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
				PTime:   "20",
				Formats: formats,
			},
		},
	}, nil
}

func BuildFormat(codec uint8) (*Format, error) {
	cinfo, ok := codecsInfoMap[codec]
	if !ok {
		return nil, fmt.Errorf("unknown codec information with payload %d", codec)
	}
	frmt := &Format{
		Payload:   codec,
		Name:      cinfo.Name,
		ClockRate: cinfo.ClockRate,
		Channels:  cinfo.Channels,
	}
	if cinfo.Name == RFC4733 {
		frmt.Params = append(frmt.Params, "0-16")
	}
	return frmt, nil
}

func BuildFormatByName(codecName string) (*Format, error) {
	cinfo, ok := GetCodecByName(codecName)
	if !ok {
		return nil, fmt.Errorf("unknown codec information with name %s", codecName)
	}
	frmt := &Format{
		Payload:   cinfo.PayloadType,
		Name:      cinfo.Name,
		ClockRate: cinfo.ClockRate,
		Channels:  cinfo.Channels,
	}
	if cinfo.Name == RFC4733 {
		frmt.Params = append(frmt.Params, "0-16")
	}
	return frmt, nil
}

func (ses *Session) GetAudioMediaFlow() *Media {
	return ses.GetMediaFlow(Audio)
}

func (ses *Session) GetEffectivePTime() string {
	media := ses.GetAudioMediaFlow()
	ptValue := media.PTime
	if ptValue != "" {
		return ptValue
	}
	ptValue = ses.PTime
	if ptValue != "" {
		return ptValue
	}
	return "20"
}

func (ses *Session) GetEffectiveMediaIPv4(media *Media) string {
	for i := range media.Connection {
		return media.Connection[i].Address
	}
	if ses.Connection == nil {
		return ""
	}
	return ses.Connection.Address
}

func (ses *Session) GetEffectiveMediaSocket(media *Media) string {
	var (
		mediaIPv4 string
		sessAddr  string
	)

	for i := range media.Connection {
		ip := media.Connection[i].Address
		if ip != "" && ip != "0.0.0.0" {
			mediaIPv4 = ip
			break
		}
	}

	if ses.Connection != nil {
		sessAddr = ses.Connection.Address
	}

	if mediaIPv4 = cmp.Or(mediaIPv4, sessAddr); mediaIPv4 == "" || media.Port <= 0 {
		return ""
	}

	return fmt.Sprintf("%s:%d", mediaIPv4, media.Port)
}

func (ses *Session) GetEffectiveConnectionForMedia(medType string) string {
	for _, media := range ses.Media {
		if media.Type == medType {
			return ses.GetEffectiveMediaIPv4(media)
		}
	}
	return ""
}

func (ses *Session) GetEffectiveMediaUdpAddr(medType string) (*net.UDPAddr, error) {
	media := ses.GetMediaFlow(medType)
	if media == nil {
		return nil, errors.New("media flow not exists")
	}
	skt := ses.GetEffectiveMediaSocket(media)
	if skt == "" {
		return nil, errors.New("cannot find any IPv4")
	}
	return net.ResolveUDPAddr("udp", skt)
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

// Parameters:
//   - setGlobal: if true; removes c-line from Session level
//   - removeGlobal: if true; remove c-line from Session level (only effective if setGlobal is false and medType exists)
func (ses *Session) SetConnection(medType, ipv4 string, port int, setGlobal, removeGlobal bool) *Session {
	if medType == "" {
		return ses
	}

	conn := &Connection{
		Network: NetworkInternet,
		Type:    TypeIPv4,
		Address: ipv4,
	}

	if setGlobal {
		ses.Connection = conn
		for _, media := range ses.Media {
			media.Connection = nil
			if media.Type == medType {
				media.Port = port
			}
		}

		return ses
	}

	mediaTypeFound := false
	for _, media := range ses.Media {
		if media.Type == medType {
			media.Connection = nil
			media.Connection = append(media.Connection, conn)
			media.Port = port
			mediaTypeFound = true
			break
		}
	}

	if mediaTypeFound && removeGlobal {
		ses.Connection = nil
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

func (ses *Session) RestoreMissingRtpmaps() []string {
	var missing []string
	for _, m := range ses.Media {
		switch m.Type {
		case Audio, Video:
		default:
			continue
		}
		for _, f := range m.Formats {
			if f.Payload >= DynamicPayloadStart && f.Name != "" {
				continue
			}
			if cinfo, ok := codecsInfoMap[f.Payload]; ok {
				f.Name = cinfo.Name
				f.ClockRate = cinfo.ClockRate
				f.Channels = cinfo.Channels
			} else {
				missing = append(missing, fmt.Sprintf("Media Type: %s, Payload Type: %d", m.Type, f.Payload))
			}
		}
	}
	return missing
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
	if media != nil && media.Mode != "" {
		return media.Mode
	}
	if ses.Mode != "" {
		return ses.Mode
	}
	return SendRecv
}

func (ses *Session) IsCallHeld() bool {
	media := ses.GetAudioMediaFlow()
	var mode string
	switch {
	case media != nil && media.Mode != "":
		mode = media.Mode
	case ses.Mode != "":
		mode = ses.Mode
	default:
		mode = SendRecv
	}
	if IsMedDirHolding(mode) {
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
	Attributes  Attributes    // Attributes ("a=")
	Mode        string        // Streaming mode ("sendrecv", "recvonly", "sendonly", or "inactive")
	PTime       string
	Formats     []*Format // Media Format for RTP/AVP or RTP/SAVP protocols ("rtpmap", "fmtp", "rtcp-fb")
	FormatDescr string    // Media Format for other protocols
}

const (
	SendRecv = "sendrecv"
	SendOnly = "sendonly"
	RecvOnly = "recvonly"
	Inactive = "inactive"

	PTime = "ptime"

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

func GetSubsequentMediaDirectiveMode(currentLocal string, putCallonHold bool) (out string, ok bool) {
	if putCallonHold {
		switch currentLocal {
		case "", SendRecv:
			out = SendOnly
		case RecvOnly:
			out = Inactive
		default:
			return "", false
		}
	} else {
		switch currentLocal {
		case SendOnly:
			out = SendRecv
		case Inactive:
			out = RecvOnly
		case "", SendRecv:
			out = SendRecv
		default:
			return "", false
		}
	}
	return currentLocal, true
}

// NegotiateAnswerMode negotiates streaming mode.
func NegotiateAnswerMode(oldLocal, newRemote string) (newLocal string) {
	if oldLocal == "" {
		oldLocal = SendRecv
	}
	if newRemote == "" {
		newRemote = SendRecv
	}

	switch newRemote {
	case SendRecv:
		switch oldLocal {
		case RecvOnly, SendRecv:
			return SendRecv
		case Inactive, SendOnly:
			return SendOnly
		}
	case SendOnly:
		switch oldLocal {
		case Inactive, RecvOnly, SendRecv:
			return RecvOnly
		}
	case RecvOnly:
		switch oldLocal {
		case Inactive, SendOnly:
			return SendOnly
		case SendRecv:
			return SendRecv
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
		PTime:       m.PTime,
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
				mediaClone.Formats[j] = f.Clone()
			}
		}
	}

	return mediaClone
}

// Deletes attribute by name
func (m *Media) DeleteAttribute(name string) {
	m.Attributes = slices.DeleteFunc(m.Attributes, func(at *Attr) bool { return at.Name == name })
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

func (m *Media) KeepOnlyFirstAudioCodecAlongRFC4733(audioformats ...string) (withAudio, withDtmf bool) {
	if len(m.Formats) == 0 {
		return
	}

	audioformatMap := make(map[string]struct{}, len(audioformats))
	for _, f := range audioformats {
		audioformatMap[asciiToLower(f)] = struct{}{}
	}

	selectedFormats := make([]*Format, 0, 2)

	for _, f := range m.Formats {
		if len(selectedFormats) == 2 {
			break
		}
		frmt := asciiToLower(f.Name)
		switch frmt {
		case RFC4733:
			if !withDtmf {
				withDtmf = true
				selectedFormats = append(selectedFormats, f)
			}
		case ComfortNoise:
		default:
			if !withAudio {
				if _, ok := audioformatMap[frmt]; ok {
					withAudio = true
					selectedFormats = append(selectedFormats, f)
				}
			}
		}
	}

	m.Formats = selectedFormats

	return
}

func (m *Media) WithAtLeastOneAudioFormat() bool {
	return slices.ContainsFunc(m.Formats, func(f *Format) bool { return f.IsAudioFormat() })
}

func (m *Media) GetFirstAudioFormatName() string {
	for _, frmt := range m.Formats {
		if frmt.IsAudioFormat() {
			return frmt.Name
		}
	}
	return ""
}

func (m *Media) GetFirstAudioFormat() *Format {
	for _, frmt := range m.Formats {
		if frmt.IsAudioFormat() {
			return frmt
		}
	}
	return nil
}

func (m *Media) IsRFC4733Available() bool {
	return m.GetFirstRFC4733() != nil
}

func (m *Media) GetFirstRFC4733() *Format {
	for _, frmt := range m.Formats {
		if frmt.LowerName() == RFC4733 {
			return frmt
		}
	}
	return nil
}

func (m *Media) OrderFormatsByName(filterformats ...string) {
	if len(filterformats) == 0 || len(filterformats) == 1 && filterformats[0] == "*" {
		return
	}

	for i, ff := range filterformats {
		filterformats[i] = asciiToLower(ff)
	}

	filterformatsmap := make(map[string]struct{}, len(filterformats))
	for _, ff := range filterformats {
		filterformatsmap[ff] = struct{}{}
	}

	formatsmap := make(map[string]*Format, len(m.Formats))
	formats := make([]string, len(m.Formats))
	for i, frmt := range m.Formats {
		ff := asciiToLower(frmt.Name)
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

func (m *Media) FormatLowerNames() []string {
	names := make([]string, len(m.Formats))
	for i, f := range m.Formats {
		names[i] = f.LowerName()
	}

	return names
}

func (m *Media) FilterFormatsByName(frmts ...string) bool {
	return m.findFormatByName(false, frmts...)

}

func (m *Media) DropFormatsByName(frmts ...string) bool {
	return m.findFormatByName(true, frmts...)
}

func (m *Media) FilterFormatsByPayload(frmts ...uint8) bool {
	return m.findFormatByPayload(false, frmts...)
}

func (m *Media) DropFormatsByPayload(frmts ...uint8) bool {
	return m.findFormatByPayload(true, frmts...)
}

func (m *Media) findFormatByName(drop bool, frmts ...string) bool {
	if len(frmts) == 0 {
		return false
	}
	formatNames := make(map[string]struct{}, len(frmts))
	for _, f := range frmts {
		ff := asciiToLower(f)
		formatNames[ff] = struct{}{}
	}
	m.filterFormats(drop, func(f *Format) bool {
		ff := asciiToLower(f.Name)
		_, ok := formatNames[ff]
		return ok
	})
	return m.WithAtLeastOneAudioFormat()
}

func (m *Media) findFormatByPayload(drop bool, frmts ...uint8) bool {
	if len(frmts) == 0 {
		return false
	}
	formatPayloads := make(map[uint8]struct{}, len(frmts))
	for _, v := range frmts {
		formatPayloads[v] = struct{}{}
	}
	m.filterFormats(drop, func(f *Format) bool {
		_, ok := formatPayloads[f.Payload]
		return ok
	})
	return m.WithAtLeastOneAudioFormat()
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

func (f *Format) LowerName() string {
	return asciiToLower(f.Name)
}

func (f *Format) IsAudioFormat() bool {
	frmtnm := asciiToLower(f.Name)
	switch frmtnm {
	case RFC4733, ComfortNoise:
		return false
	}
	return true
}

func (f *Format) String() string {
	return f.Name
}

func (f *Format) Clone() *Format {
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

func asciiToLower(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := range len(s) {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += byte(deltaRune)
		}
		b.WriteByte(c)
	}
	return b.String()
}

func asciiToLowerByte(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		b += byte(deltaRune)
	}
	return b
}

package elm327

import (
	"fmt"
	"strings"
	"vehicle-simulator/protocol"
	"vehicle-simulator/vehicle"
)

// Session tracks adapter configuration (echo, linefeeds, headers, spacing) and handles command responses.
type Session struct {
	Vehicle    *vehicle.Vehicle
	EchoOn     bool
	LinefeedOn bool
	SpacesOn   bool
	HeadersOn  bool
	Protocol   string // e.g. "6"
}

// NewSession creates an ELM327 command session matching default properties.
func NewSession(v *vehicle.Vehicle) *Session {
	return &Session{
		Vehicle:    v,
		EchoOn:     true,
		LinefeedOn: true,
		SpacesOn:   true,
		HeadersOn:  false,
		Protocol:   "6", // ISO 15765-4 CAN (11 bit ID, 500 kbaud)
	}
}

// HandleCommand processes a raw input command line, performs the requested action, and builds the return packet.
func (s *Session) HandleCommand(rawCmd string) string {
	cleaned := protocol.CleanCommand(rawCmd)
	if cleaned == "" {
		return ">"
	}

	var response string
	delim := "\r"
	if s.LinefeedOn {
		delim = "\r\n"
	}

	if strings.HasPrefix(cleaned, "AT") {
		response = s.handleATCommand(cleaned)
	} else {
		response = s.handleOBDCommand(cleaned)
	}

	var sb strings.Builder
	// If echo is enabled, send back the user's input line (minus newlines)
	if s.EchoOn {
		echoStr := strings.TrimRight(rawCmd, "\r\n")
		sb.WriteString(echoStr)
		sb.WriteString(delim)
	}

	if response != "" {
		sb.WriteString(response)
		sb.WriteString(delim)
	}
	sb.WriteString(">")

	return sb.String()
}

func (s *Session) handleATCommand(cmd string) string {
	switch {
	case cmd == "ATZ":
		// Reset to defaults
		s.EchoOn = true
		s.LinefeedOn = true
		s.SpacesOn = true
		s.HeadersOn = false
		return "ELM327 v1.5"
	case cmd == "ATPC":
		return "OK"
	case cmd == "ATD" || cmd == "AT D":
		// Restore defaults
		s.EchoOn = true
		s.LinefeedOn = true
		s.SpacesOn = true
		s.HeadersOn = false
		return "OK"
	case cmd == "ATE0":
		s.EchoOn = false
		return "OK"
	case cmd == "ATE1":
		s.EchoOn = true
		return "OK"
	case cmd == "ATL0":
		s.LinefeedOn = false
		return "OK"
	case cmd == "ATL1":
		s.LinefeedOn = true
		return "OK"
	case cmd == "ATS0":
		s.SpacesOn = false
		return "OK"
	case cmd == "ATS1":
		s.SpacesOn = true
		return "OK"
	case cmd == "ATH0":
		s.HeadersOn = false
		return "OK"
	case cmd == "ATH1":
		s.HeadersOn = true
		return "OK"
	case cmd == "ATAT1" || cmd == "ATAT0":
		return "OK"
	case strings.HasPrefix(cmd, "ATSP"):
		if len(cmd) > 4 {
			s.Protocol = cmd[4:]
		}
		return "OK"
	case cmd == "ATDPN":
		return "A" + s.Protocol
	case cmd == "ATRV":
		s.Vehicle.RLock()
		volts := s.Vehicle.Battery
		s.Vehicle.RUnlock()
		return fmt.Sprintf("%.1fV", volts)
	default:
		// Accept and reply OK for any other unsupported commands to keep communication alive
		return "OK"
	}
}

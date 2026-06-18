package elm327

import (
	"fmt"
	"strings"
	"vehicle-simulator/vehicle"
)

// formatOBDResponse applies space formatting and adds CAN bus headers (7E8) and length indicators when ATH1 is active.
func (s *Session) formatOBDResponse(payload []byte) string {
	var tokens []string

	if s.HeadersOn {
		// CAN header (7E8) followed by length of data frame
		tokens = []string{"7E8", fmt.Sprintf("%02X", len(payload))}
		for _, b := range payload {
			tokens = append(tokens, fmt.Sprintf("%02X", b))
		}
	} else {
		tokens = make([]string, len(payload))
		for i, b := range payload {
			tokens[i] = fmt.Sprintf("%02X", b)
		}
	}

	if s.SpacesOn {
		return strings.Join(tokens, " ")
	}
	return strings.Join(tokens, "")
}

func (s *Session) handleOBDCommand(cmd string) string {
	s.Vehicle.RLock()
	defer s.Vehicle.RUnlock()

	switch cmd {
	case "0101":
		// Mode 01 PID 01 - MIL status and active DTC count
		milByte := byte(0)
		if s.Vehicle.MILOn {
			milByte |= 0x80
		}
		milByte |= byte(len(s.Vehicle.StoredCodes) & 0x7F)
		payload := []byte{0x41, 0x01, milByte, 0x07, 0x65, 0x04}
		return s.formatOBDResponse(payload)

	case "010C":
		// Mode 01 PID 0C - Engine RPM
		// Formula: RPM = ((A*256)+B)/4 -> (A*256)+B = RPM*4
		rpmVal := s.Vehicle.RPM * 4
		a := byte(rpmVal / 256)
		b := byte(rpmVal % 256)
		payload := []byte{0x41, 0x0C, a, b}
		return s.formatOBDResponse(payload)

	case "010D":
		// Mode 01 PID 0D - Vehicle Speed
		speedVal := byte(s.Vehicle.Speed)
		payload := []byte{0x41, 0x0D, speedVal}
		return s.formatOBDResponse(payload)

	case "0105":
		// Mode 01 PID 05 - Coolant Temp
		// Formula: Temp = A - 40 -> A = Temp + 40
		tempVal := byte(s.Vehicle.Temp + 40)
		payload := []byte{0x41, 0x05, tempVal}
		return s.formatOBDResponse(payload)

	case "0104":
		// Mode 01 PID 04 - Calculated Engine Load
		// Formula: Load = A * 100 / 255 -> A = Load * 255 / 100
		loadVal := byte((s.Vehicle.Load * 255) / 100)
		payload := []byte{0x41, 0x04, loadVal}
		return s.formatOBDResponse(payload)

	case "0111":
		// Mode 01 PID 11 - Throttle Position
		// Formula: Throttle = A * 100 / 255 -> A = Throttle * 255 / 100
		throttleVal := byte((s.Vehicle.Throttle * 255) / 100)
		payload := []byte{0x41, 0x11, throttleVal}
		return s.formatOBDResponse(payload)

	case "010F":
		// Mode 01 PID 0F - Intake Air Temperature
		// Formula: Intake = A - 40 -> A = Intake + 40
		intakeVal := byte(s.Vehicle.Intake + 40)
		payload := []byte{0x41, 0x0F, intakeVal}
		return s.formatOBDResponse(payload)

	case "010E":
		// Mode 01 PID 0E - Ignition Timing Advance
		// Formula: Timing = A / 2.0 - 64 -> A = (Timing + 64) * 2
		timingVal := byte((s.Vehicle.Timing + 64) * 2)
		payload := []byte{0x41, 0x0E, timingVal}
		return s.formatOBDResponse(payload)

	case "0106":
		// Mode 01 PID 06 - Short Term Fuel Trim
		// Formula: STFT = (A - 128) * 100 / 128 -> A = (STFT * 128 / 100) + 128
		// Let's assume 0% fuel trim (A=128)
		payload := []byte{0x41, 0x06, 128}
		return s.formatOBDResponse(payload)

	case "0107":
		// Mode 01 PID 07 - Long Term Fuel Trim
		// Let's assume 0% fuel trim (A=128)
		payload := []byte{0x41, 0x07, 128}
		return s.formatOBDResponse(payload)

	case "0114":
		// Mode 01 PID 14 - Oxygen Sensor Voltage
		// Formula: O2V = A * 0.005 -> A = O2V / 0.005
		// Let's assume 0.45V (A=90)
		payload := []byte{0x41, 0x14, 90}
		return s.formatOBDResponse(payload)

	case "03":
		// Mode 03 - Stored DTCs
		// Response Format: 43 [count] [DTC1_B1] [DTC1_B2] ...
		count := byte(len(s.Vehicle.StoredCodes))
		payload := []byte{0x43, count}
		for _, code := range s.Vehicle.StoredCodes {
			bytes, err := vehicle.DtcToBytes(code)
			if err == nil {
				payload = append(payload, bytes...)
			}
		}
		return s.formatOBDResponse(payload)

	case "07":
		// Mode 07 - Pending DTCs
		// Response Format: 47 [count] [DTC1_B1] [DTC1_B2] ...
		count := byte(len(s.Vehicle.PendingCodes))
		payload := []byte{0x47, count}
		for _, code := range s.Vehicle.PendingCodes {
			bytes, err := vehicle.DtcToBytes(code)
			if err == nil {
				payload = append(payload, bytes...)
			}
		}
		return s.formatOBDResponse(payload)

	case "04":
		// Mode 04 - Clear Diagnostic Trouble Codes
		// Requires write lock to mutate. We temporarily unlock the read lock.
		s.Vehicle.RUnlock()
		s.Vehicle.Lock()
		s.Vehicle.StoredCodes = []string{}
		s.Vehicle.PendingCodes = []string{}
		s.Vehicle.MILOn = false
		s.Vehicle.Unlock()
		s.Vehicle.RLock() // Restore lock for the deferred unlock

		payload := []byte{0x44}
		return s.formatOBDResponse(payload)

	default:
		return "NO DATA"
	}
}

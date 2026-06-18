package vehicle

import (
	"fmt"
	"strconv"
)

// DtcToBytes converts an alphanumeric DTC string (e.g. "P0133") into its 2-byte OBD-II OBD representation.
// OBD-II DTC format:
// First character determines high 2 bits of the first byte:
// - P (Powertrain) = 00
// - C (Chassis) = 01
// - B (Body) = 10
// - U (Network) = 11
// Second character determines next 2 bits of the first byte (usually 0-3).
// Third character determines low 4 bits of the first byte.
// Fourth character determines high 4 bits of the second byte.
// Fifth character determines low 4 bits of the second byte.
func DtcToBytes(code string) ([]byte, error) {
	if len(code) != 5 {
		return nil, fmt.Errorf("invalid DTC length: %s", code)
	}

	var b1, b2 byte

	// 1. Prefix (first character)
	switch code[0] {
	case 'P':
		b1 |= 0x00 << 6
	case 'C':
		b1 |= 0x01 << 6
	case 'B':
		b1 |= 0x02 << 6
	case 'U':
		b1 |= 0x03 << 6
	default:
		return nil, fmt.Errorf("invalid DTC prefix char: %c", code[0])
	}

	// 2. Second character (digit 0-3)
	d2 := code[1] - '0'
	if d2 > 3 {
		return nil, fmt.Errorf("invalid DTC digit 2: %c", code[1])
	}
	b1 |= d2 << 4

	// 3. Third character (hex digit)
	d3, err := strconv.ParseUint(string(code[2]), 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid DTC digit 3: %s", err)
	}
	b1 |= byte(d3)

	// 4. Fourth character (hex digit)
	d4, err := strconv.ParseUint(string(code[3]), 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid DTC digit 4: %s", err)
	}
	b2 |= byte(d4) << 4

	// 5. Fifth character (hex digit)
	d5, err := strconv.ParseUint(string(code[4]), 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid DTC digit 5: %s", err)
	}
	b2 |= byte(d5)

	return []byte{b1, b2}, nil
}

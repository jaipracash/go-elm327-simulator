package protocol

import "strings"

// CleanCommand converts incoming packets to uppercase and strips whitespace and line terminators.
func CleanCommand(cmd string) string {
	cmd = strings.ReplaceAll(cmd, " ", "")
	cmd = strings.ReplaceAll(cmd, "\r", "")
	cmd = strings.ReplaceAll(cmd, "\n", "")
	return strings.ToUpper(cmd)
}

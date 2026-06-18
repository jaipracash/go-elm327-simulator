package bluetooth

import (
	"io"
	"log"
	"vehicle-simulator/elm327"
	"vehicle-simulator/vehicle"
)

// HandleClient manages the command loop for an active connection (network or serial).
func HandleClient(conn io.ReadWriteCloser, v *vehicle.Vehicle, clientName string) {
	defer conn.Close()
	log.Printf("[%s] Client connected", clientName)

	session := elm327.NewSession(v)
	buffer := make([]byte, 1024)
	var inputBuffer []byte

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("[%s] Read error: %v", clientName, err)
			}
			break
		}

		inputBuffer = append(inputBuffer, buffer[:n]...)

		// Scan inputBuffer for command lines terminated by \r or \n
		for {
			idx := -1
			for i, b := range inputBuffer {
				if b == '\r' || b == '\n' {
					idx = i
					break
				}
			}

			if idx == -1 {
				break
			}

			// Extract line bytes up to the delimiter
			cmdLineBytes := inputBuffer[:idx]

			// Discard the current delimiter and any adjacent trailing delimiters (e.g. \r\n)
			skip := 1
			for idx+skip < len(inputBuffer) && (inputBuffer[idx+skip] == '\r' || inputBuffer[idx+skip] == '\n') {
				skip++
			}
			inputBuffer = inputBuffer[idx+skip:]

			cmdLine := string(cmdLineBytes)
			log.Printf("[%s] <- %q", clientName, cmdLine)

			// Execute command (we feed the raw cmdLine plus carriage return to mimic real ELM327 newline behavior)
			resp := session.HandleCommand(cmdLine + "\r")

			log.Printf("[%s] -> %q", clientName, resp)
			_, err = conn.Write([]byte(resp))
			if err != nil {
				log.Printf("[%s] Write error: %v", clientName, err)
				return
			}
		}
	}

	log.Printf("[%s] Client disconnected", clientName)
}

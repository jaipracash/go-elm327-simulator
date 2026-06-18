package bluetooth

import (
	"log"
	"net"
	"time"

	"vehicle-simulator/vehicle"

	"go.bug.st/serial"
)

// Server handles listening on network TCP ports or physical/virtual Serial COM ports.
type Server struct {
	Vehicle    *vehicle.Vehicle
	TCPAddr    string
	SerialPort string
	SerialBaud int
}

// NewServer initializes a server structure.
func NewServer(v *vehicle.Vehicle) *Server {
	return &Server{
		Vehicle:    v,
		SerialBaud: 38400, // standard ELM327 baud rate
	}
}

// StartTCP sets up the network TCP listener on the configured address.
func (s *Server) StartTCP() error {
	listener, err := net.Listen("tcp", s.TCPAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("TCP Server active on %s. Direct your client here for testing.", s.TCPAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("TCP Accept error: %v", err)
			continue
		}

		go HandleClient(conn, s.Vehicle, conn.RemoteAddr().String())
	}
}

// StartSerial initiates the serial port reader loop, reconnecting automatically if the port closes.
func (s *Server) StartSerial() {
	mode := &serial.Mode{
		BaudRate: s.SerialBaud,
	}

	for {
		log.Printf("Opening Serial Port %s...", s.SerialPort)
		port, err := serial.Open(s.SerialPort, mode)
		if err != nil {
			log.Printf("Failed to open serial port %s: %v. Retrying in 5s...", s.SerialPort, err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("Serial Server active on %s at %d baud. Pair your device and connect.", s.SerialPort, s.SerialBaud)
		HandleClient(port, s.Vehicle, s.SerialPort)

		// Brief delay to prevent tight loop in case of continuous connect errors
		time.Sleep(1 * time.Second)
	}
}

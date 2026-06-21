package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"vehicle-simulator/bluetooth"
	"vehicle-simulator/vehicle"
)

var (
	tcpAddr    = flag.String("tcp", ":35000", "TCP address to listen on (e.g. :35000 or localhost:35000)")
	serialPort = flag.String("serial", "", "COM port to listen on (e.g. COM3 or COM4)")
	serialBaud = flag.Int("baud", 38400, "Serial port baud rate")
)

func main() {
	flag.Parse()

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	// Create and initialize vehicle state
	v := vehicle.NewVehicle()

	// Start background driving simulator
	go v.SimulateDriving()

	// Initialize the communication server
	srv := bluetooth.NewServer(v)
	srv.TCPAddr = *tcpAddr
	srv.SerialBaud = *serialBaud

	log.Println("==========================================================")
	log.Println("     OBD-II VEHICLE ECU + ELM327 SIMULATOR ACTIVE         ")
	log.Println("==========================================================")

	// Start listeners based on flags
	startedAny := false
	if srv.TCPAddr != "" {
		go func() {
			if err := srv.StartTCP(); err != nil {
				log.Printf("[Error] TCP Server failed to start: %v", err)
			}
		}()
		startedAny = true
	}

	if *serialPort != "" {
		srv.SerialPort = *serialPort
		go srv.StartSerial()
		startedAny = true
	}

	if !startedAny {
		log.Println("[Warning] No listeners started! Specify --tcp or --serial.")
	}

	// Wait brief moment to let listeners print initial state
	time.Sleep(100 * time.Millisecond)

	printHelp()
	fmt.Println("\n>>> Starting Live Dashboard automatically in 3 seconds... <<<")
	time.Sleep(3 * time.Second)

	// Start interactive console command loop for fault injection
	reader := bufio.NewReader(os.Stdin)
	runLiveDashboard(v, reader)
	for {
		fmt.Print("\nSimulator Command > ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Stdin read error: %v", err)
			break
		}

		cmd := strings.TrimSpace(strings.ToLower(input))
		if cmd == "" {
			continue
		}

		switch cmd {
		case "m":
			v.ToggleMisfire()
		case "c":
			v.ToggleCatalystFailure()
		case "o":
			v.ToggleOverheating()
		case "b":
			v.ToggleLowBattery()
		case "e":
			v.ToggleEngine()
		case "s":
			printStatus(v)
		case "l", "live":
			runLiveDashboard(v, reader)
		case "h":
			printHelp()
		case "q":
			log.Println("Stopping simulator. Goodbye!")
			return
		default:
			fmt.Printf("Unknown command %q. Type 'h' for help.\n", cmd)
		}
	}
}

func printHelp() {
	fmt.Println("\n==========================================================")
	fmt.Println("             INTERACTIVE SIMULATOR COMMANDS               ")
	fmt.Println("==========================================================")
	fmt.Println("  [m] Toggle Engine Misfire (DTC P0301 - Pending)")
	fmt.Println("  [c] Toggle Catalyst Failure (DTC P0133, P0420 - Stored)")
	fmt.Println("  [o] Toggle Engine Overheating (110°C coolant temp)")
	fmt.Println("  [b] Toggle Low Battery voltage (11.5V)")
	fmt.Println("  [e] Toggle Engine Running (RPM = 0 / 900)")
	fmt.Println("  [s] Print current vehicle metrics & DTC dashboard")
	fmt.Println("  [l] Start Live Dashboard mode (updates in real-time)")
	fmt.Println("  [h] Print this help menu")
	fmt.Println("  [q] Quit the simulator")
	fmt.Println("==========================================================")
}

func printStatus(v *vehicle.Vehicle) {
	v.PrintDashboard()
}

func runLiveDashboard(v *vehicle.Vehicle, reader *bufio.Reader) {
	v.Lock()
	v.IsLiveActive = true
	v.Unlock()

	v.StartEngine()

	defer func() {
		v.StopEngine()

		v.Lock()
		v.IsLiveActive = false
		v.Unlock()
	}()

	// Clear the screen first to have a clean slate
	fmt.Print("\033[H\033[2J")

	stopChan := make(chan struct{})
	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				// Move cursor to top-left and print the dashboard
				fmt.Print("\033[H")
				v.PrintDashboard()
				fmt.Println("  [LIVE MODE] Press ENTER to stop live view...")
			}
		}
	}()

	// Initial print so it displays immediately
	fmt.Print("\033[H")
	v.PrintDashboard()
	fmt.Println("  [LIVE MODE] Press ENTER to stop live view...")

	// Block until the user presses ENTER
	_, _ = reader.ReadString('\n')

	// Signal the printing goroutine to stop
	close(stopChan)
	<-doneChan

	// Clear screen one more time and print status of returning
	fmt.Print("\033[H\033[2J")
	fmt.Println("Returned to standard command mode.")
}

# OBD-II / ELM327 ECU Simulator

A simple, lightweight, and interactive OBD-II ECU and ELM327 adapter simulator written in Go. It simulates real-time vehicle telemetry (RPM, Speed, Coolant Temp, etc.) and allows interactive fault code (DTC) injection for testing OBD-II scanner software or apps.

---

## Features
- **TCP & Serial support**: Listen on a local TCP port or simulate a Serial/COM port connection.
- **Dynamic Driving Simulation**: Telemetry metrics change dynamically in a background loop.
- **Fault Injection**: Toggle DTCs (stored and pending) and check engine light (MIL) state dynamically from the console.
- **Interactive Dashboard**: View real-time simulated stats directly in the simulator terminal.

---

## How to Run

### Prerequisite
Make sure you have [Go](https://go.dev/dl/) installed.

### Start the Simulator
Run the simulator using `go run`:
```bash
go run cmd/main.go
```

By default, this starts a TCP server listening on port `35000` (`localhost:35000`).

### Command Line Flags
You can customize the listener settings using the following flags:
- `-tcp` : TCP address to listen on (default is `:35000`). Set to empty string `""` to disable TCP.
- `-serial` : COM/Serial port to listen on (e.g. `COM3` on Windows or `/dev/ttyUSB0` on Linux).
- `-baud` : Serial port baud rate (default is `38400`).

**Example with customized TCP and Serial:**
```bash
go run cmd/main.go -tcp :35000 -serial COM3 -baud 38400
```

---

## Interactive Console Commands

When the simulator is running, type any of the following commands into the terminal and press **Enter** to control the simulator state:

| Command | Action | Description |
| :---: | :--- | :--- |
| **`m`** | Toggle Engine Misfire | Toggles pending DTC **`P0301`** (Mode 07) |
| **`c`** | Toggle Catalyst Failure | Toggles stored DTCs **`P0133`** & **`P0420`** (Mode 03) and turns Check Engine Light (MIL) **ON/OFF** |
| **`o`** | Toggle Engine Overheating | Forces coolant temperature to **110°C** |
| **`b`** | Toggle Low Battery | Drops battery voltage to **11.5V** |
| **`e`** | Toggle Engine Status | Starts/stops engine (switches RPM between `900` and `0`) |
| **`s`** | Print Vehicle Status | Displays the current metrics dashboard & DTC list |
| **`l`** | Start Live Dashboard | Starts dynamic real-time dashboard updates (updates every 500ms). Press ENTER to exit. |
| **`h`** | Help Menu | Prints list of available console commands |
| **`q`** | Quit | Safely shuts down the simulator |

---

## Supported OBD-II commands

The simulator responds to standard OBD-II/ELM327 commands:

### Mode 01: Show Current Data (PIDs)
- **`0101`** - Malfunction Indicator Lamp (MIL) status & active DTC count
- **`0104`** - Calculated Engine Load
- **`0105`** - Engine Coolant Temperature
- **`010C`** - Engine RPM
- **`010D`** - Vehicle Speed
- **`010E`** - Ignition Timing Advance
- **`010F`** - Intake Air Temperature
- **`0111`** - Absolute Throttle Position
- **`0114`** - Oxygen Sensor Voltage
- **`012F`** - Fuel Level Input

### Diagnostic Trouble Codes (DTCs)
- **`03`** - Request Stored DTCs (returns `P0133`/`P0420` if Catalyst Failure is enabled)
- **`07`** - Request Pending DTCs (returns `P0301` if Engine Misfire is enabled)
- **`04`** - Clear Diagnostic Trouble Codes / MIL Status

### ELM327 Commands
- **`ATRV`** - Read Battery Voltage

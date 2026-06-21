package vehicle

import (
	"fmt"
	"strings"
	"sync"
)

// Vehicle holds the dynamic state of the simulated vehicle.
type Vehicle struct {
	mu           sync.RWMutex
	RPM          int      // engine revolutions per minute (010C)
	Speed        int      // vehicle speed in km/h (010D)
	Temp         int      // engine coolant temperature in Celsius (0105)
	Fuel         int      // fuel level percentage (012F)
	Battery      float64  // battery voltage in Volts (ATRV)
	Throttle     int      // throttle position percentage (0111)
	Load         int      // calculated engine load percentage (0104)
	Intake       int      // intake air temperature in Celsius (010F)
	Timing       int      // ignition timing advance in degrees (010E)
	MILOn        bool     // Malfunction Indicator Lamp status (0101)
	StoredCodes  []string // active trouble codes (03)
	PendingCodes []string // pending trouble codes (07)

	// Simulation control states
	isOverheating bool
	isLowBattery  bool
	isEngineOff   bool
	IsLiveActive  bool
}

// NewVehicle creates a vehicle with standard initialization parameters.
func NewVehicle() *Vehicle {
	return &Vehicle{
		RPM:          0,
		Speed:        0,
		Temp:         25,
		Fuel:         80,
		Battery:      12.2,
		Throttle:     0,
		Load:         0,
		Intake:       25,
		Timing:       0,
		MILOn:        false,
		StoredCodes:  []string{},
		PendingCodes: []string{},
		isEngineOff:  true,
	}
}

// StartEngine starts the engine and sets default idling parameters.
func (v *Vehicle) StartEngine() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.isEngineOff = false
	v.RPM = 900
	v.Battery = 13.8
	fmt.Println("[Simulator] Engine status: STARTED.")
}

// StopEngine shuts down the engine and resets parameters.
func (v *Vehicle) StopEngine() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.isEngineOff = true
	v.RPM = 0
	v.Speed = 0
	v.Throttle = 0
	v.Load = 0
	v.Timing = 0
	v.Temp = 25
	v.Battery = 12.2
	fmt.Println("[Simulator] Engine status: SHUT DOWN.")
}

// ToggleMisfire toggles the pending trouble code P0301.
func (v *Vehicle) ToggleMisfire() {
	v.mu.Lock()
	defer v.mu.Unlock()

	hasP0301 := false
	var updated []string
	for _, code := range v.PendingCodes {
		if code == "P0301" {
			hasP0301 = true
		} else {
			updated = append(updated, code)
		}
	}

	if hasP0301 {
		v.PendingCodes = updated
		fmt.Println("[Simulator] Injected Fault: Engine Misfire (P0301) REMOVED.")
	} else {
		v.PendingCodes = append(v.PendingCodes, "P0301")
		fmt.Println("[Simulator] Injected Fault: Engine Misfire (P0301) ENABLED.")
	}
}

// ToggleCatalystFailure toggles the stored trouble codes P0133 and P0420 and triggers the MIL.
func (v *Vehicle) ToggleCatalystFailure() {
	v.mu.Lock()
	defer v.mu.Unlock()

	hasP0420 := false
	var updated []string
	for _, code := range v.StoredCodes {
		if code == "P0420" || code == "P0133" {
			hasP0420 = true
		} else {
			updated = append(updated, code)
		}
	}

	if hasP0420 {
		v.StoredCodes = updated
		v.MILOn = len(v.StoredCodes) > 0
		fmt.Println("[Simulator] Injected Fault: Catalyst & O2 Failure (P0133, P0420) REMOVED.")
	} else {
		// Append if not already present
		foundP0133 := false
		foundP0420 := false
		for _, code := range v.StoredCodes {
			if code == "P0133" {
				foundP0133 = true
			}
			if code == "P0420" {
				foundP0420 = true
			}
		}
		if !foundP0133 {
			v.StoredCodes = append(v.StoredCodes, "P0133")
		}
		if !foundP0420 {
			v.StoredCodes = append(v.StoredCodes, "P0420")
		}
		v.MILOn = true
		fmt.Println("[Simulator] Injected Fault: Catalyst & O2 Failure (P0133, P0420) ENABLED.")
	}
}

// ToggleOverheating toggles whether the simulator heats the engine to 110C.
func (v *Vehicle) ToggleOverheating() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.isOverheating = !v.isOverheating
	if v.isOverheating {
		fmt.Println("[Simulator] Injected State: Engine Overheating (110°C) ENABLED.")
	} else {
		fmt.Println("[Simulator] Injected State: Engine Overheating REMOVED.")
	}
}

// ToggleLowBattery toggles whether the battery voltage drops to 11.5V.
func (v *Vehicle) ToggleLowBattery() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.isLowBattery = !v.isLowBattery
	if v.isLowBattery {
		fmt.Println("[Simulator] Injected State: Low Battery (11.5V) ENABLED.")
	} else {
		fmt.Println("[Simulator] Injected State: Low Battery REMOVED.")
	}
}

// ToggleEngine toggles whether the engine is running or completely turned off.
func (v *Vehicle) ToggleEngine() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.isEngineOff = !v.isEngineOff
	if v.isEngineOff {
		v.RPM = 0
		v.Speed = 0
		v.Throttle = 0
		v.Load = 0
		fmt.Println("[Simulator] Engine status: SHUT DOWN.")
	} else {
		v.RPM = 900
		fmt.Println("[Simulator] Engine status: STARTED.")
	}
}

// PrintDashboard prints the formatted vehicle state metrics to the CLI console.
func (v *Vehicle) PrintDashboard() {
	v.mu.RLock()
	defer v.mu.RUnlock()

	milStr := "OFF"
	if v.MILOn {
		milStr = "ON (Check Engine)"
	}

	engineStr := "RUNNING"
	if v.isEngineOff {
		engineStr = "OFF"
	}

	fmt.Println("\n==========================================================")
	fmt.Println("                 VEHICLE STATE DASHBOARD                  ")
	fmt.Println("==========================================================")
	fmt.Printf("  Engine Status: %-15s   Battery Voltage: %.1f V\n", engineStr, v.Battery)
	fmt.Printf("  Engine Speed:  %-15d   Vehicle Speed:   %d km/h\n", v.RPM, v.Speed)
	fmt.Printf("  Coolant Temp:  %-15d   Engine Load:     %d%%\n", v.Temp, v.Load)
	fmt.Printf("  Throttle Pos:  %-15d   Fuel Level:      %d%%\n", v.Throttle, v.Fuel)
	fmt.Printf("  MIL (Check Eng): %-13s   Intake Temp:     %d C\n", milStr, v.Intake)
	fmt.Printf("  Stored DTCs:   %s\n", strings.Join(v.StoredCodes, ", "))
	fmt.Printf("  Pending DTCs:  %s\n", strings.Join(v.PendingCodes, ", "))
	fmt.Println("==========================================================")
}

// GetSimulatorOverrides returns overheating, low battery, and engine run states.
func (v *Vehicle) GetSimulatorOverrides() (overheating, lowBatt, engineOff bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.isOverheating, v.isLowBattery, v.isEngineOff
}

// Lock acquires the write lock on the vehicle state.
func (v *Vehicle) Lock() {
	v.mu.Lock()
}

// Unlock releases the write lock on the vehicle state.
func (v *Vehicle) Unlock() {
	v.mu.Unlock()
}

// RLock acquires the read lock on the vehicle state.
func (v *Vehicle) RLock() {
	v.mu.RLock()
}

// RUnlock releases the read lock on the vehicle state.
func (v *Vehicle) RUnlock() {
	v.mu.RUnlock()
}


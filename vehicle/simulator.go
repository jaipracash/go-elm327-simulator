package vehicle

import (
	"math/rand"
	"time"
)

// SimulateDriving updates vehicle parameters periodically to simulate driving conditions.
func (v *Vehicle) SimulateDriving() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for range ticker.C {
		v.mu.Lock()

		// 1. Check if engine is shut down
		if v.isEngineOff {
			v.RPM = 0
			v.Speed = 0
			v.Throttle = 0
			v.Load = 0
			v.Temp = 25
			v.Intake = 25
			v.Timing = 0
		} else {
			// Engine is running
			v.RPM += r.Intn(300) - 130
			if v.RPM < 750 {
				v.RPM = 750
			}
			if v.RPM > 6000 {
				v.RPM = 5500
			}

			v.Speed += r.Intn(5) - 2
			if v.Speed < 0 {
				v.Speed = 0
			}
			if v.Speed > 120 {
				v.Speed = 110
			}

			v.Throttle = int(float64(v.RPM-750) / 5250.0 * 100.0)
			if v.Throttle < 5 {
				v.Throttle = 5
			}
			if v.Throttle > 95 {
				v.Throttle = 95
			}

			v.Load = 20 + r.Intn(10) + (v.Speed / 2)
			if v.Load > 99 {
				v.Load = 99
			}

			// 2. Check overheating override
			if v.isOverheating {
				v.Temp = 110
			} else {
				if v.Temp < 90 {
					v.Temp += r.Intn(2)
				} else {
					v.Temp += r.Intn(3) - 1
				}
			}

			v.Intake = 35 + r.Intn(3) - 1
			v.Timing = 15 + r.Intn(5)
		}

		// 3. Check battery override
		if v.isLowBattery {
			v.Battery = 11.5
		} else {
			if !v.isEngineOff {
				v.Battery = 13.8 + float64(r.Intn(4))/10.0
			} else {
				v.Battery = 12.2 + float64(r.Intn(3))/10.0
			}
		}

		v.mu.Unlock()
	}
}

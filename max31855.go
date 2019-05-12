package max31855

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// MAX31855 contains data for the MAX31855 sensor
type MAX31855 struct {
	device       *os.File
	Thermocouple float64
	Internal     float64
	Timestamp    time.Time
}

// Open establishes the SPI connection with the sensor.
func (m *MAX31855) Open(name string) error {
	var err error
	m.device, err = os.Open(name)
	return err
}

// Read gets the latest values from the sensor,
// and updates the thermocouple, internal and Timestamp fields.
func (m *MAX31855) Read() error {
	data := make([]byte, 4)
	numBytes, err := m.device.Read(data)
	m.Timestamp = time.Now()

	if err != nil {
		return err
	}
	if numBytes != 4 {
		return fmt.Errorf("%d bytes read instead of 4", numBytes)
	}

	// check error flags
	if (data[3] & 0x01) == 1 {
		return errors.New("open circuit")
	} else if (data[3] & 0x02) == 1 {
		return errors.New("short circuit to GND")
	} else if (data[3] & 0x04) == 1 {
		return errors.New("short circuit to VCC")
	}

	// data seems okay - calculate temps.
	thermocoupleWord := ((uint16(data[0]) << 8) | uint16(data[1])) >> 2
	m.Thermocouple = float64(int16(thermocoupleWord)) * 0.25

	internalWord := ((uint16(data[2]) << 8) | uint16(data[3])) >> 4
	m.Internal = float64(int16(internalWord)) * 0.0625
	return nil
}

// Close ends the SPI connection to thermocouple
func (m *MAX31855) Close() error {
	return m.device.Close()
}

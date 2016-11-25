// Copyright 2016 Shannon Wynter. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package aurora

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// Inverter structure for connecting to an inverter
type Inverter struct {
	Conn    io.ReadWriter
	Address byte
}

// ErrCRCFailure is returned whenever the data read in from the serial port might
// have been corrupted en route and no longer matches the crc
var ErrCRCFailure = errors.New("CRC Failure")

// Communicate encodes and transmits given commands returning a response having
// checked the CRC and transmission state if applicable
func (i *Inverter) Communicate(command Command, args ...Argument) ([]byte, error) {
	outputBuffer := outputPayload{
		Payload: [8]byte{i.Address, byte(command), 32, 32, 32, 32, 32, 32},
	}
	lastIndex := 1
	for index, arg := range args {
		if index > 5 {
			break
		}
		lastIndex = index + 2
		outputBuffer.Payload[lastIndex] = arg.Byte()
	}

	// Inverter expects 0 terminated instructions
	if lastIndex < 7 {
		outputBuffer.Payload[lastIndex+1] = 0
	}

	outputBuffer.CRC = calculateCRC(outputBuffer.Payload[:])

	if err := binary.Write(i.Conn, binary.LittleEndian, outputBuffer); err != nil {
		return nil, err
	}

	inputBuffer := inputPayload{}
	if err := binary.Read(i.Conn, binary.LittleEndian, &inputBuffer); err != nil {
		return nil, err
	}

	if crc := calculateCRC(inputBuffer.Payload[:]); crc != inputBuffer.CRC {
		return nil, ErrCRCFailure
	}

	if command == GetPartNumber || command == GetSerialNumber {
		return inputBuffer.Payload[:], nil
	}

	if inputBuffer.Payload[0] != 0 {
		return nil, errors.New(TransmissionState(inputBuffer.Payload[0]).String())
	}

	if command == GetState {
		return inputBuffer.Payload[1:], nil
	}

	return inputBuffer.Payload[2:], nil
}

// CommunicateVar works much like Communicate but expects an interface to write the response to
func (i *Inverter) CommunicateVar(v interface{}, command Command, args ...Argument) error {
	result, err := i.Communicate(command, args...)
	if err != nil {
		return err
	}
	return binary.Read(bytes.NewReader(result), binary.BigEndian, v)
}

func calculateCRC(input []byte) uint16 {
	crc := uint16(0xffff)
	for _, chr := range input {
		for i, data := 0, chr; i < 8; i, data = i+1, data>>1 {
			if (crc&0x0001)^uint16(data&0x01) == 1 {
				crc = (crc >> 1) ^ 0x8408
			} else {
				crc = crc >> 1
			}
		}
	}

	return ^crc
}

// CommCheck calls the simplest command supported by the inverter "GetVersion" just
// as a quick check to make sure it's connected and working.
// You might want to wrap a deadline around this call.
func (i *Inverter) CommCheck() error {
	_, err := i.Communicate(GetVersion)
	return err
}

// State returns the current state for the inverter
func (i *Inverter) State() (*State, error) {
	var state State
	err := i.CommunicateVar(&state, GetState)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

// Last4Alarms returns the last 4 alarm states
func (i *Inverter) Last4Alarms() ([]AlarmState, error) {
	alarms := make([]AlarmState, 4)
	err := i.CommunicateVar(&alarms, GetLast4Alarms)
	if err != nil {
		return nil, err
	}
	return alarms, nil
}

// PartNumber returns the inverters part number
func (i *Inverter) PartNumber() (string, error) {
	result, err := i.Communicate(GetPartNumber)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// SerialNumber returns the inverters serial number
func (i *Inverter) SerialNumber() (string, error) {
	result, err := i.Communicate(GetSerialNumber)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// Version returns the inverters version
func (i *Inverter) Version() (*Version, error) {
	var version Version
	if err := i.CommunicateVar(&version, GetVersion); err != nil {
		return nil, err
	}
	return &version, nil
}

// ManufactureDate returns the inverters date of manufacture
func (i *Inverter) ManufactureDate() (string, string, error) {
	result, err := i.Communicate(GetManufacturingDate)
	if err != nil {
		return "", "", err
	}
	year := string(result[2:4])
	week := string(result[0:2])
	return year, week, nil
}

// FirmwareVersion returns the inverters firmware version
func (i *Inverter) FirmwareVersion() (string, error) {
	result, err := i.Communicate(GetFirmwareVersion)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%c.%c.%c.%c",
		rune(result[0]),
		rune(result[1]),
		rune(result[2]),
		rune(result[3]),
	), nil
}

// Configuration returns the current configuration state from the inverter
func (i *Inverter) Configuration() (ConfigurationState, error) {
	result, err := i.Communicate(GetConfiguration)
	if err != nil {
		return ConfigurationState(255), err
	}
	return ConfigurationState(result[0]), nil
}

// GetCumulatedEnergy returns the cumulated energy for a given period
func (i *Inverter) GetCumulatedEnergy(period CumulationPeriod) (uint32, error) {
	result, err := i.Communicate(GetCumulatedEnergy, period)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(result), nil
}

// DailyEnergy returns the daily cumulated energy
func (i *Inverter) DailyEnergy() (uint32, error) {
	return i.GetCumulatedEnergy(CumulatedDaily)
}

// WeeklyEnergy returns the weekly cumulated energy
func (i *Inverter) WeeklyEnergy() (uint32, error) {
	return i.GetCumulatedEnergy(CumulatedWeekly)
}

// MonthlyEnergy returns the monthly cumulated energy
func (i *Inverter) MonthlyEnergy() (uint32, error) {
	return i.GetCumulatedEnergy(CumulatedMonthly)
}

// YearlyEnergy returns the yearly cumulated energy
func (i *Inverter) YearlyEnergy() (uint32, error) {
	return i.GetCumulatedEnergy(CumulatedYearly)
}

// TotalEnergy returns the total cumulated energy
func (i *Inverter) TotalEnergy() (uint32, error) {
	return i.GetCumulatedEnergy(CumulatedTotal)
}

// PartialEnergy returns the cumulated energy since last reset
func (i *Inverter) PartialEnergy() (uint32, error) {
	return i.GetCumulatedEnergy(CumulatedPartial)
}

// GetDSPData returns data for various DSParameters
func (i *Inverter) GetDSPData(parameter DSParameter) (float32, error) {
	var f float32
	err := i.CommunicateVar(&f, GetDSP, parameter)
	return f, err
}

// Frequency returns the operating frequency
func (i *Inverter) Frequency() (float32, error) {
	return i.GetDSPData(DSPFrequency)
}

// GridVoltage returns the voltage from the grid
func (i *Inverter) GridVoltage() (float32, error) {
	return i.GetDSPData(DSPGridVoltage)
}

// GridCurrent returns the amount of current (in amps) being pushed to the grid.
func (i *Inverter) GridCurrent() (float32, error) {
	return i.GetDSPData(DSPGridCurrent)
}

// GridPower returns the amount of power (in watts) being pushed to the grid.
func (i *Inverter) GridPower() (float32, error) {
	return i.GetDSPData(DSPGridPower)
}

// Input1Voltage returns the voltage received on input 1 from your solar array/wind turbine
func (i *Inverter) Input1Voltage() (float32, error) {
	return i.GetDSPData(DSPInput1Voltage)
}

// Input1Current returns the amount of current (in amps) being received from input 1
func (i *Inverter) Input1Current() (float32, error) {
	return i.GetDSPData(DSPInput1Current)
}

// Input2Voltage returns the voltage received on input 2 from your solar array/wind turbine
func (i *Inverter) Input2Voltage() (float32, error) {
	return i.GetDSPData(DSPInput2Voltage)
}

// Input2Current returns the amount of current (in amps) being received from input 2
func (i *Inverter) Input2Current() (float32, error) {
	return i.GetDSPData(DSPInput2Current)
}

// InverterTemperature returns the current temperature of the inverter in celsius
func (i *Inverter) InverterTemperature() (float32, error) {
	return i.GetDSPData(DSPInverterTemperature)
}

// BoosterTemperature returns the current temperature of the booster in celsius
func (i *Inverter) BoosterTemperature() (float32, error) {
	return i.GetDSPData(DSPBoosterTemperature)
}

// Joules returns the amount of power produced in the last 10 seconds as Joules
func (i *Inverter) Joules() (uint16, error) {
	var s uint16
	err := i.CommunicateVar(&s, GetLast10SecEnergy)
	return s, err
}

// GetTime returns the current timestamp from the inverter, returns as a unix epoch based timestamp
func (i *Inverter) GetTime() (time.Time, error) {
	result, err := i.Communicate(GetTime)
	if err != nil {
		return time.Unix(0, 0), err
	}
	return time.Unix(int64(InverterEpochOffset+binary.BigEndian.Uint32(result)), 0), nil
}

// SetTime sets the time in the inverter to the given timestamp.
// Warning: this may result in the resetting of partial counters/cumulaters.
func (i *Inverter) SetTime(t time.Time) error {
	value := uint32(t.Unix() - InverterEpochOffset)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, value)
	bvalue := buf.Bytes()
	_, err := i.Communicate(SetTime, Byte(bvalue[0]), Byte(bvalue[1]), Byte(bvalue[2]), Byte(bvalue[3]))
	return err
}

// GetCounterData returns the value (seconds?) from one of the counters being total, partial, grid, and reset runtimes.
func (i *Inverter) GetCounterData(counter Counter) (uint32, error) {
	result, err := i.Communicate(GetCounters, counter)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(result), nil
}

func (i *Inverter) getDuration(counter Counter) (time.Duration, error) {
	result, err := i.GetCounterData(counter)
	if err != nil {
		return 0, err
	}
	return time.Duration(result) * time.Second, nil
}

// TotalRunTime returns the total runtime for the inverter
func (i *Inverter) TotalRunTime() (time.Duration, error) {
	return i.getDuration(CounterTotal)
}

// PartialRunTime returns the partial runtime of the inverter...
func (i *Inverter) PartialRunTime() (time.Duration, error) {
	return i.getDuration(CounterPartial)
}

// GridRunTime returns the amount of time the inverter has been on grid
func (i *Inverter) GridRunTime() (time.Duration, error) {
	return i.getDuration(CounterGrid)
}

// ResetRunTime resets the counter
func (i *Inverter) ResetRunTime() error {
	_, err := i.GetCounterData(CounterReset)
	return err
}

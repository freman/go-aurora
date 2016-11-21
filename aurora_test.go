// Copyright 2016 Shannon Wynter. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package aurora_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/freman/go-aurora"
)

type mockSerial struct {
	*io.PipeReader
	*io.PipeWriter
}

func mockSerialPair() (io.ReadWriter, io.ReadWriter) {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	return &mockSerial{r1, w2}, &mockSerial{r2, w1}
}

func mockInverterExpect(t *testing.T, in, out []byte) *aurora.Inverter {
	ttys0, ttys1 := mockSerialPair()
	i := &aurora.Inverter{Conn: ttys0, Address: 2}

	go func() {
		// Queue up a regular test as expected
		ttys1.(*mockSerial).expect(t, in, out)

		// Queue up a CRC error
		makeCRCError(t, ttys1)
	}()
	return i
}

func (m *mockSerial) expect(t *testing.T, in, out []byte) {
	tmp := make([]byte, 10)
	c, err := m.Read(tmp)
	if err != nil {
		t.Error(err)
	}
	if c < 10 {
		t.Errorf("Expected %d, got %d", 10, c)
	}

	if !bytes.Equal(in, tmp) {
		t.Errorf("Expected [%x] got [%x]", in, tmp)
	}

	if err := binary.Write(m, binary.LittleEndian, out); err != nil {
		t.Error(err)
	}
	if err := binary.Write(m, binary.LittleEndian, calculateCRC(out)); err != nil {
		t.Error(err)
	}
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

func makeCRCError(t *testing.T, ttys1 io.ReadWriter) {
	tmp := make([]byte, 10)
	c, err := ttys1.Read(tmp)
	if err != nil {
		t.Error(err)
	}
	if l := len(tmp); c < l {
		t.Errorf("Expected %d, got %d", l, c)
	}
	res := []byte{0, 2, 3, 4, 5, 6}
	binary.Write(ttys1, binary.LittleEndian, res)
	binary.Write(ttys1, binary.LittleEndian, calculateCRC(res)+1)
}

func TestCommunicate(t *testing.T) {
	ttys0, ttys1 := mockSerialPair()
	i := &aurora.Inverter{Conn: ttys0}

	// Simplest execution
	go func() {
		tmp := make([]byte, 10)
		c, err := ttys1.Read(tmp)
		if err != nil {
			t.Error(err)
		}
		if l := len(tmp); c < l {
			t.Errorf("Expected %d, got %d", l, c)
		}

		res := []byte{0, 2, 3, 4, 5, 6}
		binary.Write(ttys1, binary.LittleEndian, res)
		binary.Write(ttys1, binary.LittleEndian, calculateCRC(res))
	}()
	i.Communicate(aurora.GetCumulatedEnergy, aurora.CumulatedMonthly)

	// Push more than 10 bytes through, extra byte should be dropped
	go func() {
		tmp := make([]byte, 11)
		c, err := ttys1.Read(tmp)
		if err != nil {
			t.Error(err)
		}
		if c != 10 {
			t.Errorf("Expected %d, got %d", 10, c)
		}

		res := []byte{0, 2, 3, 4, 5, 6}
		binary.Write(ttys1, binary.LittleEndian, res)
		binary.Write(ttys1, binary.LittleEndian, calculateCRC(res))
	}()
	b := aurora.Byte(0x01)
	i.Communicate(aurora.GetCumulatedEnergy, b, b, b, b, b, b, b)

	// Push a bad CRC in the response
	go makeCRCError(t, ttys1)
	_, err := i.Communicate(aurora.GetCumulatedEnergy, aurora.CumulatedMonthly)
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}

	// Push a bad transmission code in the response
	go func() {
		tmp := make([]byte, 10)
		c, err := ttys1.Read(tmp)
		if err != nil {
			t.Error(err)
		}
		if l := len(tmp); c < l {
			t.Errorf("Expected %d, got %d", l, c)
		}

		res := []byte{52, 2, 3, 4, 5, 6}
		binary.Write(ttys1, binary.LittleEndian, res)
		binary.Write(ttys1, binary.LittleEndian, calculateCRC(res))
	}()
	_, err = i.Communicate(aurora.GetCumulatedEnergy, aurora.CumulatedMonthly)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestCommunicateWriteError(t *testing.T) {
	ttys0, ttys1 := mockSerialPair()
	i := &aurora.Inverter{Conn: ttys0}
	go func() {
		tmp := make([]byte, 5)
		c, err := ttys1.Read(tmp)
		if err != nil {
			t.Error(err)
		}
		if c != 5 {
			t.Errorf("Expected %d, got %d", 10, c)
		}
		ttys1.(*mockSerial).PipeReader.Close()
	}()
	_, err := i.Communicate(aurora.GetCumulatedEnergy, aurora.CumulatedMonthly)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestCommunicateReadError(t *testing.T) {
	ttys0, ttys1 := mockSerialPair()
	i := &aurora.Inverter{Conn: ttys0}
	go func() {
		tmp := make([]byte, 10)
		c, err := ttys1.Read(tmp)
		if err != nil {
			t.Error(err)
		}
		if c != 10 {
			t.Errorf("Expected %d, got %d", 10, c)
		}
		ttys1.Write([]byte{0x01, 0x02, 0x03, 0x04})
		ttys1.(*mockSerial).PipeWriter.Close()
	}()
	_, err := i.Communicate(aurora.GetCumulatedEnergy, aurora.CumulatedMonthly)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestCommunicateVarError(t *testing.T) {
	ttys0, ttys1 := mockSerialPair()
	i := &aurora.Inverter{Conn: ttys0}

	resp := make([]byte, 5, 5)
	// Push a bad CRC in the response
	go makeCRCError(t, ttys1)
	err := i.CommunicateVar(&resp, aurora.GetCumulatedEnergy, aurora.CumulatedMonthly)
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestCommCheck(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x3a, 0x00, 0x20, 0x20, 0x20, 0x20, 0x20, 0xc9, 0x59}, []byte{0x00, 0x06, 0x49, 0x4b, 0x4e, 0x4e})
	err := i.CommCheck()
	if err != nil {
		t.Error(err)
	}

	err = i.CommCheck()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestState(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x32, 0x00, 0x20, 0x20, 0x20, 0x20, 0x20, 0x25, 0x87}, []byte{0x00, 0x06, 0x02, 0x07, 0x02, 0x00})
	state, err := i.State()
	if err != nil {
		t.Error(err)
	}

	expectedState := &aurora.State{
		Global:   aurora.GSRun,
		Inverter: aurora.ISRun,
		Channel1: aurora.DCDCInputLow,
		Channel2: aurora.DCDCMPPT,
		Alarm:    aurora.AlarmNone,
	}

	if !reflect.DeepEqual(expectedState, state) {
		t.Errorf("Expected %s got %s", expectedState, state)
	}

	_, err = i.State()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestLast4Alarms(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x56, 0x00, 0x20, 0x20, 0x20, 0x20, 0x20, 0xd6, 0x4c}, []byte{0x00, 0x06, 0x00, 0x00, 0x00, 0x00})
	alarms, err := i.Last4Alarms()
	if err != nil {
		t.Error(err)
	}

	for _, alarm := range alarms {
		if alarm != aurora.AlarmNone {
			t.Errorf("Expected %s, got %s", aurora.AlarmNone, alarm)
		}
	}

	_, err = i.Last4Alarms()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestPartNumber(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x34, 0x00, 0x20, 0x20, 0x20, 0x20, 0x20, 0xe8, 0xdf}, []byte{0x2d, 0x31, 0x32, 0x33, 0x34, 0x2d})
	partNumber, err := i.PartNumber()
	if err != nil {
		t.Error(err)
	}

	if partNumber != "-1234-" {
		t.Errorf("Expected -1234-, got %s", partNumber)
	}

	_, err = i.PartNumber()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestSerialNumber(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x3F, 0x00, 0x20, 0x20, 0x20, 0x20, 0x20, 0x6a, 0xa9}, []byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36})
	serial, err := i.SerialNumber()
	if err != nil {
		t.Error(err)
	}

	if serial != "123456" {
		t.Errorf("Expected 123456, got %s", serial)
	}

	_, err = i.SerialNumber()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestVersion(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x3a, 0x00, 0x20, 0x20, 0x20, 0x20, 0x20, 0xc9, 0x59}, []byte{0x00, 0x06, 0x49, 0x4b, 0x4e, 0x4e})
	version, err := i.Version()
	if err != nil {
		t.Error(err)
	}

	expectedVersion := &aurora.Version{
		Model:       aurora.Product3_6kWIndoor,
		Regulation:  aurora.ProductSpecAS4777,
		Transformer: aurora.InverterTransformerless,
		Type:        aurora.InputPhotovoltaic,
	}

	if !reflect.DeepEqual(expectedVersion, version) {
		t.Errorf("Expected %s got %s", expectedVersion, version)
	}

	_, err = i.Version()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestManufactureDate(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x41, 0x00, 0x20, 0x20, 0x20, 0x20, 0x20, 0x07, 0x3e}, []byte{0x00, 0x06, 0x30, 0x31, 0x31, 0x30})
	year, month, err := i.ManufactureDate()
	if err != nil {
		t.Error(err)
	}

	expectedYear := "10"
	expectedMonth := "01"

	if year != expectedYear {
		t.Errorf("Expected year %s got %s", expectedYear, year)
	}

	if month != expectedMonth {
		t.Errorf("Expected month %s got %s", expectedMonth, month)
	}

	_, _, err = i.ManufactureDate()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestFirmwareVersion(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x48, 0x00, 0x20, 0x20, 0x20, 0x20, 0x20, 0x3e, 0x7f}, []byte{0x00, 0x06, 0x63, 0x31, 0x32, 0x33})
	firmwareVersion, err := i.FirmwareVersion()
	if err != nil {
		t.Error(err)
	}

	expectedVersion := "c.1.2.3"

	if firmwareVersion != expectedVersion {
		t.Errorf("Expected %s got %s", expectedVersion, firmwareVersion)
	}

	_, err = i.FirmwareVersion()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestConfiguration(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x4d, 0x00, 0x20, 0x20, 0x20, 0x20, 0x20, 0x9d, 0x8f}, []byte{0x00, 0x06, 0x00, 0x00, 0x00, 0x00})
	configuration, err := i.Configuration()
	if err != nil {
		t.Error(err)
	}

	expectedConfiguration := aurora.ConfigBoth

	if configuration != expectedConfiguration {
		t.Errorf("Expected %v got %v", expectedConfiguration, configuration)
	}

	_, err = i.Configuration()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestDailyEnergy(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x4e, 0x00, 0x00, 0x20, 0x20, 0x20, 0x20, 0x62, 0x47}, []byte{0x00, 0x06, 0x00, 0x00, 0x30, 0x39})
	energy, err := i.DailyEnergy()
	if err != nil {
		t.Error(err)
	}

	expectedEnergy := uint32(12345)

	if energy != expectedEnergy {
		t.Errorf("Expected %d got %d", expectedEnergy, energy)
	}

	_, err = i.DailyEnergy()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestWeeklyEnergy(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x4e, 0x01, 0x00, 0x20, 0x20, 0x20, 0x20, 0x49, 0x43}, []byte{0x00, 0x06, 0x00, 0x01, 0x51, 0x8f})
	energy, err := i.WeeklyEnergy()
	if err != nil {
		t.Error(err)
	}

	expectedEnergy := uint32(86415)

	if energy != expectedEnergy {
		t.Errorf("Expected %d got %d", expectedEnergy, energy)
	}

	_, err = i.WeeklyEnergy()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestMonthlyEnergy(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x4e, 0x03, 0x00, 0x20, 0x20, 0x20, 0x20, 0x1f, 0x4b}, []byte{0x00, 0x06, 0x00, 0x05, 0xa6, 0xae})
	energy, err := i.MonthlyEnergy()
	if err != nil {
		t.Error(err)
	}

	expectedEnergy := uint32(370350)

	if energy != expectedEnergy {
		t.Errorf("Expected %d got %d", expectedEnergy, energy)
	}

	_, err = i.MonthlyEnergy()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestYearlyEnergy(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x4e, 0x04, 0x00, 0x20, 0x20, 0x20, 0x20, 0xce, 0x57}, []byte{0x00, 0x06, 0x00, 0x44, 0xc1, 0x45})
	energy, err := i.YearlyEnergy()
	if err != nil {
		t.Error(err)
	}

	expectedEnergy := uint32(4505925)

	if energy != expectedEnergy {
		t.Errorf("Expected %d got %d", expectedEnergy, energy)
	}

	_, err = i.YearlyEnergy()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestTotalEnergy(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x4e, 0x05, 0x00, 0x20, 0x20, 0x20, 0x20, 0xe5, 0x53}, []byte{0x00, 0x06, 0x19, 0xc8, 0x87, 0x9e})
	energy, err := i.TotalEnergy()
	if err != nil {
		t.Error(err)
	}

	expectedEnergy := uint32(432572318)

	if energy != expectedEnergy {
		t.Errorf("Expected %d got %d", expectedEnergy, energy)
	}

	_, err = i.TotalEnergy()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestPartialEnergy(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x4e, 0x06, 0x00, 0x20, 0x20, 0x20, 0x20, 0x98, 0x5f}, []byte{0x00, 0x06, 0x00, 0x52, 0x81, 0x86})
	energy, err := i.PartialEnergy()
	if err != nil {
		t.Error(err)
	}

	expectedEnergy := uint32(5407110)

	if energy != expectedEnergy {
		t.Errorf("Expected %d got %d", expectedEnergy, energy)
	}

	_, err = i.PartialEnergy()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestFrequency(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x3b, 0x04, 0x00, 0x20, 0x20, 0x20, 0x20, 0x21, 0xb6}, []byte{0x00, 0x06, 0x42, 0x47, 0xf1, 0xab})
	frequency, err := i.Frequency()
	if err != nil {
		t.Error(err)
	}

	expectedFrequency := float32(49.9860038757324)

	if frequency != expectedFrequency {
		t.Errorf("Expected %f got %f", expectedFrequency, frequency)
	}

	_, err = i.Frequency()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestGridVoltage(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x3b, 0x01, 0x00, 0x20, 0x20, 0x20, 0x20, 0xa6, 0xa2}, []byte{0x00, 0x06, 0x43, 0x6a, 0xe0, 0xfd})
	gridVoltage, err := i.GridVoltage()
	if err != nil {
		t.Error(err)
	}

	expectedGridVoltage := float32(234.878860473633)

	if gridVoltage != expectedGridVoltage {
		t.Errorf("Expected %f got %f", expectedGridVoltage, gridVoltage)
	}

	_, err = i.GridVoltage()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestGridCurrent(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x3b, 0x02, 0x00, 0x20, 0x20, 0x20, 0x20, 0xdb, 0xae}, []byte{0x00, 0x06, 0x3f, 0x6d, 0x5d, 0xad})
	gridCurrent, err := i.GridCurrent()
	if err != nil {
		t.Error(err)
	}

	expectedGridCurrent := float32(0.927210628986359)

	if gridCurrent != expectedGridCurrent {
		t.Errorf("Expected %f got %f", expectedGridCurrent, gridCurrent)
	}

	_, err = i.GridCurrent()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestGridPower(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x3b, 0x03, 0x00, 0x20, 0x20, 0x20, 0x20, 0xf0, 0xaa}, []byte{0x00, 0x06, 0x42, 0x93, 0x61, 0xd8})
	gridPower, err := i.GridPower()
	if err != nil {
		t.Error(err)
	}

	expectedGridPower := float32(73.6911010742188)

	if gridPower != expectedGridPower {
		t.Errorf("Expected %f got %f", expectedGridPower, gridPower)
	}

	_, err = i.GridPower()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestInput1Voltage(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x3b, 0x17, 0x00, 0x20, 0x20, 0x20, 0x20, 0xec, 0xf8}, []byte{0x00, 0x06, 0x42, 0x81, 0xd2, 0xb0})
	input1Voltage, err := i.Input1Voltage()
	if err != nil {
		t.Error(err)
	}

	expectedInput1Voltage := float32(64.9114990234375)

	if input1Voltage != expectedInput1Voltage {
		t.Errorf("Expected %f got %f", expectedInput1Voltage, input1Voltage)
	}

	_, err = i.Input1Voltage()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestInput1Current(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x3b, 0x19, 0x00, 0x20, 0x20, 0x20, 0x20, 0x4e, 0xc1}, []byte{0x00, 0x06, 0x3c, 0x99, 0xba, 0x86})
	input1Current, err := i.Input1Current()
	if err != nil {
		t.Error(err)
	}

	expectedInput1Current := float32(0.0187656991183758)

	if input1Current != expectedInput1Current {
		t.Errorf("Expected %f got %f", expectedInput1Current, input1Current)
	}

	_, err = i.Input1Current()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestInput2Voltage(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x3b, 0x1a, 0x00, 0x20, 0x20, 0x20, 0x20, 0x33, 0xcd}, []byte{0x00, 0x06, 0x43, 0x89, 0xa4, 0xd7})
	input2Voltage, err := i.Input2Voltage()
	if err != nil {
		t.Error(err)
	}

	expectedInput2Voltage := float32(275.287811279297)

	if input2Voltage != expectedInput2Voltage {
		t.Errorf("Expected %f got %f", expectedInput2Voltage, input2Voltage)
	}

	_, err = i.Input2Voltage()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestInput2Current(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x3b, 0x1b, 0x00, 0x20, 0x20, 0x20, 0x20, 0x18, 0xc9}, []byte{0x00, 0x06, 0x3e, 0xc2, 0x09, 0x17})
	input2Current, err := i.Input2Current()
	if err != nil {
		t.Error(err)
	}

	expectedInput2Current := float32(0.378975600004196)

	if input2Current != expectedInput2Current {
		t.Errorf("Expected %f got %f", expectedInput2Current, input2Current)
	}

	_, err = i.Input2Current()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestInverterTemperature(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x3b, 0x15, 0x00, 0x20, 0x20, 0x20, 0x20, 0xba, 0xf0}, []byte{0x00, 0x06, 0x42, 0x7c, 0x0f, 0xde})
	inverterTemperature, err := i.InverterTemperature()
	if err != nil {
		t.Error(err)
	}

	expectedInverterTemperature := float32(63.015495300293)

	if inverterTemperature != expectedInverterTemperature {
		t.Errorf("Expected %f got %f", expectedInverterTemperature, inverterTemperature)
	}

	_, err = i.InverterTemperature()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestBoosterTemperature(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x3b, 0x16, 0x00, 0x20, 0x20, 0x20, 0x20, 0xc7, 0xfc}, []byte{0x00, 0x06, 0x42, 0x60, 0x74, 0xbb})
	boosterTemperature, err := i.BoosterTemperature()
	if err != nil {
		t.Error(err)
	}

	expectedBoosterTemperature := float32(56.1139945983887)

	if boosterTemperature != expectedBoosterTemperature {
		t.Errorf("Expected %f got %f", expectedBoosterTemperature, boosterTemperature)
	}

	_, err = i.BoosterTemperature()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestJoules(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x4c, 0x00, 0x20, 0x20, 0x20, 0x20, 0x20, 0x48, 0x10}, []byte{0x00, 0x06, 0x00, 0x52, 0x00, 0x00})
	joules, err := i.Joules()
	if err != nil {
		t.Error(err)
	}

	expectedJoules := uint16(82)

	if joules != expectedJoules {
		t.Errorf("Expected %d got %d", expectedJoules, joules)
	}

	_, err = i.Joules()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestSetTime(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x47, 0x1f, 0xc4, 0x15, 0xff, 0x00, 0x20, 0x5b, 0xdb}, []byte{0x00, 0x06, 0x01, 0x22, 0x00, 0x64})
	expectedTime, err := time.Parse(time.RFC3339, "2016-11-21T00:06:23+10:00")
	if err != nil {
		t.Errorf("Unable to parse time for test: %v", err)
	}

	err = i.SetTime(expectedTime)
	if err != nil {
		t.Errorf("Unexpected error setting time: %v", err)
	}

	err = i.SetTime(expectedTime)
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestGetTime(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x46, 0x00, 0x20, 0x20, 0x20, 0x20, 0x20, 0x1f, 0xf9}, []byte{0x00, 0x06, 0x1f, 0xc4, 0x15, 0xff})
	iTime, err := i.GetTime()
	if err != nil {
		t.Error(err)
	}

	expectedTime, err := time.Parse(time.RFC3339, "2016-11-21T00:06:23+10:00")
	if err != nil {
		t.Errorf("Unable to parse time for test: %v", err)
	}

	if !iTime.Equal(expectedTime) {
		t.Errorf("Expected %v got %v", expectedTime, iTime)
	}

	_, err = i.GetTime()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestTotalRunTime(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x50, 0x00, 0x00, 0x20, 0x20, 0x20, 0x20, 0x8a, 0x74}, []byte{0x00, 0x06, 0x00, 0x00, 0x00, 0x64})
	runTime, err := i.TotalRunTime()
	if err != nil {
		t.Error(err)
	}

	expectedRunTime := time.Duration(100 * time.Second)

	if runTime != expectedRunTime {
		t.Errorf("Expected %v got %v", expectedRunTime, runTime)
	}

	_, err = i.TotalRunTime()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestPartialRunTime(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x50, 0x01, 0x00, 0x20, 0x20, 0x20, 0x20, 0xa1, 0x70}, []byte{0x00, 0x06, 0x00, 0x00, 0x10, 0x64})
	runTime, err := i.PartialRunTime()
	if err != nil {
		t.Error(err)
	}

	expectedRunTime := time.Duration(4196 * time.Second)

	if runTime != expectedRunTime {
		t.Errorf("Expected %v got %v", expectedRunTime, runTime)
	}

	_, err = i.PartialRunTime()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestGridRunTime(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x50, 0x02, 0x00, 0x20, 0x20, 0x20, 0x20, 0xdc, 0x7c}, []byte{0x00, 0x06, 0x00, 0x00, 0x10, 0x65})
	runTime, err := i.GridRunTime()
	if err != nil {
		t.Error(err)
	}

	expectedRunTime := time.Duration(4197 * time.Second)

	if runTime != expectedRunTime {
		t.Errorf("Expected %v got %v", expectedRunTime, runTime)
	}

	_, err = i.GridRunTime()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

func TestResetRunTime(t *testing.T) {
	i := mockInverterExpect(t, []byte{0x02, 0x50, 0x03, 0x00, 0x20, 0x20, 0x20, 0x20, 0xf7, 0x78}, []byte{0x00, 0x06, 0x00, 0x00, 0x00, 0x00})
	err := i.ResetRunTime()
	if err != nil {
		t.Error(err)
	}

	err = i.ResetRunTime()
	if err != aurora.ErrCRCFailure {
		t.Errorf("Expected %v got %v", aurora.ErrCRCFailure, err)
	}
}

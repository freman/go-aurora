package aurora_test

import (
	"testing"

	"github.com/freman/go-aurora"
)

func TestStateString(t *testing.T) {
	state := &aurora.State{}
	str := state.String()

	if str != "Global: Sending Parameters, Inverter: Stand By, Channel1: DcDc OFF, Channel2: DcDc OFF, Alarm: No Alarm" {
		t.Errorf("Unexpected string returned: %s", str)
	}
}

func TestVersionString(t *testing.T) {
	version := &aurora.Version{}
	str := version.String()

	if str != "Model: Unknown Product (0), Regulation: Unknown ProductSpec (0), Transformer: Unknown InverterType (0), Type: Unknown InputType (0)" {
		t.Errorf("Unexpected string returned: %s", str)
	}

	version.Model = aurora.Product3_6kWOutdoor
	version.Regulation = aurora.ProductSpecAS4777
	version.Transformer = aurora.InverterTransformerless
	version.Type = aurora.InputPhotovoltaic

	str = version.String()
	if str != "Model: Aurora 3.0-3.6 kW outdoor, Regulation: AS 4777, Transformer: Transformerless, Type: Photovoltaic" {
		t.Errorf("Unexpected string returned: %s", str)
	}
}

func TestAlarmStatesString(t *testing.T) {
	alarms := make(aurora.AlarmStates, 4, 4)
	str := alarms.String()

	if str != "No Alarm, No Alarm, No Alarm, No Alarm" {
		t.Errorf("Unexpected string returned: %s", str)
	}

	alarms[0] = aurora.AlarmState(99)
	alarms[1] = aurora.AlarmState(99)
	alarms[2] = aurora.AlarmState(99)
	alarms[3] = aurora.AlarmState(99)

	str = alarms.String()
	if str != "Unknown AlarmState(99), Unknown AlarmState(99), Unknown AlarmState(99), Unknown AlarmState(99)" {
		t.Errorf("Unexpected string returned: %s", str)
	}
}

func TestTransmissionStateString(t *testing.T) {
	if str := aurora.TransmissionState(0).String(); str != "Ok" {
		t.Errorf("Unexpected string returned: %s", str)
	}

	if str := aurora.TSEEpromNotAccessible.String(); str != "EEProm not accessible" {
		t.Errorf("Unexpected string returned: %s", str)
	}

	if str := aurora.TransmissionState(99).String(); str != "Unknown TransmissionState(99)" {
		t.Errorf("Unexpected string returned: %s", str)
	}
}

func TestDCDCStateString(t *testing.T) {
	if str := aurora.DCDCState(0).String(); str != "DcDc OFF" {
		t.Errorf("Unexpected string returned: %s", str)
	}

	if str := aurora.DCDCMPPT.String(); str != "MPPT" {
		t.Errorf("Unexpected string returned: %s", str)
	}

	if str := aurora.DCDCState(99).String(); str != "Unknown DCDCState(99)" {
		t.Errorf("Unexpected string returned: %s", str)
	}
}

func TestDSParameterString(t *testing.T) {
	if str := aurora.DSParameter(0).String(); str != "Unknown DSParameter(0)" {
		t.Errorf("Unexpected string returned: %s", str)
	}

	if str := aurora.DSPIleakInverter.String(); str != "Ileak (Inverter)" {
		t.Errorf("Unexpected string returned: %s", str)
	}

	if str := aurora.DSParameter(99).String(); str != "Unknown DSParameter(99)" {
		t.Errorf("Unexpected string returned: %s", str)
	}
}

func TestInverterStateString(t *testing.T) {
	if str := aurora.InverterState(0).String(); str != "Stand By" {
		t.Errorf("Unexpected string returned: %s", str)
	}

	if str := aurora.ISRun.String(); str != "Run" {
		t.Errorf("Unexpected string returned: %s", str)
	}

	if str := aurora.InverterState(99).String(); str != "Unknown InverterState(99)" {
		t.Errorf("Unexpected string returned: %s", str)
	}
}

func TestGlobalStateString(t *testing.T) {
	if str := aurora.GlobalState(0).String(); str != "Sending Parameters" {
		t.Errorf("Unexpected string returned: %s", str)
	}

	if str := aurora.GSRun.String(); str != "Run" {
		t.Errorf("Unexpected string returned: %s", str)
	}

	if str := aurora.GlobalState(250).String(); str != "Unknown GlobalState(250)" {
		t.Errorf("Unexpected string returned: %s", str)
	}
}


func TestConfigurationStateString(t *testing.T) {
	if str := aurora.ConfigurationState(0).String(); str != "System operating with both strings." {
		t.Errorf("Unexpected string returned: %s", str)
	}

	if str := aurora.ConfigString1.String(); str != "String 1 connected, String 2 disconnected." {
		t.Errorf("Unexpected string returned: %s", str)
	}

	if str := aurora.ConfigurationState(250).String(); str != "Unknown ConfigurationState(250)" {
		t.Errorf("Unexpected string returned: %s", str)
	}
}

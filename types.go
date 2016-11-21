// Copyright 2016 Shannon Wynter. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package aurora

import "fmt"

// Argument is an interface that exposes Byte() to return a single byte
// representation of the given argument
type Argument interface {
	Byte() byte
}

// State holds the structure for the inverter state, returned by the State() func
type State struct {
	Global   GlobalState
	Inverter InverterState
	Channel1 DCDCState
	Channel2 DCDCState
	Alarm    AlarmState
}

// String returns state as an easy to read string
func (s *State) String() string {
	return fmt.Sprintf("Global: %s, Inverter: %s, Channel1: %s, Channel2: %s, Alarm: %s", s.Global, s.Inverter, s.Channel1, s.Channel2, s.Alarm)
}

// Version holds the structure for the inverter version/model, returned byt he Version() func
type Version struct {
	Model       Product
	Regulation  ProductSpec
	Transformer InverterType
	Type        InputType
}

// String returns the version as an easy to read string
func (v *Version) String() string {
	return fmt.Sprintf("Model: %s, Regulation: %s, Transformer: %s, Type: %s", v.Model, v.Regulation, v.Transformer, v.Type)
}

// outputPayload is a convenient struct for writing binary to the inverter
type outputPayload struct {
	Payload [8]byte
	CRC     uint16
}

// inputPayload is a convenient struct for reading binary from the inverter
type inputPayload struct {
	Payload [6]byte
	CRC     uint16
}

// Byte is a concrete Argument
type Byte byte

// Command is a command to send to the inverter
type Command byte

// Counter is a counter parameter to request from the inverter with the GetCounter command
type Counter byte

// CumulationPeriod is a period parameters to request from the inverter
type CumulationPeriod byte

// DSParameter is a DSP parameter to request from the inverter
type DSParameter byte

// TransmissionState is a status flag returned by the inverter
type TransmissionState byte

// AlarmState is a state of alarm returned by the inverter as part of the State struct or as an
// array of 4 AlarmStates in response to the Last4Alarms request
type AlarmState byte

// ConfigurationState is a state of configuration returned by the inverter
type ConfigurationState byte

// DCDCState is a DCDC status flag from the inverter as part of the State struct
type DCDCState byte

// GlobalState is a global status flag returned by the inverter as part of the State sturct
type GlobalState byte

// Product is a product/model as returned by the inverter as part of the Version struct
type Product byte

// ProductSpec is a product specification/standard returned by the inverter as part of the Version struct
type ProductSpec byte

// InverterState is a status flag returned by the inverter as part of the State struct
type InverterState byte

// InverterType is a value returned as part of the Version struct
type InverterType byte

// InputType is a value returned as part of the Version struct
type InputType byte

// AlarmStates an array of AlarmState returned byt he Last4Alarms request
type AlarmStates []AlarmState

// Byte implement Argument.Byte()
func (b Byte) Byte() byte {
	return byte(b)
}

// Byte implement Argument.Byte()
func (c Counter) Byte() byte {
	return byte(c)
}

// Byte implement Argument.Byte()
func (c CumulationPeriod) Byte() byte {
	return byte(c)
}

// Byte implements Argument.Byte()
func (d DSParameter) Byte() byte {
	return byte(d)
}

func (o *outputPayload) String() string {
	return fmt.Sprintf("% X (%d)", o.Payload, o.CRC)
}

func (i *inputPayload) String() string {
	return fmt.Sprintf("% X (%d)", i.Payload, i.CRC)
}

func (t TransmissionState) String() string {
	if str, ok := transmissionStates[t]; ok {
		return str
	}

	return fmt.Sprintf("Unknown TransmissionState(%d)", byte(t))
}

func (a AlarmState) String() string {
	if str, ok := alarmStates[a]; ok {
		return str
	}

	return fmt.Sprintf("Unknown AlarmState(%d)", byte(a))
}

func (a AlarmStates) String() string {
	return fmt.Sprintf("%s, %s, %s, %s", a[0], a[1], a[2], a[3])
}

func (g GlobalState) String() string {
	if str, ok := globalStates[g]; ok {
		return str
	}

	return fmt.Sprintf("Unknown GlobalState(%d)", byte(g))
}

func (i InverterState) String() string {
	if str, ok := inverterStates[i]; ok {
		return str
	}

	return fmt.Sprintf("Unknown InverterState(%d)", byte(i))
}

func (d DCDCState) String() string {
	if str, ok := dcdcStates[d]; ok {
		return str
	}

	return fmt.Sprintf("Unknown DCDCState(%d)", byte(d))
}

func (d DSParameter) String() string {
	if str, ok := dsParameterStrings[d]; ok {
		return str
	}
	return fmt.Sprintf("Unknown DSParameter(%d)", byte(d))
}

func (p Product) String() string {
	if str, ok := productNames[p]; ok {
		return str
	}
	return fmt.Sprintf("Unknown Product (%d)", byte(p))
}

func (p ProductSpec) String() string {
	if str, ok := productSpecs[p]; ok {
		return str
	}
	return fmt.Sprintf("Unknown ProductSpec (%d)", byte(p))
}

func (i InverterType) String() string {
	if str, ok := inverterTypes[i]; ok {
		return str
	}
	return fmt.Sprintf("Unknown InverterType (%d)", byte(i))
}

func (m InputType) String() string {
	if str, ok := inputTypes[m]; ok {
		return str
	}
	return fmt.Sprintf("Unknown InputType (%d)", byte(m))
}

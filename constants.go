// Copyright 2016 Shannon Wynter. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package aurora

// InverterEpochOffset is the number of seconds since unix epoch that
// Power-One started the Aurora firmware epoch
const InverterEpochOffset = 946706400

// Command values
const (
	GetState Command = 50 + iota // Get the inverter state
	_
	GetPartNumber // Get the inverters part number
	_
	_
	_
	_
	_
	GetVersion // Get the hardware build version
	GetDSP     // Get a value from the DSP
	_
	_
	_
	GetSerialNumber // Get the inverters serial number
	_
	GetManufacturingDate // Get the year and month of manufacture
	_
	_
	_
	_
	GetTime            // Get the time from the inverter
	SetTime            // Set the time for the inverter
	GetFirmwareVersion // Get the inverters firmware version
	_
	_
	_
	GetLast10SecEnergy // Get the amount of energy exported in the past 10 seconds
	GetConfiguration   // Get the inverter configuration
	GetCumulatedEnergy // Get a value from the cumulated energy table
	_
	GetCounters // Get a counter
	_
	_
	_
	_
	_
	GetLast4Alarms // Get the last 4 alarms
)

// Available cumulation values
const (
	CumulatedDaily CumulationPeriod = iota
	CumulatedWeekly
	_
	CumulatedMonthly
	CumulatedYearly
	CumulatedTotal
	CumulatedPartial
)

// Available DSP values
const (
	DSPGridVoltage DSParameter = (1 + iota)
	DSPGridCurrent
	DSPGridPower
	DSPFrequency
	DSPVbulk
	DSPIleakDCDC
	DSPIleakInverter
	DSPPin1
	DSPPin2
	_ DSParameter = (11 + iota)
	DSPInverterTemperature
	DSPBoosterTemperature
	DSPInput1Voltage
	_
	DSPInput1Current
	DSPInput2Voltage
	DSPInput2Current
	DSPGridVoltageDCDC
	DSPGridFrequencyDCDC
	DSPIsolationResistance
	DSPVbulkDCDC
	DSPAverageGridVoltage
	DSPVbulkMid
	DSPPowerPeak
	DSPPowerPeakToday
	DSPGridVoltageNeutral
	DSPWindGeneratorFrequency
	DSPGridVoltageNeutralPhase
	DSPGridCurrentPhaseR
	DSPGridCurrentPhaseS
	DSPGridCurrentPhaseT
	DSPFrequencyPhaseR
	DSPFrequencyPhaseS
	DSPFrequencyPhaseT
	DSPVbulkPositive
	DSPVbulkNegative
	DSPSupervisorTemperature
	DSPAlimTemperature
	DSPHeatSinkTemperature
	DSPTemperature1
	DSPTemperature2
	DSPTemperature3
	DSPFan1Speed
	DSPFan2Speed
	DSPFan3Speed
	DSPFan4Speed
	DSPFan5Speed
	DSPPowerSaturationLimit
	DSPRiferimentoAnelloBulk
	DSPVpanelMicro
	DSPGridVoltagePhaseR
	DSPGridVoltagePhaseS
	DSPGridVoltagePhaseT
)

// Known products/models
const (
	Product2kWIndoor       Product = 'i'
	Product2kWOutdoor      Product = 'o'
	Product3_6kWIndoor     Product = 'I'
	Product3_6kWOutdoor    Product = 'O'
	Product5kWOutdoor      Product = '5'
	Product6kWOutdoor      Product = '6'
	Product3PhaseInterface Product = 'P'
	Product50kWModule      Product = 'C'
	Product4_2kWNew        Product = '4'
	Product3_6kWNew        Product = '3'
	Product3_3kWNew        Product = '2'
	Product3_0kWNew        Product = '1'
	Product12kW            Product = 'D'
	Product10kW            Product = 'X'
)

// Known product specifications/regulations
const (
	ProductSpecUL1741      ProductSpec = 'A'
	ProductSpecVDE0126     ProductSpec = 'E'
	ProductSpecDR1663_2000 ProductSpec = 'S'
	ProductSpecENELDK5950  ProductSpec = 'I'
	ProductSpecUKG83       ProductSpec = 'U'
	ProductSpecAS4777      ProductSpec = 'K'
	ProductSpecVDEFrench   ProductSpec = 'F'
)

// Transmission states
const (
	TSOk                    TransmissionState = 0
	TSCommandNotImplemented TransmissionState = (51 + iota)
	TSVariableDoesNotExist
	TSValueOutOfRange
	TSEEpromNotAccessible
	TSMicroError
	TSNotExecuted
	TSVariableNotAvailable
)

// Global states
const (
	GSSendingParameters GlobalState = iota
	GSWaitingSunGrid
	GSCheckingGrid
	GSMeasuringRiso
	GSDCDCStart
	GSInverterTurnOn
	GSRun
	GSRecovery
	GSPause
	GSGroundFault
	GSOTHFault
	GSAddressSetting
	GSSelfTest
	GSSelfTestFail
	GSSensorTestMeasureRiso
	GSLeakFault
	GSWaitingManualReset
	GSInternalErrorE026
	GSInternalErrorE027
	GSInternalErrorE028
	GSInternalErrorE029
	GSInternalErrorE030
	GSSendingWindTable
	GSFailedSendingTable
	GSUTHFault
	GSRemoteOff
	GSInterlockFail
	GSExecutingAutotest
	_
	_
	GSWaitingSun
	GSTemperatureFault
	GSFanStaucked
	GSIntComFail
	GSSlaveInsertion
	GSDCSwitchOpen
	GSTrasSwitchOpen
	GSMasterExclusion
	GSAutoExclusion
	_ GlobalState = iota + 58
	GSErasingInternalEEprom
	GSErasingExternalEEprom
	GSCountingEEprom
	GSFreeze
)

// Inverter states
const (
	ISStandBy InverterState = iota
	ISCheckingGrid
	ISRun
	ISBulkOverVoltage
	ISOutOverCurrent
	ISIGBTSat
	ISBulkUnderVoltage
	ISDegaussError
	ISNoParameters
	ISBulkLow
	ISGridOverVoltage
	ISCommunicationError
	ISDegaussing
	ISStarting
	ISBulkCapFail
	ISLeakFail
	ISDCDCFail
	ISIleakSensorFail
	ISSelfTestRelayInverter
	ISSelfTestWaitSensorTest
	ISSelfTestTestRelayDCDCSensor
	ISSelfTestRelayInverterFail
	ISSelfTestTimeoutFail
	ISSelfTestRelayDCDCFail
	ISSelfTest1
	ISWaitingSelfTestStart
	ISDCInjection
	ISSelfTest2
	ISSelfTest3
	ISSelfTest4
	ISInternalError30
	ISInternalError31
	_ InverterState = iota + 7
	ISForbiddenState
	ISInputUC
	ISZeroPower
	ISGridNotPresent
	ISWaitingStart
	ISMPPT
	ISGRIDFAIL
	ISINPUTOC
)

// DCDC states
const (
	DCDCOff DCDCState = iota
	DCDCRampStart
	DCDCMPPT
	_
	DCDCInputOverCurrent
	DCDCInputUnderVoltage
	DCDCInputOverVoltage
	DCDCInputLow
	DCDCNoParameters
	DCDCBulkOverVoltage
	DCDCCommunicationError
	DCDCRampFail
	DCDCInternalError
	DCDCInputModeError
	DCDCGroundFault
	DCDCInverterFail
	DCDCIGBTSat
	DCDCILEAKFail
	DCDCGridFail
	DCDCCommError
)

// Alarm states
const (
	AlarmNone AlarmState = iota
	AlarmSunLow1
	AlarmInputOverCurrent
	AlarmInputUnderVoltage
	AlarmInputOverVoltage
	AlarmSunLow5
	AlarmNoParameters
	AlarmBulkOverVoltage
	AlarmCommError
	AlarmOutputOverCurrent
	AlarmIGBTSat
	AlarmBulkUV11
	AlarmE009
	AlarmGridFail
	AlarmBulkLow
	AlarmRampFail
	AlarmDCDCFail16
	AlarmWrongMode
	AlarmGroundFault18
	AlarmOverTemp
	AlarmBulkCapFail
	AlarmInverterFail
	AlarmStartTimeout
	AlarmGroundFault23
	AlarmDegaussError
	AlarmIleakSensFail
	AlarmDCDCFail25
	AlarmSelfTestError1
	AlarmSelfTestError2
	AlarmSelfTestError3
	AlarmSelfTestError4
	AlarmDCInjError
	AlarmGridOverVoltage
	AlarmGridUnderVoltage
	AlarmGridOF
	AlarmGridUF
	AlarmZGridHi
	AlarmE024
	AlarmRisoLow
	ALarmVrefError
	AlarmErrorMeasV
	AlarmErrorMeasF
	AlarmErrorMeasI
	AlarmErrorMeasIleak
	AlarmReadErrorV
	AlarmReadErrorI
	AlarmTableFail
	AlarmFanFail
	AlarmUTH
	AlarmInterlockFail
	AlarmRemoteOff
	AlarmVoutAvgError
	AlarmBatteryLow
	AlarmClkFail
	AlarmInputUC
	AlarmZeroPower
	AlarmFanStucked
	AlarmDCSwitchOpen
	AlarmBulkUV58
	AlarmAutoexclusion
	AlarmGridDFDT
	AlarmDenSwitchOpen
	AlarmJboxFail
)

// Configuration states
const (
	ConfigBoth ConfigurationState = iota
	ConfigString1
	ConfigString2
)

// Counter values
const (
	CounterTotal Counter = iota
	CounterPartial
	CounterGrid
	CounterReset
)

// Inverter types
const (
	InverterTransformerless InverterType = 78
	InverterTransformer     InverterType = 84
)

// Inverter input types
const (
	InputPhotovoltaic InputType = 78
	InputWind         InputType = 87
)

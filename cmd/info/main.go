package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/freman/go-aurora"
	"github.com/jacobsa/go-serial/serial"
)

func errCheck(what string, err error) {
	if err != nil {
		log.Fatalf("inverter.%s: %v", what, err)
	}
}

func main() {
	fPort := flag.String("p", "/dev/ttyUSB0", "Serial port")
	flag.Parse()

	options := serial.OpenOptions{
		PortName:              *fPort,
		BaudRate:              19200,
		DataBits:              8,
		StopBits:              1,
		ParityMode:            serial.PARITY_NONE,
		InterCharacterTimeout: 250,
		MinimumReadSize:       4,
	}

	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	defer port.Close()

	inverter := &aurora.Inverter{
		Conn:    port,
		Address: 2,
	}

	errCheck("CommCheck", inverter.CommCheck())

	configuration, err := inverter.Configuration()
	errCheck("Configuration", err)

	version, err := inverter.Version()
	errCheck("Version", err)

	time, err := inverter.GetTime()
	errCheck("GetTime", err)

	inverterTemp, err := inverter.InverterTemperature()
	errCheck("InverterTemperature", err)

	boosterTemp, err := inverter.BoosterTemperature()
	errCheck("BoosterTemperature", err)

	year, week, err := inverter.ManufactureDate()
	errCheck("ManufactureDate", err)

	serialNumber, err := inverter.SerialNumber()
	errCheck("SerialNumber", err)

	fmt.Printf(`%v
Serial #: %s (Manufactured week %s of 20%s)
Inverter time: %v
Temperature %fºC / %fºC inverter/booster
String Configuration: %v
`,
		version,
		serialNumber,
		week,
		year,
		time,
		inverterTemp,
		boosterTemp,
		configuration,
	)

}

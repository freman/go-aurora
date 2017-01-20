package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/freman/go-aurora"

	"github.com/BurntSushi/toml"
	"github.com/matryer/try"
	"github.com/tarm/serial"
)

type result struct {
	Time                time.Time
	Address             byte
	BoosterTemperature  float32
	InverterTemperature float32
	Frequency           float32
	GridVoltage         float32
	GridCurrent         float32
	GridPower           float32
	GridRunTime         duration
	Input1Voltage       float32
	Input1Current       float32
	Input2Voltage       float32
	Input2Current       float32
	Joules              uint16
	DailyEnergy         uint32
	WeeklyEnergy        uint32
	MonthlyEnergy       uint32
	YearlyEnergy        uint32
	TotalEnergy         uint32
	TotalRunTime        duration
	SerialNumber        string
}

type results struct {
	sync.RWMutex
	Results map[byte]*result
}

type serialConfig struct {
	serial.Config
	ReadTimeout duration
}

type duration struct {
	time.Duration
}

func (o *serialConfig) Normalise() *serial.Config {
	o.Config.ReadTimeout = o.ReadTimeout.Duration
	return &o.Config
}

func (d *duration) UnmarshalText(text []byte) (err error) {
	d.Duration, err = time.ParseDuration(string(text))
	return
}

func (d duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(d.Duration.Seconds()))
}

func withDeadline(deadline time.Duration, f func() error) error {
	c := make(chan error, 1)
	defer close(c)
	go func() {
		c <- try.Do(func(attempt int) (bool, error) {
			time.Sleep(time.Duration(attempt) * time.Millisecond)
			err := f()
			return attempt < 3, err
		})
	}()
	select {
	case err := <-c:
		return err
	case <-time.After(deadline):
		return errors.New("Timeout while waiting for operation to complete")
	}
}

func main() {
	config := struct {
		UnitAddresses []byte
		UpdateRate    duration
		Deadline      duration
		Listen        string
		Comms         serialConfig
	}{
		UnitAddresses: []byte{2},
		UpdateRate: duration{
			Duration: time.Minute,
		},
		Deadline: duration{
			Duration: 5 * time.Second,
		},
		Listen: ":8080",
		Comms: serialConfig{
			Config: serial.Config{
				Name:     " /dev/ttyUSB0",
				Baud:     19200,
				Parity:   serial.ParityNone,
				StopBits: serial.Stop1,
			},
			ReadTimeout: duration{
				Duration: time.Duration(0),
			},
		},
	}

	fConfig := flag.String("config", "config.toml", "Path to the configuration file")
	flag.Parse()

	_, err := toml.DecodeFile(*fConfig, &config)
	if err != nil {
		log.Fatalf("Unable to parse configuration file due to %v.", err)
	}

	buffer := results{
		Results: map[byte]*result{},
	}

	port, err := serial.OpenPort(config.Comms.Normalise())
	if err != nil {
		log.Fatalf("Unable to open serial port due to %v.", err)
	}

	inverter := &aurora.Inverter{
		Conn: port,
	}

	for _, address := range config.UnitAddresses {
		inverter.Address = address
		buffer.Results[address] = &result{
			Address: address,
		}

		err := withDeadline(config.Deadline.Duration, func() (err error) {
			buffer.Results[address].SerialNumber, err = inverter.SerialNumber()
			return
		})

		if err != nil {
			log.Fatalf("Unable to communicate with inverter at address %d, error was %v", address, err)
		}
	}

	go func() {
		ticker := time.NewTicker(config.UpdateRate.Duration)
		now := time.Now()
		for {
			for _, address := range config.UnitAddresses {
				buffer.RLock()
				inverter.Address = address
				r := &result{
					Address:      address,
					SerialNumber: buffer.Results[address].SerialNumber,
					Time:         now,
				}
				buffer.RUnlock()

				err := withDeadline(config.Deadline.Duration, func() error {
					var err error
					if r.BoosterTemperature, err = inverter.BoosterTemperature(); err != nil {
						return err
					}
					if r.InverterTemperature, err = inverter.InverterTemperature(); err != nil {
						return err
					}
					if r.Frequency, err = inverter.Frequency(); err != nil {
						return err
					}
					if r.GridVoltage, err = inverter.GridVoltage(); err != nil {
						return err
					}
					if r.GridCurrent, err = inverter.GridCurrent(); err != nil {
						return err
					}
					if r.GridPower, err = inverter.GridPower(); err != nil {
						return err
					}
					if r.GridRunTime.Duration, err = inverter.GridRunTime(); err != nil {
						return err
					}
					if r.Input1Voltage, err = inverter.Input1Voltage(); err != nil {
						return err
					}
					if r.Input1Current, err = inverter.Input1Current(); err != nil {
						return err
					}
					if r.Input2Voltage, err = inverter.Input2Voltage(); err != nil {
						return err
					}
					if r.Input2Current, err = inverter.Input2Current(); err != nil {
						return err
					}
					if r.Joules, err = inverter.Joules(); err != nil {
						return err
					}
					if r.DailyEnergy, err = inverter.DailyEnergy(); err != nil {
						return err
					}
					if r.WeeklyEnergy, err = inverter.WeeklyEnergy(); err != nil {
						return err
					}
					if r.MonthlyEnergy, err = inverter.MonthlyEnergy(); err != nil {
						return err
					}
					if r.YearlyEnergy, err = inverter.YearlyEnergy(); err != nil {
						return err
					}
					if r.TotalEnergy, err = inverter.TotalEnergy(); err != nil {
						return err
					}
					r.TotalRunTime.Duration, err = inverter.TotalRunTime()
					return err
				})

				if err == nil {
					buffer.Lock()
					buffer.Results[address] = r
					buffer.Unlock()
				}
			}
			now = <-ticker.C
		}
	}()

	http.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		buffer.RLock()
		defer buffer.RUnlock()
		js, err := json.Marshal(buffer.Results)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})

	log.Fatal(http.ListenAndServe(config.Listen, nil))
}

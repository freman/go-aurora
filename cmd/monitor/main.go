package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/freman/go-aurora"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
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
	Results map[string]*result
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
		if err != nil {
			log.WithError(err).Errorf("Call to f() failed with error, %s", err.Error())
		}
		return err
	case <-time.After(deadline):
		log.WithField("deadline", deadline).Warning("Timeout while reading from inverter")
		return errors.New("Timeout while waiting for operation to complete")
	}
}

type configStruct struct {
	Name          string
	Comms         serialConfig
	UpdateRate    duration
	Deadline      duration
	UnitAddresses []byte
}

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	config := struct {
		LogPath    string
		UpdateRate duration
		Deadline   duration
		Listen     string
		Devices    []configStruct
	}{
		LogPath: filepath.Join(dir, "main.log"),
		UpdateRate: duration{
			Duration: time.Minute,
		},
		Deadline: duration{
			Duration: 5 * time.Second,
		},
		Listen: ":8080",
	}

	fConfig := flag.String("config", "config.toml", "Path to the configuration file")
	flag.Parse()

	_, err = toml.DecodeFile(*fConfig, &config)
	if err != nil {
		log.Fatalf("Unable to parse configuration file due to %v.", err)
	}

	f, err := os.OpenFile(config.LogPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.Infof("Logging to %s, there should be no further output", config.LogPath)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(f)
	log.SetLevel(log.DebugLevel)

	log.WithField("config", config).Debug("Configuration")

	buffer := results{
		Results: map[string]*result{},
	}

	for _, device := range config.Devices {
		go func(device configStruct) {
			logger := log.WithField("coms", device.Comms.Name)
			deadline := device.Deadline.Duration
			if deadline == 0 {
				deadline = config.Deadline.Duration
			}

			updateRate := device.UpdateRate.Duration
			if updateRate == 0 {
				updateRate = config.UpdateRate.Duration
			}

			port, err := serial.OpenPort(device.Comms.Normalise())
			if err != nil {
				log.WithError(err).Fatal("Startup error: Unable to open serial port")
			}

			inverter := &aurora.Inverter{
				Conn: port,
			}

			for _, address := range device.UnitAddresses {
				logger := logger.WithField("address", address)
				name := fmt.Sprintf("%s::%d", device.Comms.Name, address)
				inverter.Address = address
				buffer.Results[name] = &result{
					Address: address,
				}

				err := withDeadline(deadline, func() (err error) {
					buffer.Results[name].SerialNumber, err = inverter.SerialNumber()
					return
				})

				if err != nil {
					logger.WithError(err).Fatal("Startup error: Unable to communicate with inverter")
				}
			}

			ticker := time.NewTicker(updateRate)
			now := time.Now()
			for {
				for _, address := range device.UnitAddresses {
					name := fmt.Sprintf("%s::%d", device.Comms.Name, address)
					buffer.RLock()
					logger := logger.WithFields(log.Fields{
						"address": address,
						"serial":  buffer.Results[name].SerialNumber,
					})
					inverter.Address = address
					r := &result{
						Address:      address,
						SerialNumber: buffer.Results[name].SerialNumber,
						Time:         now,
					}
					buffer.RUnlock()

					err := withDeadline(deadline, func() error {
						var err error
						if r.BoosterTemperature, err = inverter.BoosterTemperature(); err != nil {
							logger.WithError(err).Warning("Unable to read BoosterTemperature")
							return err
						}
						if r.InverterTemperature, err = inverter.InverterTemperature(); err != nil {
							logger.WithError(err).Warning("Unable to read InverterTemperature")
							return err
						}
						if r.Frequency, err = inverter.Frequency(); err != nil {
							logger.WithError(err).Warning("Unable to read Frequency")
							return err
						}
						if r.GridVoltage, err = inverter.GridVoltage(); err != nil {
							logger.WithError(err).Warning("Unable to read GridVoltage")
							return err
						}
						if r.GridCurrent, err = inverter.GridCurrent(); err != nil {
							logger.WithError(err).Warning("Unable to read GridCurrent")
							return err
						}
						if r.GridPower, err = inverter.GridPower(); err != nil {
							logger.WithError(err).Warning("Unable to read GridPower")
							return err
						}
						if r.GridRunTime.Duration, err = inverter.GridRunTime(); err != nil {
							logger.WithError(err).Warning("Unable to read GridRunTime")
							return err
						}
						if r.Input1Voltage, err = inverter.Input1Voltage(); err != nil {
							logger.WithError(err).Warning("Unable to read Input1Voltage")
							return err
						}
						if r.Input1Current, err = inverter.Input1Current(); err != nil {
							logger.WithError(err).Warning("Unable to read Input1Current")
							return err
						}
						if r.Input2Voltage, err = inverter.Input2Voltage(); err != nil {
							logger.WithError(err).Warning("Unable to read Input2Voltage")
							return err
						}
						if r.Input2Current, err = inverter.Input2Current(); err != nil {
							logger.WithError(err).Warning("Unable to read Input2Current")
							return err
						}
						if r.Joules, err = inverter.Joules(); err != nil {
							logger.WithError(err).Warning("Unable to read Joules")
							return err
						}
						if r.DailyEnergy, err = inverter.DailyEnergy(); err != nil {
							logger.WithError(err).Warning("Unable to read DailyEnergy")
							return err
						}
						if r.WeeklyEnergy, err = inverter.WeeklyEnergy(); err != nil {
							logger.WithError(err).Warning("Unable to read WeeklyEnergy")
							return err
						}
						if r.MonthlyEnergy, err = inverter.MonthlyEnergy(); err != nil {
							logger.WithError(err).Warning("Unable to read MonthlyEnergy")
							return err
						}
						if r.YearlyEnergy, err = inverter.YearlyEnergy(); err != nil {
							logger.WithError(err).Warning("Unable to read YearlyEnergy")
							return err
						}
						if r.TotalEnergy, err = inverter.TotalEnergy(); err != nil {
							logger.WithError(err).Warning("Unable to read TotalEnergy")
							return err
						}
						r.TotalRunTime.Duration, err = inverter.TotalRunTime()
						if err != nil {
							logger.WithError(err).Warning("Unable to read TotalRunTime")
						}
						return err
					})

					if err == nil {
						buffer.Lock()
						buffer.Results[name] = r
						buffer.Unlock()
					}
				}
				now = <-ticker.C
			}
		}(device)
	}

	http.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		logger := log.WithField("remoteaddr", r.RemoteAddr)
		logger.Info("GET /json")
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

package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/freman/go-aurora"
	"github.com/tarm/serial"
)

const errorStringFormat = "String %d didn't reach %f, the highest recorded output current was %f"

func errCheck(what string, err error) {
	if err != nil {
	}
}

func sleep(hour int, tomorrow bool) {
	log.Printf("Sleeping until %d:00\n", hour)
	now := time.Now()
	then := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	then = then.Add(1 * time.Second)
	if tomorrow {
		then = then.Add(24 * time.Hour)
	}
	time.Sleep(then.Sub(now))
}

func sendMail(username, password, server, from, recipient string, lines []string) {
	rubbish := strings.Split(server, "#")
	var auth smtp.Auth
	if len(rubbish) > 1 {
		auth = smtp.PlainAuth("", username, password, rubbish[1])
		server = rubbish[0]
	} else {
		rubbish = strings.Split(server, ":")
		auth = smtp.PlainAuth("", username, password, rubbish[0])
	}
	to := []string{recipient}
	msgLines := append([]string{
		"To: " + recipient,
		"Subject: Aurora Alert",
		""}, lines...)
	msg := []byte(strings.Join(msgLines, "\r\n"))
	err := smtp.SendMail(server, auth, from, to, msg)
	if err != nil {
		log.Printf("Can't send email %v", err)
	}
}

func main() {
	fPort := flag.String("p", "/dev/ttyUSB0", "Serial port")
	fString1 := flag.Float64("1", 4, "Minimum current threshold string 1")
	fString2 := flag.Float64("2", 4, "Minimum current threshold string 2")
	fCheckStart := flag.Int("s", 9, "Start checking at this hour")
	fCheckEnd := flag.Int("e", 17, "Stop checking at this hour")

	fServer := flag.String("m", "localhost:25", "SMTP server")
	fUsername := flag.String("u", "", "SMTP auth username")
	fPassword := flag.String("P", "", "SMTP auth password (env PASSWORD)")
	fRecipient := flag.String("r", "help@example.com", "Email recipient")
	fSender := flag.String("S", *fRecipient, "Email sender")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	flag.Parse()

	if *fPassword == "" {
		*fPassword = os.Getenv("PASSWORD")
	}

	sendMail(*fUsername, *fPassword, *fServer, *fSender, *fRecipient, []string{"Starting up..."})

	options := &serial.Config{
		Name:   *fPort,
		Baud:   19200,
		Parity: serial.ParityNone,
	}

	string1Max := float64(0)
	string2Max := float64(0)
	haveChecked := false

	for {
		if tomorrow := time.Now().Hour() > *fCheckEnd; tomorrow || time.Now().Hour() < *fCheckStart {
			if haveChecked {
				emailLines := []string{}
				if string1Max < *fString1 {
					emailLines = append(emailLines, fmt.Sprintf(errorStringFormat, 1, *fString1, string1Max))
				}
				if string2Max < *fString2 {
					emailLines = append(emailLines, fmt.Sprintf(errorStringFormat, 2, *fString2, string2Max))
				}
				if len(emailLines) > 0 {
					sendMail(*fUsername, *fPassword, *fServer, *fSender, *fRecipient, emailLines)
				}
			}
			haveChecked = false
			string1Max = 0
			string2Max = 0
			sleep(*fCheckStart, tomorrow)
			continue
		}

		func() {
			port, err := serial.OpenPort(options)
			if err != nil {
				log.Printf("serial.Open: %v", err)
				return
			}

			defer port.Close()

			inverter := &aurora.Inverter{
				Conn: port,

				Address: 2,
			}

			if err := inverter.CommCheck(); err != nil {
				log.Printf("inverter.CommCheck: %v", err)
				return
			}

			c1, err := inverter.Input1Current()
			if err != nil {
				log.Printf("inverter.Input1Current: %v", err)
				return
			}

			c2, err := inverter.Input2Current()
			if err != nil {
				log.Printf("inverter.Input2Current: %v", err)
				return
			}

			haveChecked = true
			string1Max = math.Max(float64(c1), string1Max)
			string2Max = math.Max(float64(c2), string2Max)
		}()
		time.Sleep(time.Minute)
	}
}

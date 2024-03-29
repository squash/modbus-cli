package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/goburrow/modbus"
	"github.com/gofrs/flock"
)

type config struct {
	Port       string
	Device     uint
	Baud       int
	Retries    uint
	Address    string
	Count      uint
	OutputAs   string
	Write      bool
	WriteValue string
}

// getUint16FromString will parse a hex or decimal value to a uint16, also detect '=' for 8 bit masking
func getUint16FromString(in string) (uint16, bool) {
	var a uint16
	iseight := false
	// Add support for =8 suffix on addresses
	tmp := strings.Split(in, "=")
	in = tmp[0]
	if len(tmp) == 2 {
		iseight = true
	}
	if strings.HasPrefix(in, "0x") {
		tmp := in[2:]
		b, err := strconv.ParseUint(tmp, 16, 16)
		if err != nil {
			log.Fatal(err)
		}
		a = uint16(b)
	} else {
		b, err := strconv.ParseUint(in, 10, 16)
		if err != nil {
			log.Fatal(err)
		}
		a = uint16(b)
	}
	return a, iseight
}

type result struct {
	Address uint16
	Values  []uint16
}

// readRegister actually pulls data via the serial modbus connection and outputs as it goes unless we're using json
func readRegister(client modbus.Client, a string, outputas string, count uint) (result, error) {
	var r result
	iseight := false
	r.Address, iseight = getUint16FromString(a)
	v, err := client.ReadHoldingRegisters(r.Address, uint16(count))
	if err != nil {
		if err.Error() == "serial: timeout" {
			return r, errors.New("Timeout")
		}
		return r, err
	}
	for x := uint(0); x < count; x++ {
		tmp := binary.BigEndian.Uint16(v[x*2:])
		if iseight {
			tmp = tmp & 255
		}
		r.Values = append(r.Values, tmp)
		switch outputas {
		case "go":
			fmt.Printf("%#v ", v)
		case "hex":
			fmt.Printf("0x%x ", tmp)
		case "decimal":
			fmt.Printf("%d ", tmp)
		}
	}
	return r, nil
}

func main() {
	var c config
	flag.StringVar(&c.Port, "port", "/dev/ttyUSB0", "Serial device (example: /dev/ttyUSB0")
	flag.UintVar(&c.Device, "device", 1, "Modbus device id to query")
	flag.IntVar(&c.Baud, "baud", 9600, "Baud rate")
	flag.StringVar(&c.Address, "address", "0x100", "Data address to query (hex prefixed with 0x, or decimal). Separate multiple addresses with a comma to read multiple registers. Add =8 to an address to only return the lower 8 bits.")
	flag.UintVar(&c.Count, "count", 1, "Number of 16 bit registers to read")
	flag.UintVar(&c.Retries, "retries", 1, "Retry query this many times")
	flag.StringVar(&c.OutputAs, "output-as", "go", "Format to print output values. Options: hex, decimal, go, json")
	flag.BoolVar(&c.Write, "write", false, "Write [value] to [address], then read it back")
	flag.StringVar(&c.WriteValue, "value", "", "Value to be written (hex prefixed with 0x, or decimal)")
	flag.Parse()

	if c.OutputAs != "decimal" && c.OutputAs != "hex" && c.OutputAs != "go" && c.OutputAs != "json" {
		log.Fatal("Output format invalid")
	}
	locker := flock.New(c.Port)
	lock := false
	var err error
	for !lock {
		lock, err = locker.TryLock()
		if err != nil {
			log.Println(err.Error())
			time.Sleep(10 * time.Millisecond)
		}
	}
	defer locker.Unlock()
	handler := modbus.NewRTUClientHandler(c.Port)
	handler.BaudRate = c.Baud
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = byte(c.Device)
	handler.Timeout = 5 * time.Second

	err = handler.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer handler.Close()
	client := modbus.NewClient(handler)
	success := false
	checks := strings.Split(c.Address, ",")
	if c.Write {
		if len(checks) != 1 {
			log.Fatal("Write can only target a single register.")
		}
		a, _ := getUint16FromString(checks[0])
		v, _ := getUint16FromString(c.WriteValue)

		for x := 0; x < int(c.Retries) && !success; x++ {
			client := modbus.NewClient(handler)
			results, err := client.WriteSingleRegister(a, v)
			if err != nil {
				log.Println(err)
			} else {
				success = true
			}
			time.Sleep(10 * time.Millisecond) // modbus serial seems to need a rest between queries
			log.Printf("%#v", results)
		}
	}
	var results []result
	for _, address := range checks {
		success = false
		for x := uint(0); x < c.Retries && !success; x++ {
			r, err := readRegister(client, address, c.OutputAs, c.Count)

			if err != nil {
				if err.Error() == "serial: timeout" {
					continue
				}
				log.Println(err)
			} else {
				success = true
				time.Sleep(10 * time.Millisecond) // modbus serial seems to need a rest between queries
				results = append(results, r)
			}

		}
	}

	if c.OutputAs == "json" {
		j, err := json.Marshal(&results)
		if err == nil {
			fmt.Println(string(j))
		} else {
			log.Println(err)
		}

	}
}

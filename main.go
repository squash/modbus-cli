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
)

type RoverData struct {
	PVVolts     float32
	PVAmps      float32
	ChargeVolts float32
	ChargeAmps  float32
	ChargeMode  string
	Timestamp   int64
}

type config struct {
	Port     string
	Device   uint
	Baud     int
	Retries  uint
	Address  string
	Count    uint
	OutputAs string
	Write bool
	WriteValue string
}

func getUint16FromString(in string) uint16 {
		var a uint16
		if strings.HasPrefix(in, "0x") {
			tmp := in[2:]
			b, err := strconv.ParseUint(tmp, 16, 16)
			if err != nil {
				log.Fatal(err)
			}
			a=uint16(b)
		} else {
			b, err := strconv.ParseUint(in, 10, 16)
			if err != nil {
				log.Fatal(err)
			}
			a=uint16(b)
		}
		return a
}
type result struct {
	Address uint16
	Values []uint16
}

func readRegister(client modbus.Client, a string, outputas string, count uint) (result, error) {
	var r result
		r.Address=getUint16FromString(a)
		v, err := client.ReadHoldingRegisters(r.Address, uint16(count))
		if err != nil {
			if err.Error()=="serial: timeout" { 
				return r, errors.New("Timeout")
			}
			return r, err
		}
		for x:=uint(0);x<count;x++ {
			tmp := binary.BigEndian.Uint16(v[x*2:])
			r.Values=append(r.Values, tmp)
			switch outputas {
				case "go":
				fmt.Printf("%#v ", v)
				case "hex":
				fmt.Printf("0x%x ", tmp)
				case "decimal":
				fmt.Printf("%d ", tmp)
			}
		}
		return r,nil
}

func main() {
	var c config
	flag.StringVar(&c.Port, "port", "/dev/ttyUSB0", "Serial device (example: /dev/ttyUSB0")
	flag.UintVar(&c.Device, "device", 1, "Modbus device id to query")
	flag.IntVar(&c.Baud, "baud", 9600, "Baud rate")
	flag.StringVar(&c.Address, "address", "0x100", "Data address to query (hex prefixed with 0x, or decimal). Separate multiple addresses with a comma to read multiple registers.")
	flag.UintVar(&c.Count, "count", 1, "Number of 16 bit registers to read")
	flag.UintVar(&c.Retries, "retries", 1, "Retry query this many times")
	flag.StringVar(&c.OutputAs, "output-as", "go", "Format to print output values. Options: hex, decimal, go")
	flag.BoolVar(&c.Write, "write", false, "Write [value] to [address], then read it back" )
	flag.StringVar(&c.WriteValue, "value", "", "Value to be written (hex prefixed with 0x, or decimal)")
	flag.Parse()

	if c.OutputAs != "decimal" && c.OutputAs != "hex" && c.OutputAs != "go" {
		log.Fatal("Output format invalid")
	}
		handler := modbus.NewRTUClientHandler(c.Port)
		handler.BaudRate = c.Baud
		handler.DataBits = 8
		handler.Parity = "N"
		handler.StopBits = 1
		handler.SlaveId = byte(c.Device)
		handler.Timeout = 5 * time.Second

		err := handler.Connect()
		if err != nil {
			log.Fatal(err)
		}
		defer handler.Close()

	checks:=strings.Split(",", c.Address)
		client := modbus.NewClient(handler)
		if c.Write {
			if len(checks)!=1 {
				log.Fatal("Write can only target a sigle register")
			}
			a:=getUint16FromString(checks[0])
			v:=getUint16FromString(c.WriteValue)
			results,err:=client.WriteSingleRegister(a, v)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%#v", results)
		}
		var results []result
	for _,address:=range(checks) {
		for x := uint(0); x < c.Retries; x++ {
			r, err:=readRegister(client, address, c.OutputAs, c.Count)
			if err != nil {
			if err.Error()=="serial: timeout" { 
				continue
			}
			log.Println(err)
			results=append(results, r)
		}
	}
}

		if c.OutputAs=="json" {
			j,err:=json.Marshal(&results)
			if err != nil {
				fmt.Println(j)
			} else {
				log.Println(err)
			}

		}
	}
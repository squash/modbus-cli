package main

import (
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
	Port string
	Device int
	Baud int
	Retries int
	Address string
	Bytes uint
}


func main() {
	var c config
	flag.StringVar(&c.Port, "port", "/dev/ttyUSB0", "Serial device (example: /dev/ttyUSB0")
	flag.IntVar(&c.Device, "device", 1, "Modbus device id to query")
	flag.IntVar(&c.Baud, "baud", 9600, "Baud rate")
	flag.StringVar(&c.Address, "address", "0x100", "Data address to query (hex prefixed with 0x, or decimal)")
	flag.UintVar(&c.Bytes, "bytes", 1, "Number of 8 bit bytes for result")
	flag.IntVar(&c.Retries, "retries", 1, "Retry query this many times")
	flag.Parse()

	handler := modbus.NewRTUClientHandler(c.Port)
	handler.BaudRate = c.Baud
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 5 * time.Second

	err := handler.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer handler.Close()
	client := modbus.NewClient(handler)
	var a uint64
	if strings.HasPrefix(c.Address, "0x") {
		tmp:=c.Address[2:]
		a,err=strconv.ParseUint(tmp, 16, 64)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		a,err=strconv.ParseUint(c.Address, 16, 64)
		if err != nil {
			log.Fatal(err)
		}
	}
	results,err:=client.ReadHoldingRegisters(uint16(a), uint16(c.Bytes))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v", results)
}

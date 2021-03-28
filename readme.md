# go-pa1010d

A Go library to read GNSS (e.g. GPS) sensor data from a "CD-PA1010D GNSS patch antenna module".

This was written against the I2C interface of a Raspberry Pi using the `github.com/d2r2/go-i2c` library although it is not a dependency - you just have to provide a `ReadBytes` method.

### Example usage

```go
package main

import (
	"log"

	"github.com/g-wilson/go-pa1010d"

	"github.com/d2r2/go-i2c"
	"github.com/kr/pretty"
)

const sensorAddress = 0x10

func main() {
	bus, err := i2c.NewI2C(sensorAddress, 1)
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	gnssReader := pa1010d.New(bus)

	results, errors := gnssReader.Listen()
	for {
		select {
		case err := <-errors:
			log.Println(err)
		case sentence := <-results:
			pretty.Println(sentence)
		}
	}
}

```

### Known issues

The `$PMTK011,MTKGPS*08` bootup message cannot be parsed by the go-nmea PMTK parser, because it expects integer commands - MTKGPS is a string. You will see an error on the errors channel, but it will happily continue.

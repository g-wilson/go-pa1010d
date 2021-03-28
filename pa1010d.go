package pa1010d

import (
	"fmt"

	"github.com/adrianmo/go-nmea"
)

type Bus interface {
	ReadBytes(buf []byte) (int, error)
}

type PA1010DReader struct {
	bus          Bus
	nmeaChannel  chan nmea.Sentence
	errorChannel chan error
}

func New(b Bus) *PA1010DReader {
	return &PA1010DReader{
		bus:          b,
		nmeaChannel:  make(chan nmea.Sentence, 1),
		errorChannel: make(chan error, 1),
	}
}

// Listen continuously polls the bus for new nmea messages
// and broadcasts them to the returned channel(s)
func (r *PA1010DReader) Listen() (<-chan nmea.Sentence, <-chan error) {
	go (func() {
		for {
			line, err := r.readMessage()
			if err != nil {
				r.errorChannel <- err
				continue
			}
			if len(line) == 0 {
				continue
			}

			sentence, err := nmea.Parse(string(line))
			if err != nil {
				r.errorChannel <- fmt.Errorf("line: %s nmea: %w", string(line), err)
				continue
			}

			r.nmeaChannel <- sentence
		}
	})()

	return r.nmeaChannel, r.errorChannel
}

// readMessage continuously collects bytes from the i2c device
// until either an error occurs, or a potential NMEA sentence is found
// https://en.wikipedia.org/wiki/NMEA_0183#Message_structure
func (r *PA1010DReader) readMessage() ([]byte, error) {
	s := []byte{}

	// keep reading until we reach the $ NMEA start-of-message
	for {
		var buf = make([]byte, 1)
		numread, err := r.bus.ReadBytes(buf)
		if err != nil {
			return []byte{}, nil
		}
		if numread == 0 {
			continue
		}
		if buf[0] == 0x24 { // $ character
			s = append(s, buf[0])
			break
		}
	}

	// keep reading and store each byte in the line until the end
	for {
		var buf = make([]byte, 1)
		numread, err := r.bus.ReadBytes(buf)
		if err != nil {
			return []byte{}, nil
		}
		if numread == 0 {
			continue
		}
		if buf[0] == 0x0d { // \r character
			break
		}
		// for some reason it puts tons of pointless "\n" mid-message
		if buf[0] != 0x0a {
			s = append(s, buf[0])
		}
	}

	return s, nil
}

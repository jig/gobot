package firmata

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/hybridgroup/gobot"
	"github.com/tarm/goserial"
)

var _ gobot.AdaptorInterface = (*FirmataAdaptor)(nil)

type FirmataAdaptor struct {
	gobot.Adaptor
	board      *board
	i2cAddress byte
	connect    func(*FirmataAdaptor) (err error)
}

// NewFirmataAdaptor returns a new firmata adaptor with specified name and optionally accepts:
//
//	string: port the FirmataAdaptor uses to connect to a serial port with a baude rate of 57600
//	io.ReadWriteCloser: connection the FirmataAdaptor uses to communication with the hardware
//
// If an io.ReadWriteCloser is not supplied, the FirmataAdaptor will open a connection
// to a serial port with a baude rate of 57600. If an io.ReadWriteCloser
// is supplied, then the FirmataAdaptor will use the provided io.ReadWriteCloser and use the
// string port as a label to be displayed in the log and api.
func NewFirmataAdaptor(name string, args ...interface{}) *FirmataAdaptor {
	var conn io.ReadWriteCloser
	var port string = ""

	for _, arg := range args {
		switch arg.(type) {
		case string:
			port = arg.(string)
		case io.ReadWriteCloser:
			conn = arg.(io.ReadWriteCloser)
		}
	}

	return &FirmataAdaptor{
		Adaptor: *gobot.NewAdaptor(
			name,
			"FirmataAdaptor",
			port,
		),
		connect: func(f *FirmataAdaptor) (err error) {
			if conn == nil {
				conn, err = serial.OpenPort(&serial.Config{Name: f.Port(), Baud: 57600})
				if err != nil {
					return err
				}
			}
			f.board = newBoard(conn)
			return
		},
	}
}

// Connect returns true if connection to board is succesfull
func (f *FirmataAdaptor) Connect() (errs []error) {
	if err := f.connect(f); err != nil {
		return []error{err}
	}
	f.SetConnected(true)
	return
}

// close finishes connection to serial port
// Prints error message on error
func (f *FirmataAdaptor) Disconnect() (err error) {
	if f.board != nil {
		return f.board.serial.Close()
	}
	return
}

// Finalize disconnects firmata adaptor
func (f *FirmataAdaptor) Finalize() (errs []error) {
	if err := f.Disconnect(); err != nil {
		return []error{err}
	}
	return
}

// InitServo (not yet implemented)
func (f *FirmataAdaptor) InitServo() (err error) {
	return errors.New("InitServo is not yet implemented")
}

// ServoWrite sets angle form 0 to 360 to specified servo pin
func (f *FirmataAdaptor) ServoWrite(pin string, angle byte) (err error) {
	p, err := strconv.Atoi(pin)
	if err != nil {
		return err
	}

	err = f.board.setPinMode(byte(p), servo)
	if err != nil {
		return err
	}
	err = f.board.analogWrite(byte(p), angle)
	return
}

// PwmWrite writes analog value to specified pin
func (f *FirmataAdaptor) PwmWrite(pin string, level byte) (err error) {
	p, err := strconv.Atoi(pin)
	if err != nil {
		return err
	}

	err = f.board.setPinMode(byte(p), pwm)
	if err != nil {
		return err
	}
	err = f.board.analogWrite(byte(p), level)
	return
}

// DigitalWrite writes digital values to specified pin
func (f *FirmataAdaptor) DigitalWrite(pin string, level byte) (err error) {
	p, err := strconv.Atoi(pin)
	if err != nil {
		return
	}

	err = f.board.setPinMode(byte(p), output)
	if err != nil {
		return
	}

	err = f.board.digitalWrite(byte(p), level)
	return
}

// DigitalRead retrieves digital value from specified pin
// Returns -1 if response from board is timed out
func (f *FirmataAdaptor) DigitalRead(pin string) (val int, err error) {
	ret := make(chan int)

	p, err := strconv.Atoi(pin)
	if err != nil {
		return
	}
	if err = f.board.setPinMode(byte(p), input); err != nil {
		return
	}
	if err = f.board.togglePinReporting(byte(p), high, reportDigital); err != nil {
		return
	}
	if err = f.board.readAndProcess(); err != nil {
		return
	}

	gobot.Once(f.board.events[fmt.Sprintf("digital_read_%v", pin)], func(data interface{}) {
		ret <- int(data.([]byte)[0])
	})

	select {
	case data := <-ret:
		return data, nil
	case <-time.After(10 * time.Millisecond):
	}
	return -1, nil
}

// AnalogRead retrieves value from analog pin.
// NOTE pins are numbered A0-A5, which translate to digital pins 14-19
func (f *FirmataAdaptor) AnalogRead(pin string) (val int, err error) {
	ret := make(chan int)

	p, err := strconv.Atoi(pin)
	if err != nil {
		return
	}
	p = f.digitalPin(p)
	if err = f.board.setPinMode(byte(p), analog); err != nil {
		return
	}

	if err = f.board.togglePinReporting(byte(p), high, reportAnalog); err != nil {
		return
	}

	if err = f.board.readAndProcess(); err != nil {
		return
	}

	gobot.Once(f.board.events[fmt.Sprintf("analog_read_%v", pin)], func(data interface{}) {
		b := data.([]byte)
		ret <- int(uint(b[0])<<24 | uint(b[1])<<16 | uint(b[2])<<8 | uint(b[3]))
	})

	select {
	case data := <-ret:
		return data, nil
	case <-time.After(10 * time.Millisecond):
	}
	return -1, nil
}

// AnalogWrite writes value to ananlog pin
func (f *FirmataAdaptor) AnalogWrite(pin string, level byte) (err error) {
	return f.PwmWrite(pin, level)
}

// digitalPin converts pin number to digital mapping
func (f *FirmataAdaptor) digitalPin(pin int) int {
	return pin + 14
}

// I2cStart initializes board with i2c configuration
func (f *FirmataAdaptor) I2cStart(address byte) (err error) {
	f.i2cAddress = address
	return f.board.i2cConfig([]byte{0})
}

// I2cRead reads from I2c specified size
// Returns empty byte array if response is timed out
func (f *FirmataAdaptor) I2cRead(size uint) (data []byte, err error) {
	ret := make(chan []byte)
	if err = f.board.i2cReadRequest(f.i2cAddress, size); err != nil {
		return
	}

	if err = f.board.readAndProcess(); err != nil {
		return
	}

	gobot.Once(f.board.events["i2c_reply"], func(data interface{}) {
		ret <- data.(map[string][]byte)["data"]
	})

	select {
	case data := <-ret:
		return data, nil
	case <-time.After(10 * time.Millisecond):
	}
	return []byte{}, nil
}

// I2cWrite retrieves i2c data
func (f *FirmataAdaptor) I2cWrite(data []byte) (err error) {
	return f.board.i2cWriteRequest(f.i2cAddress, data)
}

package sphero

import (
	"io"

	"github.com/hybridgroup/gobot"
	"github.com/tarm/goserial"
)

var _ gobot.AdaptorInterface = (*SpheroAdaptor)(nil)

// Represents a Connection to a Sphero
type SpheroAdaptor struct {
	gobot.Adaptor
	sp      io.ReadWriteCloser
	connect func(*SpheroAdaptor) (err error)
}

// NewSpheroAdaptor returns a new SpheroAdaptor given a name and port
func NewSpheroAdaptor(name string, port string) *SpheroAdaptor {
	return &SpheroAdaptor{
		Adaptor: *gobot.NewAdaptor(
			name,
			"SpheroAdaptor",
			port,
		),
		connect: func(a *SpheroAdaptor) (err error) {
			c := &serial.Config{Name: a.Port(), Baud: 115200}
			s, err := serial.OpenPort(c)
			if err != nil {
				return err
			}
			a.sp = s
			return
		},
	}
}

// Connect initiates a connection to the Sphero. Returns true on successful connection.
func (a *SpheroAdaptor) Connect() (errs []error) {
	if err := a.connect(a); err != nil {
		return []error{err}
	}
	a.SetConnected(true)
	return
}

// Reconnect attempts to reconnect to the Sphero. If the Sphero has an active connection
// it will first close that connection and then establish a new connection.
// Returns true on Successful reconnection
func (a *SpheroAdaptor) Reconnect() (errs []error) {
	if a.Connected() == true {
		a.Disconnect()
	}
	return a.Connect()
}

// Disconnect terminates the connection to the Sphero. Returns true on successful disconnect.
func (a *SpheroAdaptor) Disconnect() (errs []error) {
	if err := a.sp.Close(); err != nil {
		return []error{err}
	}
	a.SetConnected(false)
	return
}

// Finalize finalizes the SpheroAdaptor
func (a *SpheroAdaptor) Finalize() (errs []error) {
	return
}

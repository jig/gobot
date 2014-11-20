package i2c

import (
	"github.com/hybridgroup/gobot"
	"testing"
)

// --------- HELPERS
func initTestBlinkMDriver() (driver *BlinkMDriver) {
	driver, _ = initTestBlinkDriverWithStubbedAdaptor()
	return
}

func initTestBlinkDriverWithStubbedAdaptor() (*BlinkMDriver, *i2cTestAdaptor) {
	adaptor := newI2cTestAdaptor("adaptor")
	return NewBlinkMDriver(adaptor, "bot"), adaptor
}

// --------- TESTS

func TestBlinkMDriver(t *testing.T) {
	// Does it implement gobot.DriverInterface?
	var _ gobot.DriverInterface = (*BlinkMDriver)(nil)

	// Does its adaptor implements the I2cInterface?
	driver := initTestBlinkMDriver()
	var _ I2cInterface = driver.adaptor()
}

func TestNewBlinkMDriver(t *testing.T) {
	// Does it return a pointer to an instance of BlinkMDriver?
	var bm interface{} = NewBlinkMDriver(newI2cTestAdaptor("adaptor"), "bot")
	_, ok := bm.(*BlinkMDriver)
	if !ok {
		t.Errorf("NewBlinkMDriver() should have returned a *BlinkMDriver")
	}
}

// Commands
func TestNewBlinkMDriverCommands_Rgb(t *testing.T) {
	blinkM := initTestBlinkMDriver()

	result := blinkM.Driver.Command("Rgb")(rgb)
	gobot.Assert(t, result, nil)
}

func TestNewBlinkMDriverCommands_Fade(t *testing.T) {
	blinkM := initTestBlinkMDriver()

	result := blinkM.Driver.Command("Fade")(rgb)
	gobot.Assert(t, result, nil)
}

func TestNewBlinkMDriverCommands_FirmwareVersion(t *testing.T) {
	blinkM, adaptor := initTestBlinkDriverWithStubbedAdaptor()

	param := make(map[string]interface{})

	// When len(data) is 2
	adaptor.i2cReadImpl = func() []byte {
		return []byte{99, 1}
	}

	result := blinkM.Driver.Command("FirmwareVersion")(param)

	version, _ := blinkM.FirmwareVersion()
	gobot.Assert(t, result.(map[string]interface{})["version"].(string), version)

	// When len(data) is not 2
	adaptor.i2cReadImpl = func() []byte {
		return []byte{99}
	}
	result = blinkM.Driver.Command("FirmwareVersion")(param)

	version, _ = blinkM.FirmwareVersion()
	gobot.Assert(t, result.(map[string]interface{})["version"].(string), version)
}

func TestNewBlinkMDriverCommands_Color(t *testing.T) {
	blinkM := initTestBlinkMDriver()

	param := make(map[string]interface{})

	result := blinkM.Driver.Command("Color")(param)

	color, _ := blinkM.Color()
	gobot.Assert(t, result.(map[string]interface{})["color"].([]byte), color)
}

// Methods
func TestBlinkMDriverStart(t *testing.T) {
	blinkM := initTestBlinkMDriver()

	gobot.Assert(t, len(blinkM.Start()), 0)
}

func TestBlinkMDriverHalt(t *testing.T) {
	blinkM := initTestBlinkMDriver()
	gobot.Assert(t, len(blinkM.Halt()), 0)
}

func TestBlinkMDriverFirmwareVersion(t *testing.T) {
	blinkM, adaptor := initTestBlinkDriverWithStubbedAdaptor()

	// when len(data) is 2
	adaptor.i2cReadImpl = func() []byte {
		return []byte{99, 1}
	}

	version, _ := blinkM.FirmwareVersion()
	gobot.Assert(t, version, "99.1")

	// when len(data) is not 2
	adaptor.i2cReadImpl = func() []byte {
		return []byte{99}
	}

	version, _ = blinkM.FirmwareVersion()
	gobot.Assert(t, version, "")
}

func TestBlinkMDriverColor(t *testing.T) {
	blinkM, adaptor := initTestBlinkDriverWithStubbedAdaptor()

	// when len(data) is 3
	adaptor.i2cReadImpl = func() []byte {
		return []byte{99, 1, 2}
	}

	color, _ := blinkM.Color()
	gobot.Assert(t, color, []byte{99, 1, 2})

	// when len(data) is not 3
	adaptor.i2cReadImpl = func() []byte {
		return []byte{99}
	}

	color, _ = blinkM.Color()
	gobot.Assert(t, color, []byte{})

}

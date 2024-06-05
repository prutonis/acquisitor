package cfg

var AcqConfig Config

type Conversion struct {
	Unit   string  `yaml:"unit"`
	Factor float32 `yaml:"factor"`
}

type AdcInput struct {
	Name    string     `yaml:"name"`
	Channel uint16     `yaml:"channel"`
	Gain    uint16     `yaml:"gain"`
	Conv    Conversion `yaml:"conv"`
}

type Adc struct {
	Name      string     `yaml:"name"`
	Type      string     `yaml:"type"`
	I2cBus    int        `yaml:"i2cbus"`
	I2cAddr   int        `yaml:"i2caddr"`
	ConvDelay int        `yaml:"convdelay"`
	Inputs    []AdcInput `yaml:"inputs"`
}

type Hardware struct {
	Adc Adc `yaml:"adc"`
}

type Telemetry struct {
	Server Server `yaml:"server"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	User string `yaml:"user"`
}

type Config struct {
	Hardware  Hardware  `yaml:"hardware"`
	Telemetry Telemetry `yaml:"telemetry"`
}

// PGA_6_144 = 0 // Full Scale Range = +/- 6.144V
// PGA_4_096 = 1 // Full Scale Range = +/- 4.096V
// PGA_2_048 = 2 // Full Scale Range = +/- 2.048V
// PGA_1_024 = 3 // Full Scale Range = +/- 1.024V
// PGA_0_512 = 4 // Full Scale Range = +/- 0.512V
// PGA_0_256 = 5 // Full Scale Range = +/- 0.128V

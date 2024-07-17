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
	Enabled   bool       `yaml:"enabled"`
}

type Hardware struct {
	Adc Adc `yaml:"adc"`
}

type Telemetry struct {
	Server     Server      `yaml:"server"`
	Collectors []Collector `yaml:"collectors"`
	Pusher     Pusher      `yaml:"pusher"`
}

type Server struct {
	Host  string `yaml:"host"`
	Port  int    `yaml:"port"`
	User  string `yaml:"user"`
	Topic string `yaml:"topic"`
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

// create struct for telemetry data configured in acq-conf.yaml
type Collector struct {
	Name     string `yaml:"name"`
	Enabled  bool   `yaml:"enabled"`
	Interval int    `yaml:"interval"`
	Keys     []CollectorKey
}

type CollectorKey struct {
	Name   string  `yaml:"name"`
	Unit   string  `yaml:"unit"`
	Type   int     `yaml:"type"`
	Median bool    `yaml:"median"`
	Source string  `yaml:"source"`
	Factor float32 `yaml:"factor"`
}

type Pusher struct {
	Enabled   bool `yaml:"enabled"`
	Interval  int  `yaml:"interval"`
	Precision int  `yaml:"precision"`
	Keys      []PusherKey
}

type PusherKey struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
}

func (t *Telemetry) ResolveCollector(name string) *Collector {
	var collector *Collector = nil
	for _, c := range t.Collectors {
		if c.Name == name {
			return &c
		}
	}
	return collector
}

func (c *Collector) ResolveKey(name string) *CollectorKey {
	var key *CollectorKey = nil
	for _, k := range c.Keys {
		if k.Name == name {
			return &k
		}
	}
	return key
}

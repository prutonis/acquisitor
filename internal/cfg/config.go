package cfg

var AcqConfig Config

type AdcInput struct {
	Name    string `yaml:"name"`
	Channel uint16 `yaml:"channel"`
	Gain    uint16 `yaml:"gain"`
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

type GpioPin struct {
	Name    string `yaml:"name"`
	Pin     int    `yaml:"pin"`
	Default int    `yaml:"default"`
}

type Gpio struct {
	Enabled bool      `yaml:"enabled"`
	Pins    []GpioPin `yaml:"pins"`
}

// Charger is the solar charge controller read over Modbus RTU.
type Charger struct {
	Enabled bool   `yaml:"enabled"`
	Device  string `yaml:"device"`  // serial port, e.g. /dev/ttyS2 ("fake" = simulator)
	Baud    int    `yaml:"baud"`    // default 9600
	SlaveId int    `yaml:"slaveid"` // Modbus address 1..247, 255 = "FFH" bypass address (default 1)
	Timeout int    `yaml:"timeout"` // response timeout in ms (default 1000)
	// MaxGap: registers closer than this many words are read in one request
	// (unused words in between are read and ignored). 0 = only adjacent registers.
	MaxGap    int               `yaml:"maxgap"`
	Registers []ChargerRegister `yaml:"registers"`
}

// ChargerRegister describes one value to be read from the controller (function 03).
//
//	raw = (word(s) >> shift) & mask
//
// Words are big endian; for 32 bit types the high word is at `address`.
type ChargerRegister struct {
	Name    string `yaml:"name"`
	Address int    `yaml:"address"`
	Type    string `yaml:"type"`  // u16 (default), i16, u32, i32
	Shift   uint   `yaml:"shift"` // bits to shift right (default 0)
	Mask    uint32 `yaml:"mask"`  // 0 = no mask. Signed types are applied only without shift/mask
}

type Hardware struct {
	Adc     Adc     `yaml:"adc"`
	Gpio    Gpio    `yaml:"gpio"`
	Charger Charger `yaml:"charger"`
}

type Telemetry struct {
	Server     Server      `yaml:"server"`
	Collectors []Collector `yaml:"collectors"`
	Pusher     Pusher      `yaml:"pusher"`
}

type Logging struct {
	File  string `yaml:"file"`
	Level string `yaml:"level"`
}

type Server struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	PublishTopic   string `yaml:"publishTopic"`
	SubscribeTopic string `yaml:"subscribeTopic"`
	ResponseTopic  string `yaml:"responseTopic"`
}

type Config struct {
	Hardware  Hardware  `yaml:"hardware"`
	Telemetry Telemetry `yaml:"telemetry"`
	Logging   Logging   `yaml:"logging"`
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
	Name     string `yaml:"name"`
	Unit     string `yaml:"unit"`
	Type     int    `yaml:"type"`
	Median   bool   `yaml:"median"`
	Source   string `yaml:"source"`
	Function string `yaml:"function"`
	// text mapping (charger collector). Applied on the result of Function:
	//   values: {"2": mppt}                -> the number becomes a text
	//   bits:   {"0": overdischarge, ...}  -> comma separated names of the bits set
	//   empty:  text when no listed bit is set (default "ok")
	Values map[string]string `yaml:"values"`
	Bits   map[string]string `yaml:"bits"`
	Empty  string            `yaml:"empty"`
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

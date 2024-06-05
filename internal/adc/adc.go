package adc

type AdcOps interface {
	Init() error
	ReadValue(channel int) (int16, error)
	Close() error
}

type Adc struct {
	adsOps AdsOps
}

func NewAdc(ops AdsOps) AdcOps {
	return &Adc{adsOps: ops}
}

func (a *Adc) Init() error {
	return a.adsOps.SetCurrentChannel(0)
}

func (a *Adc) ReadValue(channel int) (int16, error) {
	a.adsOps.SetCurrentChannel(channel)
	return a.adsOps.ReadValue()
}

func (a *Adc) Close() error {
	return a.adsOps.Close()
}

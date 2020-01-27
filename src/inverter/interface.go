package inverter

type Inverter interface {
	Process(*[]byte) ([]byte, error)
}

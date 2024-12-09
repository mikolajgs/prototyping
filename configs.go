package prototyping

type Config struct {
	DatabaseDSN        string
	UserConstructor    func() interface{}
	SessionConstructor func() interface{}
}

package model

type InvalidMethodeCPH struct {
	s string
}

func (e *InvalidMethodeCPH) Error() string {
	return e.s
}

func NewInvalidMethodeCPH(msg string) error {
	return &InvalidMethodeCPH{s: msg}
}

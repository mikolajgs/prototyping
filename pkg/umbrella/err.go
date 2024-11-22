package umbrella

// ErrUmbrella wraps original error that occurred in Err with name of the
// operation/step that failed, which is in Op field
type ErrUmbrella struct {
	Op  string
	Err error
}

func (e *ErrUmbrella) Error() string {
	return e.Err.Error()
}

func (e *ErrUmbrella) Unwrap() error {
	return e.Err
}

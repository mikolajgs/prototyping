package struct2sql

// ErrStruct2sql wraps original error with operation/step where the error occured and optionally with a tag when
// parsing "crud" failed
type ErrStruct2sql struct {
	Op  string
	Tag string
	Err error
}

func (e ErrStruct2sql) Error() string {
	return e.Err.Error()
}

func (e ErrStruct2sql) Unwrap() error {
	return e.Err
}

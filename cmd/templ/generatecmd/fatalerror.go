package generatecmd

type FatalError struct {
	Err error
}

func (e FatalError) Error() string {
	return e.Err.Error()
}

func (e FatalError) Unwrap() error {
	return e.Err
}

func (e FatalError) Is(target error) bool {
	_, ok := target.(FatalError)
	return ok
}

func (e FatalError) As(target any) bool {
	_, ok := target.(*FatalError)
	return ok
}

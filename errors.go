package multiio

type errOutOfRange struct{}

func (errOutOfRange) Error() string {
	return "out of range"
}

func (errOutOfRange) InvalidArgument() {}

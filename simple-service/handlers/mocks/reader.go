package mocks

import "errors"

type ErrReader struct{}

func (r ErrReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("test error")
}

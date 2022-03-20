package logging

import (
	"io"
	"log"
	"os"

	"github.com/KaiserWerk/go-log-rotator"
)

func New(dir string) (*log.Logger, func() error, error) {
	var (
		rot     *rotator.Rotator
		err     error
		cleanup = func() error {
			return rot.Close()
		}
	)
	rot, err = rotator.New(dir, "phproxy.log", 3<<20, 0664, 0, true)
	if err != nil {
		return nil, nil, err
	}

	logger := log.New(io.MultiWriter(os.Stdout, rot), "", 0)
	return logger, cleanup, nil
}

package routine

import (
	logger "g_matching_engine/pkg/mlog"

	"go.uber.org/zap"
)

func GoSafe(fn func()) {
	go RunSafe(fn, nil)
}

func Serial(fns ...func()) func() {
	return func() {
		for _, fn := range fns {
			fn()
		}
	}
}

func RunSafe(fn func(), cleanups ...func()) {
	if cleanups != nil {
		defer func() {
			for _, cleaner := range cleanups {
				cleaner()
			}
		}()
	}
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error("recover", zap.Any("err", err))
		}
	}()
	fn()
}

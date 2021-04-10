package logrotator

import (
	"github.com/lestrrat/go-strftime"
	"github.com/pkg/errors"
	"time"
)

func NewLogScissors(pattern string, period time.Duration) (*LogScissors, error) {
	strfobj, err := strftime.New(pattern)
	if err != nil {
		return nil, errors.Wrap(err, `invalid time duration pattern`)
	}

	var tw LogScissors
	tw.pattern = strfobj
	tw.period = period.Nanoseconds()
	_, offset := time.Now().Zone()
	tw.timeDiffToUTC = (time.Duration(offset) * time.Second).Nanoseconds()

	return &tw, nil
}

func NewLogScissorsWithPreFilename(pattern string, period time.Duration, preFilename string) (*LogScissors, error) {
	strfobj, err := strftime.New(pattern)
	if err != nil {
		return nil, errors.Wrap(err, `invalid time duration pattern`)
	}

	var tw LogScissors
	tw.pattern = strfobj
	tw.period = period.Nanoseconds()
	_, offset := time.Now().Zone()
	tw.timeDiffToUTC = (time.Duration(offset) * time.Second).Nanoseconds()
	tw.preFilename = preFilename

	return &tw, nil
}

func NewLogCleaner(pattern string, maxAge time.Duration) (*LogCleaner, error) {
	var tc LogCleaner
	tc.pattern = pattern
	if maxAge < 0 {
		return nil, errors.New(`maxAge must be greater than 0`)
	}
	tc.maxAge = maxAge

	return &tc, nil
}

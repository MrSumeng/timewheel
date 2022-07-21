/*
Copyright 2022 The KubePort Authors.
*/

package timewheel

import "time"

const (
	DefaultSlotNum   int64 = 3600
	DefaultPrecision       = time.Second
)

type Options struct {
	precision time.Duration
	slotNum   int64
}

var DefaultOptions = defaultOptions()

func defaultOptions() *Options {
	return &Options{
		precision: DefaultPrecision,
		slotNum:   DefaultSlotNum,
	}
}

func (o *Options) SetSlotNum(slotNum int64) {
	if slotNum > 0 {
		o.slotNum = slotNum
	}
}

func (o *Options) GetSlotNum() int64 {
	return o.slotNum
}

func (o *Options) SetPrecision(precision time.Duration) {
	if precision > 0 {
		o.precision = precision
	}
}

func (o *Options) GetPrecision() time.Duration {
	return o.precision
}

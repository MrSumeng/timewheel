/*
Copyright 2022 The KubePort Authors.
*/

package timewheel

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

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

type TaskOption func(*TaskOptions)

type TaskOptions struct {
	Key    interface{}
	Times  int64
	Jitter bool
	Period time.Duration
}

func defaultTaskOpts() *TaskOptions {
	return &TaskOptions{
		Key:    uuid.NewV4().String(),
		Times:  1,
		Jitter: false,
		Period: time.Second,
	}
}

func WithKey(key interface{}) TaskOption {
	return func(o *TaskOptions) {
		if key != nil {
			o.Key = key
		}
	}
}

func WithTimes(times int64) TaskOption {
	return func(o *TaskOptions) {
		if times >= -1 && times != TimesStop {
			o.Times = times
		}
	}
}

func WithJitter(jitter bool) TaskOption {
	return func(o *TaskOptions) {
		o.Jitter = jitter
	}
}

func WithPeriod(period time.Duration) TaskOption {
	return func(o *TaskOptions) {
		o.Period = period
	}
}

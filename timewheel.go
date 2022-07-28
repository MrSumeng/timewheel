package timewheel

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

const (
	TimesForever int64 = -1
	TimesStop    int64 = 0

	PrecisionThreshold int64 = 2
)

type TimeWheel interface {
	Start()
	Stop()
	After(job Job, after time.Duration) (interface{}, error)
	Add(job Job, options ...TaskOption) (interface{}, error)
	Remove(key interface{}) error
}

type Job func()

type timewheel struct {
	opt        *Options
	ticker     *time.Ticker
	slots      []*list.List
	jobs       []int64
	currentPos int64
	stopCh     chan struct{}
	taskRecord map[interface{}]*task
	lock       sync.RWMutex
}

// New create an empty time timewheel
func New(opt *Options) TimeWheel {
	if opt == nil {
		opt = DefaultOptions
	}
	if opt.precision <= 0 || opt.slotNum <= 0 {
		return nil
	}

	tw := &timewheel{
		opt:        opt,
		slots:      make([]*list.List, opt.slotNum),
		jobs:       make([]int64, opt.slotNum),
		currentPos: 0,
		stopCh:     make(chan struct{}),
		taskRecord: make(map[interface{}]*task),
	}

	tw.init()

	return tw
}

// Start  the time timewheel
func (tw *timewheel) Start() {
	tw.ticker = time.NewTicker(tw.opt.precision)
	go tw.start()
}

func (tw *timewheel) start() {
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case <-tw.stopCh:
			tw.ticker.Stop()
			return
		}
	}
}

func (tw *timewheel) After(job Job, after time.Duration) (interface{}, error) {
	return tw.Add(job, WithPeriod(after))
}

// Add new task to the time wheel
func (tw *timewheel) Add(job Job, options ...TaskOption) (interface{}, error) {
	opts := defaultTaskOpts()
	for _, option := range options {
		option(opts)
	}

	// if task period less than time timewheel precision,
	// we will ignore this task or let the task interval equal the time timewheel precision.
	// the threshold is PrecisionThreshold
	if tw.opt.precision > opts.Period {
		if int64(tw.opt.precision/opts.Period) < PrecisionThreshold {
			opts.Period = tw.opt.precision
		} else {
			return nil, errors.New("the task period is much smaller than the time timewheel precision")
		}
	}

	tw.lock.Lock()
	defer tw.lock.Unlock()
	_, ok := tw.taskRecord[opts.Key]

	if ok {
		return nil, errors.New("duplicate task key")
	}

	// 60/2/60=0
	// 60/1/60=1
	// 90/1/60=1
	cycle := int64(opts.Period/tw.opt.precision) / tw.opt.slotNum
	// 60/2%60=30
	// 60/1%60=0
	// 90/1%60=30
	_period := int64(opts.Period/tw.opt.precision) % tw.opt.slotNum

	_task := &task{
		key:    opts.Key,
		period: _period,
		times:  opts.Times,
		cycle:  cycle,
		start:  tw.currentPos,
		pos:    tw.currentPos,
		jitter: opts.Jitter,
		job:    job,
	}
	tw.taskRecord[opts.Key] = _task

	pos := _task.next(tw.opt.slotNum, tw.jobs)
	tw.slots[pos].PushBack(_task)
	tw.jobs[pos] = int64(tw.slots[pos].Len())
	return opts.Key, nil
}

// Remove the task from time wheel
func (tw *timewheel) Remove(key interface{}) error {
	if key == "" {
		return nil
	}
	tw.lock.Lock()
	defer tw.lock.Unlock()
	task, ok := tw.taskRecord[key]

	if !ok {
		return errors.New("task not exists, please check you task key")
	} else {
		// lazy remove task
		task.times = 0
		delete(tw.taskRecord, key)
	}
	return nil
}

// time timewheel initialize
func (tw *timewheel) init() {
	for i := 0; i < int(tw.opt.slotNum); i++ {
		tw.slots[i] = list.New()
		tw.jobs[i] = 0
	}
}

//
func (tw *timewheel) tickHandler() {
	l := tw.slots[tw.currentPos]
	tw.scanAndRun(l, tw.currentPos)

	if tw.currentPos == tw.opt.slotNum-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
}

// scan task list and run the task
func (tw *timewheel) scanAndRun(l *list.List, currentPos int64) {

	if l == nil || l.Len() == 0 {
		return
	}

	for item := l.Front(); item != nil; {
		_task := item.Value.(*task)
		next := item.Next()

		if _task.cycle > _task.runCycle {
			_task.runCycle++
			item = next
			continue
		}

		l.Remove(item)
		item = next
		tw.jobs[currentPos] = int64(l.Len())

		if _task.times > 0 {
			go _task.job()
			_task.times--
		}

		if _task.times == TimesForever {
			go _task.job()
		}

		if _task.times == 0 {
			tw.lock.Lock()
			delete(tw.taskRecord, _task.key)
			tw.lock.Unlock()
			continue
		}

		pos := _task.next(tw.opt.slotNum, tw.jobs)

		tw.slots[pos].PushBack(_task)
		tw.jobs[pos] = int64(tw.slots[pos].Len())

	}

}

// Stop  the time timewheel
func (tw *timewheel) Stop() {
	tw.stopCh <- struct{}{}
}

type task struct {
	key interface{}
	//period The time period that needs to be traveled on the time timewheel each time the task is executed
	period int64
	// run times
	times int64 //-1:no limit >=1:run times
	//cycle is how many cycles need run
	cycle int64
	//runCycle is already running cycles
	runCycle int64
	//start is task next run calculation position
	start int64
	//pos is task next run true position
	pos int64
	//if jitter is true，it means that the task can jitter within the jitter range.
	//the jitter range indicates the delay range allowed for the task execution
	jitter bool
	job    Job
}

//calculate next run time
func (t *task) next(slotNum int64, jobs []int64) int64 {
	t.runCycle = 0

	max := t.start + t.period
	if !t.jitter {
		t.start = max % slotNum
		t.pos = t.start
		return t.pos
	}

	min := t.start

	repeatJobs := jobs[:]
	repeatJobs = append(repeatJobs, jobs...)
	repeatJobs = repeatJobs[min+1 : max+1]

	index := 0
	jobMin := repeatJobs[0]
	for i, job := range repeatJobs {
		if jobMin > job {
			index = i
			jobMin = job
		}
	}
	t.pos = (t.start + int64(index+1)) % slotNum
	t.start = max % slotNum
	t.runCycle = 0

	return t.pos
}

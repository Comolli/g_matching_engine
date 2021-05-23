package time

import (
	"errors"
	"sync"
	"time"

	"github.com/golang/glog"
)

var ErrNoInstance = errors.New("ErrNoInstance")
var ErrExpireTimerNoFunc = errors.New("ErrExpireTimerNoFunc")
var Duration time.Duration = time.Duration(1<<63 - 1)

const timerFormat = "2000-01-02 10:00:00"

type TimerSilce struct {
	Key    string
	fn     func()
	index  int
	expire time.Time
	next   *TimerSilce
}

type Timer struct {
	lock   sync.Mutex
	free   *TimerSilce
	signal *time.Timer
	timers []*TimerSilce
	num    int
}

func (t *Timer) init(num int) {
	t.signal = time.NewTimer(Duration)
	t.timers = make([]*TimerSilce, 0, num)
	t.num = num
	t.create()
	go t.start()
}

//get the free timerSilce
func (t *Timer) get() (ts *TimerSilce) {
	if ts = t.free; ts == nil {
		t.create()
		ts = t.free
	}
	t.free = ts.next
	return
}

func (t *Timer) create() {
	var i int
	var ts *TimerSilce
	tss := make([]TimerSilce, t.num)
	t.free = &(tss[0])
	ts = t.free
	for i = 1; i < t.num; i++ {
		ts.next = &(tss[i])
		ts = ts.next
	}
	ts.next = nil
}

//
func (t *Timer) Add(expire time.Duration, fn func()) (ts *TimerSilce) {
	t.lock.Lock()
	ts = t.get()
	ts.expire = time.Now().Add(expire)
	ts.fn = fn
	t.add(ts)
	t.lock.Unlock()
	return
}

func (t *Timer) start() {
	for {
		t.expire()
		<-t.signal.C
	}
}

func (t *Timer) expire() {
	var ts *TimerSilce
	var d time.Duration
	t.lock.Lock()
	for {
		if len(t.timers) == 0 {
			d = Duration
			glog.Info(ErrNoInstance)
			break
		}
		ts = t.timers[0]
		if d = ts.Delay(); d > 0 {
			break
		}
		t.del(ts)
		t.lock.Unlock()
		switch temp := (ts.fn == nil); temp {
		case true:
			glog.Warning(ErrExpireTimerNoFunc)
		case false:
			glog.Info("timer key: %s, expire: %s, index: %d expired, call fn", ts.Key, ts.ExpireString(), ts.index)
			ts.fn()
		}
		t.lock.Lock()
	}
	t.signal.Reset(d)
	t.lock.Unlock()
}

func (ts *TimerSilce) Delay() time.Duration {
	return time.Until(ts.expire)
}

func (t *Timer) del(ts *TimerSilce) {
	//return time.Until(ts.expire)
}

func (t *Timer) add(ts *TimerSilce) {

}

// ExpireString expire string.
func (td *TimerSilce) ExpireString() string {
	return td.expire.Format(timerFormat)
}

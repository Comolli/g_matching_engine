package etcdv3

import (
	"context"
	"g_matching_engine/pkg/discovery/etcdv3/connectivity"
	"sync"

	"github.com/coreos/etcd/clientv3"
)

type etcdConn interface {
	GetState() connectivity.State
	WaitForStateChange(ctx context.Context, sourceState connectivity.State) bool
}

type Watcher struct {
	*sync.Mutex
	revision  uint64
	eventChan chan *clientv3.Event
	states    connectivity.State
	cancel    context.CancelFunc
	connected bool
	cb        []func()
}

func (w *Watcher) Eventchan() chan *clientv3.Event {
	return w.eventChan
}
func (w *Watcher) addListener(l func()) {
	w.Lock()
	w.cb = append(w.cb, l)
	w.Unlock()
}

func (w *Watcher) Close() error {
	if w.cancel != nil {
		w.cancel()
	}
	return nil
}
func (client *Client) WatchPrefix(ctx context.Context, prefix string) (*Watcher, error) {

}
func (w *Watcher) execute() {
	w.Lock()
	for _, fn := range w.cb {
		fn()
	}
	w.Unlock()
}

func (w *Watcher) updateState(conn etcdConn) {
	w.states = conn.GetState()
	switch w.states {
	case connectivity.TransientFailure, connectivity.Shutdown:
		w.connected = false
	case connectivity.Ready:
		if w.connected {
			w.connected = true
			w.execute()
		}
	}
}

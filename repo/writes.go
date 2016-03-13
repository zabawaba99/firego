package repo

import "sync"

type write struct {
	action action
	data   interface{}
}

type pendingWrites struct {
	writes     map[float64]write
	writeCount float64
	writeMtx   sync.RWMutex
}

func newPendingWrites() *pendingWrites {
	return &pendingWrites{
		writes: map[float64]write{},
	}
}

func (pw *pendingWrites) add(a action, d interface{}) float64 {
	w := write{
		action: a,
		data:   d,
	}

	pw.writeMtx.Lock()
	id := pw.writeCount
	pw.writeCount++

	pw.writes[id] = w
	pw.writeMtx.Unlock()

	return id
}

func (pw *pendingWrites) get(id float64) write {
	pw.writeMtx.RLock()
	w := pw.writes[id]
	pw.writeMtx.RUnlock()
	return w
}

func (pw *pendingWrites) delete(id float64) {
	pw.writeMtx.Lock()
	delete(pw.writes, id)
	pw.writeMtx.Unlock()
}

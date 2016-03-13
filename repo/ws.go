package repo

import (
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var connClosedErr = errors.New("connection closed")

type connState string

const (
	connecting   connState = "connecting"
	connected              = "connected"
	disconnected           = "disconnected"
)

type ws struct {
	cb  wsCallback
	u   *url.URL
	c   *websocket.Conn
	mtx sync.Mutex

	state    connState
	stateMtx sync.RWMutex

	callbacksMtx sync.Mutex
	callbacks    map[float64]func(error)

	done chan struct{}
}

func newWS(u *url.URL, cb wsCallback) *ws {
	return &ws{
		u:         u,
		cb:        cb,
		state:     connecting,
		callbacks: map[float64]func(error){},
		done:      make(chan struct{}),
	}
}

func (w *ws) dial() error {
	w.stateMtx.Lock()
	defer w.stateMtx.Unlock()

	dLogField("dialing websocket", "uri", w.u)
	c, _, err := websocket.DefaultDialer.Dial(w.u.String(), nil)
	if err != nil {
		w.state = disconnected
		return err
	}
	w.c = c

	go w.keepAlive()
	go w.readMessages()
	return nil
}

func (w *ws) keepAlive() {
	for {
		select {
		case <-time.After(45 * time.Second):
			dLog("Sending keep alive")
			w.c.WriteMessage(websocket.TextMessage, []byte("0"))
		case <-w.done:
			return
		}
	}
}

func (w *ws) readMessages() {
	for {
		var msg wsMsg
		err := w.c.ReadJSON(&msg)
		if err != nil {
			eLog("Could not read message", err)
			w.redial(w.u.Host)
			return
		}

		switch msg.Type {
		case "c":
			err := w.handleControl(msg.Data)
			if err == connClosedErr {
				return
			}
			if err != nil {
				eLog("Error handling control msg", err)
			}
		case "e":
			err = errors.New(msg.Data.getString("e"))
			fallthrough
		case "d":
			reqNum := msg.Data.getFloat("r")

			w.callbacksMtx.Lock()
			fn := w.callbacks[reqNum]
			delete(w.callbacks, reqNum)
			w.callbacksMtx.Unlock()

			if fn != nil {
				fn(err)
			}
		default:
			wLogField("Received unknown message type", "msg", msg)
		}
	}
}

func (w *ws) handleControl(data wsData) error {
	t := data.getString("t")
	switch t {
	// redirect
	case "r":
		newHost := data.getString("d")
		w.redial(newHost)
		return connClosedErr
	case "h":
		idata := data.getInternalData()
		ts := idata.getFloat("ts")
		newHost := idata.getString("h")
		sessionID := idata.getString("s")

		w.u.Host = newHost
		w.setState(connected)
		w.cb.ready(ts, sessionID)

	default:
		wLogField("Received unknown control type", "data", data)
	}

	return nil
}

func (w *ws) send(m sendMsg, fn func(err error)) {
	w.callbacksMtx.Lock()
	w.callbacks[m.ReqNum] = fn
	w.callbacksMtx.Unlock()

	type msg struct {
		Type string  `json:"t"`
		Data sendMsg `json:"d"`
	}
	v := msg{
		Type: "d",
		Data: m,
	}

	iLogField("Sending message", "msg", v)
	err := w.c.WriteJSON(v)
	if err != nil {
		eLog("Error sending stats", err)
	}
}

func (w *ws) close() {
	w.stateMtx.Lock()

	dLog("Closing connection")
	w.c.Close()
	close(w.done)
	w.done = make(chan struct{})
	w.state = disconnected

	w.stateMtx.Unlock()
}

func (w *ws) redial(newHost string) {
	dLogFields("redialing connection", map[string]interface{}{"oldHost": w.u.Host, "newHost": newHost})
	w.u.Host = newHost
	w.close()
	w.dial()
}

func (w *ws) getState() connState {
	w.stateMtx.RLock()
	s := w.state
	w.stateMtx.RUnlock()
	return s
}

func (w *ws) setState(s connState) {
	w.stateMtx.Lock()
	w.state = s
	w.stateMtx.Unlock()
}

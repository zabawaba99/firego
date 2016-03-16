package repo

import (
	"net/url"
	"sort"
	"strings"
)

type action string

const (
	actionStatus action = "s"
	actionPut           = "p"
	actionMerge         = "m"
)

type wsCallback interface {
	ready(ts float64, sessionID string)
}

type conn struct {
	u  *url.URL
	ws *ws

	writes *pendingWrites

	requestCount    float64
	firstConnection bool
}

func newConn(u *url.URL) *conn {
	u.Scheme = strings.Replace(u.Scheme, "http", "ws", 1)
	u.Path = ".ws"
	internalHost := strings.Split(u.Host, ".")[0]
	u.RawQuery = "ns=" + internalHost + "&v=5"

	c := &conn{
		u:      u,
		writes: newPendingWrites(),
	}
	c.ws = newWS(u, c)
	return c
}

func (c *conn) Dial() error {
	if err := c.ws.dial(); err != nil {
		return err
	}

	return nil
}

func (c *conn) ready(ts float64, sessionID string) {
	c.u.RawQuery += "&ls=" + sessionID
	dLogField("Conn is ready", "u", c.u)

	if !c.firstConnection {
		c.sendStatus()
	}

	c.firstConnection = true

	c.sendPendingWrites()
}

func (c *conn) sendPendingWrites() {
	c.writes.writeMtx.RLock()
	keys := make([]float64, len(c.writes.writes))
	var index int
	for k, _ := range c.writes.writes {
		keys[index] = k
		index++
	}

	sort.Float64s(keys)

	for _, k := range keys {
		c.sendWrite(k)
	}

	c.writes.writeMtx.RUnlock()
}

func (c *conn) put(path string, data interface{}) {
	c.createWrite(actionPut, path, data)
}

func (c *conn) merge(path string, data map[string]interface{}) {
	c.createWrite(actionMerge, path, data)
}

func (c *conn) createWrite(a action, p string, d interface{}) {
	put := writeData{Path: p, Data: d}
	id := c.writes.add(a, put)

	if c.ws.getState() == connected {
		c.sendWrite(id)
	}
}

func (c *conn) sendWrite(id float64) {
	w := c.writes.get(id)
	c.send(w.action, w.data, func(err error) {
		if err != nil {
			eLog("Error sending write", err)
		}

		iLogField("Successfully wrote op", "w", w)
		c.writes.delete(id)
	})
}

type statusMsg struct {
	Control struct {
		Version int `json:"sdk.java.2-5-2"`
	} `json:"c"`
}

func (c *conn) sendStatus() {
	data := map[string]interface{}{
		"c": map[string]int{
			"sdk.java.2-5-2": 1,
		},
	}
	c.send(actionStatus, data, func(err error) {
		if err != nil {
			eLog("Error sending status", err)
		}
		dLog("Status sent successfully")
	})
}

func (c *conn) reqNum() float64 {
	n := c.requestCount
	c.requestCount++
	return n
}

func (c *conn) send(a action, data interface{}, fn func(err error)) {
	m := sendMsg{
		Action: a,
		ReqNum: c.reqNum(),
		Data:   data,
	}

	c.ws.send(m, fn)
}

package repo

import (
	"net/url"
	"strings"
)

type Repo struct {
	secure       bool
	namespace    string
	host         string
	internalHost string

	conn *conn
}

func New(uri *url.URL) *Repo {
	r := &Repo{
		secure:       uri.Scheme == "https",
		namespace:    strings.Split(uri.Host, ".")[0],
		host:         uri.Host,
		internalHost: uri.Host,
	}
	r.conn = newConn(r.URL())

	go r.connect()
	return r
}

func (r *Repo) connect() {
	if err := r.conn.Dial(); err != nil {
		eLog("Could not establish connection", err)
	}
}

func (r Repo) Host() string {
	scheme := "http"
	if r.secure {
		scheme += "s"
	}
	return scheme + "://" + r.host
}

func (r *Repo) URL() *url.URL {
	u := &url.URL{
		Scheme: "ws",
		Host:   r.internalHost,
		Path:   ".ws",
	}
	if r.secure {
		u.Scheme += "s"
	}

	u.RawQuery = "ns=" + r.internalHost + "&v=5"
	return u
}

func (r *Repo) SetValue(p string, v interface{}) {
	// TODO validate that the value meets restrictions

	// TODO set value into local tree

	// TODO broadcast notification for listeners

	// write values to firebase
	r.conn.put(p, v)
}

func (r *Repo) UpdateChildren(p string, children map[string]interface{}) {
	// TODO validate that the value meets restrictions

	// TODO set value into local tree

	// TODO broadcast notification for listeners

	// write values to firebase
	r.conn.merge(p, children)
}

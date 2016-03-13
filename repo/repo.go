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
	return r.host
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
	// validate that the value meets restrictions
	r.conn.put(p, v)
}

# Firego
---
[![Build Status](https://travis-ci.org/CloudCom/firego.svg?branch=master)](https://travis-ci.org/CloudCom/firego) [![Coverage Status](https://coveralls.io/repos/CloudCom/firego/badge.svg)](https://coveralls.io/r/CloudCom/firego)
---

A Firebase client written in Go

## Installation

```bash
go get -u github.com/CloudCom/firego
```

## Usage

Import firego

```go
import "github.com/CloudCom/firego"
```

Create a new firego reference

```go
f := firego.New("https://my-firebase-app.firebaseIO.com")
```

### Request Timeouts

By default, the `Firebase` reference will timeout after 30 seconds of trying
to reach a Firebase server. You can configure this value by setting the global
timeout duration

```go
firego.TimeoutDuration = time.Minute
```

### Auth Tokens

```go
f.Auth("some-token-that-was-created-for-me")
f.Unauth()
```

Visit [Fireauth](https://github.com/CloudCom/fireauth) if you'd like to generate your own auth tokens

### Get Value

```go
var v map[string]interface{}
if err := f.Value(&v); err != nil {
  log.Fatal(err)
}
fmt.Printf("%s\n", v)
```

### Set Value

```go
v := map[string]string{"foo":"bar"}
if err := f.Set(v); err != nil {
  log.Fatal(err)
}
```

### Push Value

```go
v := "bar"
pushedFirego, err := f.Push(v)
if err != nil {
	log.Fatal(err)
}

var bar string
if err := pushedFirego.Value(&bar); err != nil {
	log.Fatal(err)
}

// prints "https://my-firebase-app.firebaseIO.com/-JgvLHXszP4xS0AUN-nI: bar"
fmt.Printf("%s: %s\n", pushedFirego, bar)
```

### Update Child

```go
v := map[string]string{"foo":"bar"}
if err := f.Update(v); err != nil {
  log.Fatal(err)
}
```

### Remove Value

```go
if err := f.Remove(); err != nil {
  log.Fatal(err)
}
```

### Watch a Node

```go
notifications := make(chan firego.Event)
if err := f.Watch(notifications); err != nil {
	log.Fatal(err)
}

defer f.StopWatching()
for event := range notifications {
	fmt.Printf("Event %#v\n", event)
}
fmt.Printf("Notifications have stopped")
```

Check the [GoDocs](http://godoc.org/github.com/CloudCom/firego) or
[Firebase Documentation](https://www.firebase.com/docs/rest/) for more details

## Running Tests

In order to run the tests you need to `go get`:

* `github.com/stretchr/testify/require`
* `github.com/stretchr/testify/assert`

## Issues Management

Feel free to open an issue if you come across any bugs or
if you'd like to request a new feature.

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b new-feature`)
3. Commit your changes (`git commit -am 'Some cool reflection'`)
4. Push to the branch (`git push origin new-feature`)
5. Create new Pull Request

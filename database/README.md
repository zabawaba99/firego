# Database
---

A Go package for the [Firebase Realtime Database](https://firebase.google.com/docs/database/) service.

## Installation

```bash
go get -u gopkg.in/zabawaba99/firego/database.v2
```

## Usage

Import firego

```go
import "gopkg.in/zabawaba99/firego/database.v2"
```

Create a new firego reference

```go
db := database.New("https://my-firebase-app.firebaseIO.com", nil)
```

with existing http client

```go
db := database.New("https://my-firebase-app.firebaseIO.com", client)
```

### Auth Tokens

```go
db.Auth("some-token-that-was-created-for-me")
db.Unauth()
```

Visit [Fireauth](https://github.com/zabawaba99/fireauth) if you'd like to generate your own auth tokens

### Get Value

```go
var v map[string]interface{}
if err := db.Value(&v); err != nil {
  log.Fatal(err)
}
fmt.Printf("%s\n", v)
```

#### Querying

Take a look at Firebase's [query parameters](https://www.firebase.com/docs/rest/guide/retrieving-data.html#section-rest-filtering)
for more information on what each function does.

```go
var v map[string]interface{}
if err := db.StartAt("a").EndAt("c").LimitToFirst(8).OrderBy("field").Value(&v); err != nil {
	log.Fatal(err)
}
fmt.Printf("%s\n", v)
```

### Set Value

```go
v := map[string]string{"foo":"bar"}
if err := db.Set(v); err != nil {
  log.Fatal(err)
}
```

### Push Value

```go
v := "bar"
pushedFirego, err := db.Push(v)
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
if err := db.Update(v); err != nil {
  log.Fatal(err)
}
```

### Remove Value

```go
if err := db.Remove(); err != nil {
  log.Fatal(err)
}
```

### Watch a Node

```go
notifications := make(chan firego.Event)
if err := db.Watch(notifications); err != nil {
	log.Fatal(err)
}

defer db.StopWatching()
for event := range notifications {
	fmt.Printf("Event %#v\n", event)
}
fmt.Printf("Notifications have stopped")
```

Check the [GoDocs](http://godoc.org/gopkg.in/zabawaba99/firego/database.v2-unstable) or
[Firebase Documentation](https://www.firebase.com/docs/rest/) for more details

## Running Tests

In order to run the tests you need to `go get -t ./...`
first to go-get the test dependencies.

## Issues Management

Feel free to open an issue if you come across any bugs or
if you'd like to request a new feature.

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b new-feature`)
3. Commit your changes (`git commit -am 'Some cool reflection'`)
4. Push to the branch (`git push origin new-feature`)
5. Create new Pull Request

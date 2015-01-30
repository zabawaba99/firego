# Firego

A Firebase client written in Go

##### Under Development
The API may or may not change radically within the next upcoming weeks. 

## Installation

```bash
go get -u github.com/zabawaba99/firego
```

## Usage

Import firego

```go
import "github.com/zabawaba99/firego"
```

Create a new firego reference

```go
f := firego.New("https://my-firebase-app.firebaseIO.com")
```

### Auth Tokens

```go
f.SetAuth("some-token-that-was-created-for-me")
f.RemoveAuth()
```

Visit [Fireauth](https://github.com/zabawaba99/fireauth) if you'd like to generate your own auth tokens

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

Check the [GoDocs](http://godoc.org/github.com/zabawaba99/firego) or 
[Firebase Documentation](https://www.firebase.com/docs/rest/) for more details

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b new-feature`)
3. Commit your changes (`git commit -am 'Some cool reflection'`)
4. Push to the branch (`git push origin new-feature`)
5. Create new Pull Request

# Messaging
---

A Go package for the [Firebase Cloud Messaging](https://firebase.google.com/docs/cloud-messaging/) service.

## Installation

```bash
go get -u gopkg.in/zabawaba99/firego/messaging.v2
```

## Usage

Import the messaing package

```go
import "gopkg.in/zabawaba99/firego/messaging.v2"
```

Create a new messaging instance backed by the default http client

```go
fcm := messaging.New("server-key", nil)
```

with existing http client

```go
fcm := messaing.New("server-client", client)
```

If you are not sure what your `server-key` is, you can find it [here](https://console.firebase.google.com/project/_/settings/cloudmessaging) and
read more about it [here](https://firebase.google.com/docs/cloud-messaging/server#auth).


### Sending a Message

You can send messages by building a `Message` struct and passing it to the client. For more information on the meaning of each field,
you can read [this](https://firebase.google.com/docs/cloud-messaging/http-server-ref#downstream-http-messages-json).

```go
msg := messaging.Message{
	RegistrationIDs: []string{"foo", "bar"},
	Data: &fcm.Data{
		"hello": "world",
	},
}
resp, err := fcm.Send(msg)
if err != nil {
	log.Fatalf("Error sending msg %s", err)
}

if resp.Failure != 0 {
	log.Fatal("All these reuslts failed to send %s", resp.Failures())
}
```
# Ravelin Coding Challenge

*The server is run by typing* `go run server/main.go` *at the root level.*

## Sessions

On server start, a `SessionManger` called `globalSessions` is created. It is defined as:

```go
// The struct and methods for a SessionManager
type SessionManager struct {
  name        string
  lock        sync.Mutex                   // protects session from races
  maxLifeTime int64                        // expiry of a session
  sessions    map[string]Session           // map of sid's to sessions
}
```

It maintains a map of session ids to `Session`s. It is protected by a mutex, as concurrent requests could overwrite
each other. The http package runs each handler in a new goroutine, so concurrent api requests (should) be supported.

A `Session` is defined as follows:

```go
// A session
type Session struct {
  Sid string `json:"SessionId"`
  timeAccessed time.Time
  userData Data
  lock sync.Mutex
  copyFormField map[string]bool // map[fieldId]true
  pasteFormField map[string]bool // map[fieldId]true
}
```

A session is protected by a mutex so that concurrent requests by the same user do not interfere, similar to above.

## dataHandler

The handler `dataHandler` handles post requests as detailed in the spec. If it receives a request with GET status,
it serves up the home page.

 - If a request is made without a session id, 401 is returned
 - If data not in correct format, 400 is returned
 - If json marshalling fails, 500 is returned
 - 201 is returned on creating a session
 - 200 is returned on a post request with more information for a user's `Data` struct (e.g. upon a resize)

If a request is unsuccessful, the status code is set as per the above list and the response body includes a plaintext
error message. If a request is successful, a json of the user's struct so far is returned.

## Sessions

Clients make a post request to the path `/session` to ask for a new session id. They must include a json with field
`websiteUrl` in the request body.
In all future requests, they must include the field `"sessionId":string` else a 401 is returned.

## Frontend

Detecting resizes is done by defining an event listener and assigning it to `window.resize`.
Detecting copy/pastes are done by defining methods for the events `oncopy` and `onpaste` for each of the form fields.

Calculating time taken is done as follows. An event listener for 'onkeydown' is defined and attached to each form field.
Only if its the first time the event listener is called, then a global `keydown.time_start` is recorded.
On pressing form submit, an event listener is triggered that records 'keydown.time_end', calculates the time taken
and makes the suitable post request.

I do not use any third party js libraries.


## Notes

This is my first time using Go. I've definitely enjoyed and will try to do a future project using it.
Being my first time, it's quite possible there are patterns I've used which are considered bad practice in Go.
However, I've tried to avoid that.
I used the Tour of Go, the go docs and some of this [book](https://astaxie.gitbooks.io/build-web-application-with-golang/)
to learn it.

I didn't have time to implement some really fancy stuff like (good) input validation, tokens (to avoid cross site
request forgery) and some other stuff.

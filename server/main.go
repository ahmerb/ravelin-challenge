package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "io"
  "errors"
  "log"
  "net/http"
  "encoding/base64"
  "crypto/rand"
  "time"
  "sync"
)

// the global session manager
var globalSessions *SessionManager

// launch the server
func main() {
  // init session manager
  globalSessions = NewSessionManager("ravelin-test", 0)

  // handle requests from the form
  http.HandleFunc("/form", formHandler)

  // static files
  http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./client"))))

  // start a session
  http.HandleFunc("/session", newSessionHandler)

  // handler for (all other) requests
  http.HandleFunc("/", dataHandler)

  fmt.Println("Server now running on localhost:8080")
  fmt.Println(`Try running: curl -X POST -d '{"SessionId":"test123"}' http://localhost:8080/`)
  log.Fatal(http.ListenAndServe(":8080", nil))
}


func newSessionHandler(w http.ResponseWriter, r *http.Request) {
  // create a new session
  session := globalSessions.NewSession()

  // return the session id to the client
  response, err := json.Marshal(session)

  // if failed to marshal json, panic
  if err != nil {
    panic(err)
  }

  // set headers and construct response body
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusCreated)
  w.Write(response)
}


func formHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method == "POST" {

    // parse the form and log a msg
    r.ParseForm()
    log.Printf("received a form")

  // invalid HTTP verb
  } else {
    invalidVerb(w, r)
  }
}


func dataHandler(w http.ResponseWriter, r *http.Request) {

  // get request - serve home page
  // post request - parse and print Data json
  // other - return error status

  if r.Method == "GET" {

    // serve up the home page
    http.ServeFile(w, r, "./client/index.html")
    return

  } else if r.Method == "POST" {

    // read the body from the http request
    body, err := ioutil.ReadAll(r.Body)

    // return error status + msg if read fails
    if err != nil {
      w.WriteHeader(http.StatusBadRequest)
      w.Write([]byte("Unable to read body"))
      return
    }

    // unmarshal the request payload
    var data Data
    err = json.Unmarshal(body, &data)

    // if unmarshal fails then return error status + msg
    if err != nil {
      w.WriteHeader(http.StatusBadRequest)
      w.Write([]byte("Unable to unmarshal JSON request"))
      log.Printf(fmt.Sprintf("%v", err))
      return
    }

    // *** print the body ***
    log.Printf("%v\n\n", string(body))

    // retrieve session and update userData using request body
    var updatedData Data
    if data.SessionId != "" {

      // ask the session manager for the session
      var session Session
      session, err = globalSessions.LoadSession(data.SessionId)

      // if loading session failed, create error response
      if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte(err.Error()))
        return
      }

      // else, update this user's userData
      updatedData = updateUser(session, data)
    }

    // if all okay, return ok status, set headers and put updated userData in body
    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    response, err := json.Marshal(updatedData)

    // if marshalling json fails, panic
    if err != nil {
      panic(err)
    }

    // finally, write response
    w.Write(response)

  } else {
    invalidVerb(w, r)
  }
}

func updateUser(session Session, newData Data) Data {
  // lock the session, it can only be modified at once
  session.lock.Lock()
  defer session.lock.Unlock()

  // update session.userData
  currentData := session.userData
  updatedData := updateData(currentData, newData)

  return updatedData
}

func updateData(old, new Data) Data {
  // TODO: this is all too simplistic wrt invalid requests
  updated := old
  if new.WebsiteUrl != "" {
    updated.WebsiteUrl = new.WebsiteUrl
  }
  if (new.ResizeFrom != Dimension{}) {
    updated.ResizeFrom = new.ResizeFrom
  }
  if (new.ResizeTo != Dimension{}) {
    updated.ResizeTo = new.ResizeTo
  }
  if new.CopyAndPaste != nil {
    for k, v := range new.CopyAndPaste {
      updated.CopyAndPaste[k] = v
    }
  }
  return updated
}

func invalidVerb(w http.ResponseWriter, r *http.Request) {
  // write a 404 error if we don't support a request with this verb
  w.WriteHeader(http.StatusNotFound)
  w.Write([]byte(fmt.Sprintf("The HTTP verb %s is not supported\n", r.Method)))
}



// A session
type Session struct {
  Sid string `json:"SessionId"`
  timeAccessed time.Time
  userData Data
  lock sync.Mutex
}

// The data type we wish to maintain for each user
type Data struct {
  WebsiteUrl         string
  SessionId          string
  ResizeFrom         Dimension
  ResizeTo           Dimension
  CopyAndPaste       map[string]bool // map[fieldId]true
  FormCompletionTime int // Seconds
}

type Dimension struct {
  Width  string
  Height string
}


// The type for a SessionManager
type SessionManager struct {
  name        string
  lock        sync.Mutex                   // protects session from races
  maxLifeTime int64                        // expiry of a session
  sessions    map[string]Session           // map of sid's to sessions
}


func (manager *SessionManager) NewSession() (Session) {
  // races are bad
  manager.lock.Lock()
  defer manager.lock.Unlock()

  // create a new session id
  sid := manager.sessionId()

  // create new session object
  session := Session{Sid: sid, timeAccessed: time.Now(), userData: Data{SessionId: sid}}

  // register in sessions map
  manager.sessions[sid] = session

  return session
}


func (manager *SessionManager) LoadSession(sid string) (Session, error) {
  // races are bad
  manager.lock.Lock()
  defer manager.lock.Unlock()

  // attempt to retrieve the session
  session, exists := manager.sessions[sid]

  if !exists {
    return Session{}, errors.New("Sessions: Session Not Found")
  }

  // update access time
  session.timeAccessed = time.Now()

  return session, nil
}


func (manager *SessionManager) sessionId() string {
  // return an encoded string of 32 random digits, or "" if that fails
  b := make([]byte, 32)
  if _, err := io.ReadFull(rand.Reader, b); err != nil {
    return ""
  }
  return base64.URLEncoding.EncodeToString(b)
}


func NewSessionManager(name string, maxLifeTime int64) (*SessionManager) {
  // create a new session manager, giving it a sessions map
  sessions := make(map[string]Session)
  return &SessionManager{sessions: sessions, name: name, maxLifeTime: maxLifeTime}
}

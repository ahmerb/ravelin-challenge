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
  session := globalSessions.NewSession()

  // return the session id to the client
  response, err := json.Marshal(session)
  log.Printf(string(response))
  if err != nil {
    panic(err)
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusCreated)
  w.Write(response)
}


func formHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method == "POST" {
    r.ParseForm()
    log.Printf("%v\n", r.Form["inputEmail"])
    log.Printf("path", r.URL.Path)
    log.Printf("scheme", r.URL.Scheme)

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
      return
    }

    // print the body
    log.Printf("%v\n", string(body))

    // get session details and update data for user
    var updatedData Data
    if data.SessionId != "" {
      var session Session
      session, err = globalSessions.LoadSession(data.SessionId)
      if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte(err.Error()))
        return
      }
      updatedData = updateUser(session, data)
    }

    // return ok status and json of updatedData
    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    response, err := json.Marshal(updatedData)
    if err != nil {
      panic(err)
    }
    w.Write(response)

  } else {
    invalidVerb(w, r)
  }
}

func updateUser(session Session, newData Data) Data {
  currentData := session.userData
  updatedData := updateData(currentData, newData)
  return updatedData
}

func updateData(old, new Data) Data {
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
  w.WriteHeader(http.StatusNotFound)
  w.Write([]byte(fmt.Sprintf("The HTTP verb %s is not supported\n", r.Method)))
}



// A session
type Session struct {
  Sid string `json:"SessionId"`
  timeAccessed time.Time
  userData Data
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
  session, present := manager.sessions[sid]

  if !present {
    return Session{}, errors.New("Sessions: Session Not Found")
  }

  // update access time
  session.timeAccessed = time.Now()

  return session, nil
}


func (manager *SessionManager) sessionId() string {
  b := make([]byte, 32)
  if _, err := io.ReadFull(rand.Reader, b); err != nil {
    return ""
  }
  return base64.URLEncoding.EncodeToString(b)
}


func NewSessionManager(name string, maxLifeTime int64) (*SessionManager) {
  sessions := make(map[string]Session)
  return &SessionManager{sessions: sessions, name: name, maxLifeTime: maxLifeTime}
}

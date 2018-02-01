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
  "strconv"
)

// The global session manager
var globalSessions *SessionManager

// Launch the server
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


// Request Handlers

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

    // unmarshall the json
    var data interface{}
    err = json.Unmarshal(body, data)

    // return error status + msg if read/unmarshall fails
    if err != nil {
      w.WriteHeader(http.StatusBadRequest)
      w.Write([]byte("Unable to read request body"))
      return
    }

    // process this request body, returning the response body
    response := processPostReq(data.(map[string]interface{}), w, r) // also sets status

    // return the response
    w.Header().Set("Content-Type", "application/json")
    w.Write(response)

  } else {
    invalidVerb(w, r)
  }
}

func processPostReq(data map[string]interface{}, w http.ResponseWriter, r *http.Request) []byte {
  // 1. get the event type from the json
  eventType, exists := data["eventType"];
  if !exists || eventType == "" {

    // if the json doesn't have this field, return a 400
    w.WriteHeader(http.StatusBadRequest)
    return []byte("Request body must include an `eventType` field")
  }

  // 2. get the session id
  sid, exists := data["sessionId"]
  if !exists || eventType == "" {

    // if the json doesn't have this field, return a 400
    w.WriteHeader(http.StatusBadRequest)
    return []byte("Request must include a sessionId")
  }

  // retrieve this users session
  session, err := globalSessions.LoadSession(sid.(string))
  if err != nil {

    // if retrieving this session fails, return a 401
    w.WriteHeader(http.StatusUnauthorized)
    return []byte(err.Error())
  }

  // 3. process data based on given eventType
  session.lock.Lock()
  defer session.lock.Unlock()
  switch eventType {
  case "copyAndPaste":
    return processCopyAndPaste(data, session, w, r)
  case "resizeWindow":
    return processResizeWindow(data, session, w, r)
  case "timeTaken":
    return processTimeTaken(data, session, w, r)
  default:
    w.WriteHeader(http.StatusBadRequest)
    return []byte("Value of eventType not recognised")
  }
}


func processCopyAndPaste(data map[string]interface{}, session Session, w http.ResponseWriter, r *http.Request) []byte {
  // get the (string) pasted field from request json
  pasted, exists0 := data["pasted"]

  // get the (string) formId field from request json
  formId, exists1 := data["formId"]

  // convert field to boolean
  isPasted, err := strconv.ParseBool(pasted.(string))

  // return 401 if any operation failed
  if (err != nil || !exists0 || !exists1) {
    w.WriteHeader(http.StatusBadRequest)
    return []byte("copyAndPaste request must also include `pasted` boolean field and `formId` string field")
  }

  // update session copy- or pasteFormField maps
  formIdStr := formId.(string)
  if isPasted {
    session.pasteFormField[formIdStr] = true
  } else {
    session.copyFormField[formIdStr]  = true
  }

  // if both are true, then update the userData too
  if session.pasteFormField[formIdStr] && session.copyFormField[formIdStr] {
    session.userData.CopyAndPaste[formIdStr] = true
  }

  // return ok response with userData
  return jsonMarshal(session.userData, w, r)
}


func processTimeTaken(data map[string]interface{}, session Session, w http.ResponseWriter, r *http.Request) []byte {
  return []byte("")
}

func processResizeWindow(data map[string]interface{}, session Session, w http.ResponseWriter, r *http.Request) []byte {
  return []byte("")
}


func jsonMarshal(data Data, w http.ResponseWriter, r *http.Request) []byte {
  response, err := json.Marshal(data)

  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    return []byte("Internal Server Error: json marshalling failed")
  }

  w.WriteHeader(http.StatusOK)
  return response
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
  copyFormField map[string]bool // map[fieldId]true
  pasteFormField map[string]bool // map[fieldId]true
}

// The data type we wish to maintain for each user
type Data struct {
  WebsiteUrl         string
  SessionId          string
  ResizeFrom         Dimension
  ResizeTo           Dimension
  CopyAndPaste       map[string]bool // map[fieldId]true
  FormCompletionTime int             // Seconds
}

type Dimension struct {
  Width  string
  Height string
}


// The struct and methods for a SessionManager
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
  session := Session{
    Sid: sid,
    timeAccessed: time.Now(),
    userData: Data{SessionId: sid},
    copyFormField: make(map[string]bool),
    pasteFormField: make(map[string]bool) }

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

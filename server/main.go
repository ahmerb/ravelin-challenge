package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
)

// launch the server
func main() {
  // handle requests from the form
  http.HandleFunc("/form", formHandler)

  // static files
  http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./client"))))

  // handler for (all other) post requests
  http.HandleFunc("/", rootHandler)

  fmt.Println("Server now running on localhost:8080")
  fmt.Println(`Try running: curl -X POST -d '{"SessionId":"test123"}' http://localhost:8080/`)
  log.Fatal(http.ListenAndServe(":8080", nil))
}

// define the data type we wish to receive
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


func rootHandler(w http.ResponseWriter, r *http.Request) {

  // get request - serve home page
  // post request - parse and print Data json
  // other - return 404 error
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

    // return ok status
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Successful request."))

  } else {
    invalidVerb(w, r)
  }
}


func invalidVerb(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusNotFound)
  w.Write([]byte(fmt.Sprintf("The HTTP verb %s is not supported\n", r.Method)))
}

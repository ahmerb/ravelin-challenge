package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
)

func main() {
  http.HandleFunc("/helloworld", helloWorldHandler)

  fmt.Println("Server now running on localhost:8080")
  fmt.Println(`Try running: curl -X POST -d '{"hello":"test123"}' http://localhost:8080/helloworld`)
  log.Fatal(http.ListenAndServe(":8080", nil))
}

type helloWorldRequest struct {
  Hello string `json:"hello"`
}

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
      w.WriteHeader(http.StatusBadRequest)
      w.Write([]byte("Unable to read body"))
      return
  }

  req := &helloWorldRequest{}

  if err = json.Unmarshal(body, req); err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte("Unable to unmarshal JSON request"))
    return
  }

  log.Printf("Request received %+v", req)

  w.WriteHeader(http.StatusOK)
}

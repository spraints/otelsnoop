package main

import (
  "fmt"
  "sync"
  "net/http"
  "net/http/httputil"
)

func main() {
  var d dump
  s := &http.Server{
    Addr: "127.0.0.1:8360",
    Handler: &d,
  }
  s.ListenAndServe()
}

type dump struct{
  lock sync.Mutex
}

func (d *dump) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  s, err := httputil.DumpRequest(r, true)
  if err != nil {
    http.Error(w, err.Error(), 500)
    return
  }

  d.lock.Lock()
  fmt.Printf("%s\n", s)
  d.lock.Unlock()
}

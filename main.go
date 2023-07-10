package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"sync"

	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

func main() {
	var d dump
	s := &http.Server{
		Addr:    "127.0.0.1:8360",
		Handler: &d,
	}
	s.ListenAndServe()
}

type dump struct {
	lock sync.Mutex
}

func (d *dump) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s, err := httputil.DumpRequest(r, false)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	body, err := io.ReadAll(getBodyReader(r))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	m := &bytes.Buffer{}
	if r.Method == "POST" && r.URL.Path == "/traces/otlp/v0.9" && r.Header.Get("Content-Type") == "application/x-protobuf" {
		t := &tracepb.TracesData{}
		if err := proto.Unmarshal(body, t); err != nil {
			fmt.Fprintf(m, "error parsing body (%d bytes): %v", len(body), err)
		} else {
			for _, s := range t.GetResourceSpans() {
				fmt.Fprintln(m, s)
			}
		}
	} else {
		fmt.Fprintln(m, "unrecognized request")
	}

	d.lock.Lock()
	fmt.Printf("%s\n%s\n", s, m.Bytes())
	d.lock.Unlock()
}

func getBodyReader(r *http.Request) io.Reader {
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		r, _ := gzip.NewReader(r.Body)
		return r
	default:
		return r.Body
	}
}

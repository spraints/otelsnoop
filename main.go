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
	m := &bytes.Buffer{}
	if r.Method == "POST" && r.URL.Path == "/traces/otlp/v0.9" && r.Header.Get("Content-Type") == "application/x-protobuf" {
	body, err := io.ReadAll(getBodyReader(r))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

		t := &tracepb.TracesData{}
		if err := proto.Unmarshal(body, t); err != nil {
			fmt.Fprintf(m, "error parsing body (%d bytes): %v", len(body), err)
		} else {
			for _, rs := range t.GetResourceSpans() {
				fmt.Fprintf(m, "%v\n", rs.GetResource())
				for _, ss := range rs.GetScopeSpans() {
					fmt.Fprintf(m, "- %v\n", ss.GetScope())
					for _, s := range ss.GetSpans() {
						//   + trace_id:"!v\xcay[\x17\x9a\xe3\xd7\x14\xf8\x19\x8a\xf4g\x05"  span_id:"^ÑT;˹\x8e"  name:"HTTP POST"  kind:SPAN_KIND_CLIENT  start_time_unix_nano:1689026248178653295  end_time_unix_nano:1689026253188301775  attributes:{key:"http.method"  value:{string_value:"POST"}}  attributes:{key:"http.url"  value:{string_value:"http://localhost:18081/twirp/aqueduct.api.v1.JobQueueService/Receive"}}  attributes:{key:"net.peer.name"  value:{string_value:"localhost"}}  attributes:{key:"http.status_code"  value:{int_value:200}}  status:{}
						fmt.Fprintf(m, "  [%d %v] %q\n", s.GetStartTimeUnixNano(), s.GetKind(), s.GetName())
						for _, a := range s.GetAttributes() {
							fmt.Fprintf(m, "  + %v\n", a)
						}
					}
				}
			}
		}
	} else {
		fmt.Fprintln(m, "unrecognized request")
          s, err := httputil.DumpRequest(r, false)
          if err != nil {
                  http.Error(w, err.Error(), 500)
                  return
          }
          fmt.Fprintf(m, "%s\n", s)
	}

	d.lock.Lock()
	fmt.Printf("%s", m.Bytes())
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

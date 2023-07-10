package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"sync"

	"google.golang.org/protobuf/encoding/protowire"
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var m bytes.Buffer
	if r.Header.Get("Content-Type") == "application/x-protobuf" {
		if err := decodeProto(body, &m, ""); err != nil {
			fmt.Printf("%s\nbody:\n%s\n", s, body)
			http.Error(w, err.Error(), 500)
			return
		}
	}

	d.lock.Lock()
	fmt.Printf("%s\n%s\n", s, m.Bytes())
	d.lock.Unlock()
}

func decodeProto(data []byte, w io.Writer, indent string) error {
	remaining := data
	for len(remaining) > 0 {
		fieldNum, wireType, n := protowire.ConsumeTag(remaining)
		if n < 0 {
			return fmt.Errorf("failed to consume tag: %w", protowire.ParseError(n))
		}
		remaining = remaining[n:]

		fmt.Fprintf(w, "fieldNum: %v %T\nwireType: %v\n", fieldNum, fieldNum, wireType)
		switch wireType {
		case protowire.VarintType:
			fmt.Fprintln(w, "varint")
		case protowire.Fixed32Type:
			fmt.Fprintln(w, "fixed32")
		case protowire.Fixed64Type:
			fmt.Fprintln(w, "fixed64")
		case protowire.BytesType:
			fmt.Fprintln(w, "bytes")
		case protowire.StartGroupType:
			fmt.Fprintln(w, "start group")
		case protowire.EndGroupType:
			fmt.Fprintln(w, "end group")
		default:
			//return fmt.Errorf("invalid wire type")
		}
		break
	}
	return nil
}

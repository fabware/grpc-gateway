package runtime_test

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/fabware/grpc-gateway/runtime"
)

func TestMarshalerForRequest(t *testing.T) {
	r, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf(`http.NewRequest("GET", "http://example.com", nil) failed with %v; want success`, err)
	}
	r.Header.Set("Content-Type", "application/x-example")

	mux := runtime.NewServeMux()

	in, out := runtime.MarshalerForRequest(mux, r)
	if _, ok := in.(*runtime.JSONPb); !ok {
		t.Errorf("in = %#v; want a runtime.JSONPb", in)
	}
	if _, ok := out.(*runtime.JSONPb); !ok {
		t.Errorf("out = %#v; want a runtime.JSONPb", in)
	}

	var marshalers [6]dummyMarshaler
	specs := []struct {
		opt runtime.ServeMuxOption

		wantIn  runtime.Marshaler
		wantOut runtime.Marshaler
	}{
		{
			opt:     runtime.WithMarshalerOption("*", &marshalers[0], &marshalers[0]),
			wantIn:  &marshalers[0],
			wantOut: &marshalers[0],
		},
		{
			opt:     runtime.WithInboundMarshalerOption("*", &marshalers[1]),
			wantIn:  &marshalers[1],
			wantOut: &marshalers[0],
		},
		{
			opt:     runtime.WithOutboundMarshalerOption("application/x-example", &marshalers[2]),
			wantIn:  &marshalers[1],
			wantOut: &marshalers[2],
		},
		{
			opt:     runtime.WithInboundMarshalerOption("application/x-example", &marshalers[3]),
			wantIn:  &marshalers[3],
			wantOut: &marshalers[2],
		},
		{
			opt:     runtime.WithMarshalerOption("application/x-example", &marshalers[4], &marshalers[5]),
			wantIn:  &marshalers[4],
			wantOut: &marshalers[5],
		},
	}
	for i, spec := range specs {
		var opts []runtime.ServeMuxOption
		for _, s := range specs[:i+1] {
			opts = append(opts, s.opt)
		}
		mux = runtime.NewServeMux(opts...)

		in, out = runtime.MarshalerForRequest(mux, r)
		if got, want := in, spec.wantIn; got != want {
			t.Errorf("in = %#v; want %#v", got, want)
		}
		if got, want := out, spec.wantOut; got != want {
			t.Errorf("out = %#v; want %#v", got, want)
		}
	}

	r.Header.Set("Content-Type", "application/x-another")
	in, out = runtime.MarshalerForRequest(mux, r)
	if got, want := in, &marshalers[1]; got != want {
		t.Errorf("in = %#v; want %#v", got, want)
	}
	if got, want := out, &marshalers[0]; got != want {
		t.Errorf("out = %#v; want %#v", got, want)
	}
}

type dummyMarshaler struct{}

func (dummyMarshaler) ContentType() string { return "" }
func (dummyMarshaler) Marshal(interface{}) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (dummyMarshaler) Unmarshal([]byte, interface{}) error {
	return errors.New("not implemented")
}

func (dummyMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return dummyDecoder{}
}
func (dummyMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return dummyEncoder{}
}

type dummyDecoder struct{}

func (dummyDecoder) Decode(interface{}) error {
	return errors.New("not implemented")
}

type dummyEncoder struct{}

func (dummyEncoder) Encode(interface{}) error {
	return errors.New("not implemented")
}

// Copyright (c) 2018, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the LICENSE.md file
// distributed with the sources of this project regarding your rights to use or distribute this
// software.

package jsonresp

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestError(t *testing.T) {
	tests := []struct {
		name          string
		code          int
		message       string
		wantErrString string
	}{
		{"NoMessage", http.StatusNotFound, "", "404 Not Found"},
		{"Message", http.StatusNotFound, "blah", "blah (404 Not Found)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			je := NewError(tt.code, tt.message)
			if je.Code != tt.code {
				t.Errorf("got code %v, want %v", je.Code, tt.code)
			}
			if je.Message != tt.message {
				t.Errorf("got message %v, want %v", je.Message, tt.message)
			}
			if s := je.Error(); s != tt.wantErrString {
				t.Errorf("got string %v, want %v", s, tt.wantErrString)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name        string
		error       string
		code        int
		wantMessage string
		wantCode    int
	}{
		{"NoMessage", "", http.StatusNotFound, "", http.StatusNotFound},
		{"NoMessage", "blah", http.StatusNotFound, "blah", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			if err := WriteError(rr, tt.error, tt.code); err != nil {
				t.Fatalf("failed to write error: %v", err)
			}

			if rr.Code != tt.wantCode {
				t.Errorf("got code %v, want %v", rr.Code, tt.wantCode)
			}

			var jr Response
			if err := json.NewDecoder(rr.Body).Decode(&jr); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if jr.Error == nil {
				t.Fatalf("nil error received")
			}
			if jr.Error.Message != tt.wantMessage {
				t.Errorf("got message %v, want %v", jr.Error.Message, tt.wantMessage)
			}
			if jr.Error.Code != tt.wantCode {
				t.Errorf("got code %v, want %v", jr.Error.Code, tt.wantCode)
			}
		})
	}
}

func TestWriteResponsePage(t *testing.T) {
	type TestStruct struct {
		Value string
	}

	tests := []struct {
		name      string
		data      interface{}
		pd        *PageDetails
		code      int
		wantValue string
		wantPD    *PageDetails
		wantCode  int
	}{
		{"Empty", TestStruct{""}, nil, http.StatusOK, "", nil, http.StatusOK},
		{"NotEmpty", TestStruct{"blah"}, nil, http.StatusOK, "blah", nil, http.StatusOK},
		{"PageNone", TestStruct{"blah"}, &PageDetails{}, http.StatusOK, "blah", &PageDetails{}, http.StatusOK},
		{"PagePrev", TestStruct{"blah"}, &PageDetails{Prev: "p"}, http.StatusOK, "blah", &PageDetails{Prev: "p"}, http.StatusOK},
		{"PageNext", TestStruct{"blah"}, &PageDetails{Next: "n"}, http.StatusOK, "blah", &PageDetails{Next: "n"}, http.StatusOK},
		{"PagePrevNext", TestStruct{"blah"}, &PageDetails{Prev: "p", Next: "n"}, http.StatusOK, "blah", &PageDetails{Prev: "p", Next: "n"}, http.StatusOK},
		{"PageSize", TestStruct{"blah"}, &PageDetails{TotalSize: 42}, http.StatusOK, "blah", &PageDetails{TotalSize: 42}, http.StatusOK},
		{"PagePrevSize", TestStruct{"blah"}, &PageDetails{Prev: "p", TotalSize: 42}, http.StatusOK, "blah", &PageDetails{Prev: "p", TotalSize: 42}, http.StatusOK},
		{"PageNextSize", TestStruct{"blah"}, &PageDetails{Next: "n", TotalSize: 42}, http.StatusOK, "blah", &PageDetails{Next: "n", TotalSize: 42}, http.StatusOK},
		{"PagePrevNextSize", TestStruct{"blah"}, &PageDetails{Prev: "p", Next: "n", TotalSize: 42}, http.StatusOK, "blah", &PageDetails{Prev: "p", Next: "n", TotalSize: 42}, http.StatusOK},
		{"Created", TestStruct{"blah"}, nil, http.StatusCreated, "blah", nil, http.StatusCreated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			if err := WriteResponsePage(rr, tt.data, tt.pd, tt.code); err != nil {
				t.Fatalf("failed to write response: %v", err)
			}

			var ts TestStruct
			pd, err := ReadResponsePage(rr.Body, &ts)
			if err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if ts.Value != tt.wantValue {
				t.Errorf("got value '%v', want '%v'", ts.Value, tt.wantValue)
			}
			if got, want := (pd == nil), (tt.wantPD == nil); got != want {
				t.Errorf("got nil page %v, want %v", got, want)
			} else if pd != nil {
				if got, want := pd, tt.wantPD; !reflect.DeepEqual(got, want) {
					t.Errorf("got page details %+v, want %+v", got, want)
				}
			}
			if rr.Code != tt.wantCode {
				t.Errorf("got code '%v', want '%v'", rr.Code, tt.wantCode)
			}
		})
	}
}

func TestWriteResponse(t *testing.T) {
	type TestStruct struct {
		Value string
	}

	tests := []struct {
		name      string
		data      interface{}
		code      int
		wantValue string
		wantCode  int
	}{
		{"Empty", TestStruct{""}, http.StatusOK, "", http.StatusOK},
		{"NotEmpty", TestStruct{"blah"}, http.StatusOK, "blah", http.StatusOK},
		{"Created", TestStruct{"blah"}, http.StatusCreated, "blah", http.StatusCreated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			if err := WriteResponse(rr, tt.data, tt.code); err != nil {
				t.Fatalf("failed to write response: %v", err)
			}

			var ts TestStruct
			if err := ReadResponse(rr.Body, &ts); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if ts.Value != tt.wantValue {
				t.Errorf("got value '%v', want '%v'", ts.Value, tt.wantValue)
			}
			if rr.Code != tt.wantCode {
				t.Errorf("got code '%v', want '%v'", rr.Code, tt.wantCode)
			}
		})
	}
}

func getResponseBodyPage(v interface{}, p *PageDetails) io.Reader {
	rr := httptest.NewRecorder()
	if err := WriteResponsePage(rr, v, p, http.StatusOK); err != nil {
		log.Fatalf("failed to write response: %v", err)
	}
	return rr.Body
}

func getResponseBody(v interface{}) io.Reader {
	return getResponseBodyPage(v, nil)
}

func getErrorBody() io.Reader {
	rr := httptest.NewRecorder()
	if err := WriteError(rr, "blah", http.StatusNotFound); err != nil {
		log.Fatalf("failed to write error: %v", err)
	}
	return rr.Body
}

func TestReadResponsePage(t *testing.T) {
	type TestStruct struct {
		Value string
	}

	tests := []struct {
		name      string
		r         io.Reader
		wantErr   bool
		wantValue string
		wantPD    *PageDetails
	}{
		{"Empty", bytes.NewReader(nil), true, "", nil},
		{"Response", getResponseBody(TestStruct{"blah"}), false, "blah", nil},
		{"ResponsePageNone", getResponseBodyPage(TestStruct{"blah"}, &PageDetails{}), false, "blah", &PageDetails{}},
		{"ResponsePagePrev", getResponseBodyPage(TestStruct{"blah"}, &PageDetails{Prev: "prev"}), false, "blah", &PageDetails{Prev: "prev"}},
		{"ResponsePageNext", getResponseBodyPage(TestStruct{"blah"}, &PageDetails{Next: "next"}), false, "blah", &PageDetails{Next: "next"}},
		{"ResponsePagePrevNext", getResponseBodyPage(TestStruct{"blah"}, &PageDetails{Prev: "prev", Next: "next"}), false, "blah", &PageDetails{Prev: "prev", Next: "next"}},
		{"ResponsePagePrevSize", getResponseBodyPage(TestStruct{"blah"}, &PageDetails{Prev: "prev", TotalSize: 42}), false, "blah", &PageDetails{Prev: "prev", TotalSize: 42}},
		{"ResponsePageNextSize", getResponseBodyPage(TestStruct{"blah"}, &PageDetails{Next: "next", TotalSize: 42}), false, "blah", &PageDetails{Next: "next", TotalSize: 42}},
		{"ResponsePagePrevNextSize", getResponseBodyPage(TestStruct{"blah"}, &PageDetails{Prev: "prev", Next: "next", TotalSize: 42}), false, "blah", &PageDetails{Prev: "prev", Next: "next", TotalSize: 42}},
		{"Error", getErrorBody(), true, "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts TestStruct

			pd, err := ReadResponsePage(tt.r, &ts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadResponse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if ts.Value != tt.wantValue {
					t.Errorf("got value '%v', want '%v'", ts.Value, tt.wantValue)
				}
				if got, want := (pd == nil), (tt.wantPD == nil); got != want {
					t.Errorf("got nil page %v, want %v", got, want)
				} else if pd != nil {
					if got, want := pd, tt.wantPD; !reflect.DeepEqual(got, want) {
						t.Errorf("got page details %+v, want %+v", got, want)
					}
				}
			}
		})
	}
}

func TestReadResponse(t *testing.T) {
	type TestStruct struct {
		Value string
	}

	tests := []struct {
		name      string
		r         io.Reader
		wantErr   bool
		wantValue string
	}{
		{"Empty", bytes.NewReader(nil), true, ""},
		{"Response", getResponseBody(TestStruct{"blah"}), false, "blah"},
		{"Error", getErrorBody(), true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts TestStruct

			err := ReadResponse(tt.r, &ts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadResponse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if ts.Value != tt.wantValue {
					t.Errorf("got value '%v', want '%v'", ts.Value, tt.wantValue)
				}
			}
		})
	}
}

func TestReadResponseNil(t *testing.T) {
	type TestStruct struct {
		Value string
	}

	tests := []struct {
		name    string
		r       io.Reader
		wantErr bool
	}{
		{"Empty", bytes.NewReader(nil), true},
		{"Nil", getResponseBody(nil), false},
		{"Response", getResponseBody(TestStruct{"blah"}), false},
		{"Error", getErrorBody(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ReadResponse(tt.r, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReadError(t *testing.T) {
	type TestStruct struct {
		Value string
	}

	tests := []struct {
		name        string
		r           io.Reader
		wantErr     bool
		wantMessage string
		wantCode    int
	}{
		{"Empty", bytes.NewReader(nil), false, "", 0},
		{"Response", getResponseBody(TestStruct{"blah"}), false, "", 0},
		{"Error", getErrorBody(), true, "blah", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ReadError(tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadError() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				err, ok := err.(*Error)
				if !ok {
					t.Fatal("invalid error type")
				}
				if got, want := err.Message, tt.wantMessage; got != want {
					t.Errorf("got message %v, want %v", got, want)
				}
				if got, want := err.Code, tt.wantCode; got != want {
					t.Errorf("got code %v, want %v", got, want)
				}
			}
		})
	}
}

func TestSafeWriteResponse(t *testing.T) {

	type SampleData struct {
		SampleField string `json:"data"`
		Value       int    `json:"value"`
	}

	sampleData := SampleData{SampleField: "hello", Value: 123}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = SafeWriteResponse(w, sampleData)
	})

	// for this test, the URI, HTTP method (and payload, where applicable) is inconsequential.
	// the test should validate the response only
	r := httptest.NewRequest("GET", "http://example.com/endpoint", nil)
	w := httptest.NewRecorder()
	handler(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("HTTP status was incorrect, got: %d, expected %d", resp.StatusCode, http.StatusOK)
	}

	const expectedHeaderValue = "application/json"

	if resp.Header.Get("Content-Type") != expectedHeaderValue {
		t.Errorf("HTTP header Content-Type was incorrect, got: %s, expected %s", resp.Header.Get("Content-Type"), expectedHeaderValue)
	}

	// if jsonresp.ReadResponse() fails, it is an indication that jsonutil.SafeJSONWriteResponse
	// has failed to write properly formatted response data
	var result SampleData
	err := ReadResponse(resp.Body, &result)
	if err != nil {
		t.Errorf("Error reading JSON response: %v", err)
	}

	if result.SampleField != sampleData.SampleField || result.Value != sampleData.Value {
		t.Errorf("JSON unmarshaling error, got: %v, expected: %v", result, sampleData)
	}
}

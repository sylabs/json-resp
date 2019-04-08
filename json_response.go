// Copyright (c) 2018, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the LICENSE.md file
// distributed with the sources of this project regarding your rights to use or distribute this
// software.

package jsonresp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (e *Error) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%v (%v %v)", e.Message, e.Code, http.StatusText(e.Code))
	}
	return fmt.Sprintf("%v %v", e.Code, http.StatusText(e.Code))
}

// PageDetails specifies paging information.
type PageDetails struct {
	Prev      string `json:"prev,omitempty"`
	Next      string `json:"next,omitempty"`
	TotalSize int    `json:"totalSize,omitempty"`
}

// Error describes an error condition.
type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

var (
	// JSONErrorUnauthorized is a generic 401 unauthorized reponse
	JSONErrorUnauthorized = &Error{
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized",
	}
)

// Response is the top level container of all of our REST API responses.
type Response struct {
	Data  interface{}  `json:"data,omitempty"`
	Page  *PageDetails `json:"page,omitempty"`
	Error *Error       `json:"error,omitempty"`
}

// NewError returns an error that contains the given code and message.
func NewError(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// WriteError encodes the supplied error in a response, and writes to w.
func WriteError(w http.ResponseWriter, error string, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	jr := Response{
		Error: &Error{
			Code:    code,
			Message: error,
		},
	}
	if err := json.NewEncoder(w).Encode(jr); err != nil {
		return fmt.Errorf("jsonresp: failed to write error: %v", err)
	}
	return nil
}

// WriteResponsePage encodes the supplied data in a paged JSON response, and writes to w.
func WriteResponsePage(w http.ResponseWriter, data interface{}, pd *PageDetails, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	jr := Response{
		Data: data,
		Page: pd,
	}
	if err := json.NewEncoder(w).Encode(jr); err != nil {
		return fmt.Errorf("jsonresp: failed to write response: %v", err)
	}
	return nil
}

// ReadResponsePage reads a paged JSON response, and unmarshals the supplied data.
func ReadResponsePage(r io.Reader, v interface{}) (pd *PageDetails, err error) {
	var u struct {
		Data  json.RawMessage `json:"data"`
		Page  *PageDetails    `json:"page"`
		Error *Error          `json:"error"`
	}
	if err := json.NewDecoder(r).Decode(&u); err != nil {
		return nil, fmt.Errorf("jsonresp: failed to read response: %v", err)
	}
	if u.Error != nil {
		return nil, u.Error
	}
	if v != nil {
		if err := json.Unmarshal(u.Data, v); err != nil {
			return nil, fmt.Errorf("jsonresp: failed to unmarshal response: %v", err)
		}
	}
	return u.Page, nil
}

// WriteResponse encodes the supplied data in a response, and writes to w.
func WriteResponse(w http.ResponseWriter, data interface{}, code int) error {
	return WriteResponsePage(w, data, nil, code)
}

// SafeWriteResponse is wrapper around jsonresp.WriteResponse() that will
// write a HTTP status 500 error with the error message if the JSON encode
// fails.
func SafeWriteResponse(w http.ResponseWriter, payload interface{}) *Error {
	err := WriteResponse(w, payload, http.StatusOK)
	if err != nil {
		return NewError(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// ReadResponse reads a JSON response, and unmarshals the supplied data.
func ReadResponse(r io.Reader, v interface{}) error {
	_, err := ReadResponsePage(r, v)
	return err
}

// ReadError attempts to unmarshal JSON-encoded error details from the supplied reader. It returns
// nil if an error could not be parsed from the response, or if the parsed error was nil.
func ReadError(r io.Reader) error {
	var u struct {
		Error *Error `json:"error"`
	}
	if err := json.NewDecoder(r).Decode(&u); err != nil {
		return nil
	}
	if u.Error == nil {
		return nil
	}
	return u.Error
}

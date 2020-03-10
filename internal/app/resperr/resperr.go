/*
 * Copyright (c) 2020 Learning by Example maintainers.
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in
 *  all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 *  THE SOFTWARE.
 */

package resperr

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/LearningByExample/go-microservice/internal/app/constants"
	"net/http"
)

const (
	writtenJson      = "written json"
	invalidUrl       = "invalid url"
	notBodyProvided  = "not body provided"
	badRequest       = "bad request"
	resourceNotFound = "resource not found"
	invalidResource  = "invalid resource"
)

type ResponseError struct {
	error
	ErrorStr string `json:"error"`
	status   int
	Message  []string `json:"message,omitempty"`
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("ErrorStr: %q, status: %d, Message: %v", e.ErrorStr, e.status, e.Message)
}

func (e ResponseError) Status() int {
	return e.status
}

func (e ResponseError) Write(w http.ResponseWriter) {
	w.Header().Add(constants.ContentType, constants.ApplicationJsonUtf8)
	w.WriteHeader(e.status)
	encoder := json.NewEncoder(w)
	_ = encoder.Encode(e)
}

func newResponseError(e error, status int, msg []string) ResponseError {
	return ResponseError{
		error:    e,
		ErrorStr: e.Error(),
		status:   status,
		Message:  msg,
	}
}

func NewResErrForStr(str string, status int) ResponseError {
	return newResponseError(errors.New(str), status, make([]string, 0))
}

func FromError(err error) ResponseError {
	switch v := err.(type) {
	case ResponseError:
		return v
	default:
		return newResponseError(v, http.StatusInternalServerError, make([]string, 0))
	}
}

func FromErrorMessage(err ResponseError, msg []string) ResponseError {
	return newResponseError(err, err.Status(), msg)
}

var (
	WrittenJson     = NewResErrForStr(writtenJson, http.StatusInternalServerError)
	InvalidUrl      = NewResErrForStr(invalidUrl, http.StatusBadRequest)
	InvalidResource = NewResErrForStr(invalidResource, http.StatusUnprocessableEntity)
	NotBodyProvided = NewResErrForStr(notBodyProvided, http.StatusBadRequest)
	BadRequest      = NewResErrForStr(badRequest, http.StatusBadRequest)
	NotFound        = NewResErrForStr(resourceNotFound, http.StatusNotFound)
	None            = ResponseError{status: http.StatusOK}
)

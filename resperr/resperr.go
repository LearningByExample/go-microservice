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
	"github.com/LearningByExample/go-microservice/constants"
	"net/http"
)

const writtenJson = "written json"
const invalidUrl = "invalid url"
const notBodyProvided = "not body provided"
const invalidResource = "invalid resource"
const badRequest = "bad request"
const resourceNotFound = "resource not found"

type ResponseError struct {
	error
	ErrorStr string `json:"error"`
	status   int
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

func newResponseError(e error, status int) ResponseError {
	return ResponseError{
		error:    e,
		ErrorStr: e.Error(),
		status:   status,
	}
}

func newResErrForStr(str string, status int) ResponseError {
	return newResponseError(errors.New(str), status)
}

func FromError(err error) ResponseError {
	switch v := err.(type) {
	case ResponseError:
		return v
	default:
		return newResponseError(v, http.StatusInternalServerError)
	}
}

var (
	WrittenJson     = newResErrForStr(writtenJson, http.StatusInternalServerError)
	InvalidUrl      = newResErrForStr(invalidUrl, http.StatusBadRequest)
	InvalidResource = newResErrForStr(invalidResource, http.StatusUnprocessableEntity)
	NotBodyProvided = newResErrForStr(notBodyProvided, http.StatusBadRequest)
	BadRequest      = newResErrForStr(badRequest, http.StatusBadRequest)
	NotFound        = newResErrForStr(resourceNotFound, http.StatusNotFound)
	None            = ResponseError{status: http.StatusOK}
)

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

package _test

import (
	"encoding/json"
	"github.com/LearningByExample/go-microservice/internal/app/resperr"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func testRequest(handler http.Handler, url string, method string, i interface{}) *httptest.ResponseRecorder {
	var body io.Reader = nil
	if i != nil {
		switch v := i.(type) {
		case string:
			body = strings.NewReader(v)
			break
		default:
			bytes, _ := json.Marshal(v)
			body = strings.NewReader(string(bytes))
			break
		}
	}
	request, _ := http.NewRequest(method, url, body)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	return response
}

func PostRequest(handler http.Handler, url string, i interface{}) *httptest.ResponseRecorder {
	return testRequest(handler, url, http.MethodPost, i)
}

func GetRequest(handler http.Handler, url string) *httptest.ResponseRecorder {
	return testRequest(handler, url, http.MethodGet, nil)
}

func DeleteRequest(handler http.Handler, url string) *httptest.ResponseRecorder {
	return testRequest(handler, url, http.MethodDelete, nil)
}

func PatchRequest(handler http.Handler, url string, i interface{}) *httptest.ResponseRecorder {
	return testRequest(handler, url, http.MethodPatch, i)
}

func PutRequest(handler http.Handler, url string, i interface{}) *httptest.ResponseRecorder {
	return testRequest(handler, url, http.MethodPut, i)
}

func AssertResponseError(t *testing.T, response *httptest.ResponseRecorder, error resperr.ResponseError) {
	t.Helper()

	got := response.Code
	want := error.Status()

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	if error.Status() != resperr.None.Status() {
		decoder := json.NewDecoder(response.Body)
		gotErrorResponse := resperr.ResponseError{}

		err := decoder.Decode(&gotErrorResponse)

		if err != nil {
			t.Fatalf("got error, %v", err)
		}

		if gotErrorResponse.ErrorStr != error.ErrorStr {
			t.Fatalf("got %q, want %q", gotErrorResponse.ErrorStr, error.ErrorStr)
		}
		if len(gotErrorResponse.Message) != 0 {
			if !reflect.DeepEqual(gotErrorResponse.Message, error.Message) {
				t.Fatalf("got %v, want %v", gotErrorResponse.Message, error.Message)
			}
		}
	}
}

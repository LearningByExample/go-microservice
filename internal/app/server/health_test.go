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

package server

import (
	"github.com/LearningByExample/go-microservice/internal/_test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func assertHealthResponse(t *testing.T, response *httptest.ResponseRecorder, status int) {
	t.Helper()

	got := response.Code
	want := status

	if got != want {
		t.Fatalf("error in health response got %v, want %v", got, want)
	}
}

func Test_healthHandler(t *testing.T) {
	spyStore := _test.NewSpyStore()
	h := NewHealthHandler(&spyStore)

	t.Run("readiness should work", func(t *testing.T) {
		spyStore.Reset()
		spyStore.WhenIsReady(func() bool {
			return true
		})
		request := _test.GetRequest(h, readinessUrl)

		assertHealthResponse(t, request, http.StatusOK)
		got := spyStore.IsReadyWasCall
		want := true
		if got != want {
			t.Fatalf("error in health response got %t, want %t", got, want)
		}
	})

	t.Run("readiness should fail", func(t *testing.T) {
		spyStore.Reset()
		spyStore.WhenIsReady(func() bool {
			return false
		})
		request := _test.GetRequest(h, readinessUrl)

		assertHealthResponse(t, request, http.StatusInternalServerError)
		got := spyStore.IsReadyWasCall
		want := true
		if got != want {
			t.Fatalf("error in health response got %t, want %t", got, want)
		}
	})

	t.Run("liveness should work", func(t *testing.T) {
		request := _test.GetRequest(h, livenessUrl)

		assertHealthResponse(t, request, http.StatusOK)
	})

	t.Run("invalid url should fail", func(t *testing.T) {
		request := _test.GetRequest(h, "/invalid")

		assertHealthResponse(t, request, http.StatusNotFound)
	})

}

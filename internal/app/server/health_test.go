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
	"errors"
	"github.com/LearningByExample/go-microservice/internal/_test"
	"github.com/LearningByExample/go-microservice/internal/app/resperr"
	"testing"
)

var (
	mockError = errors.New("mock error")
)

func Test_healthHandler(t *testing.T) {
	spyStore := _test.NewSpyStore()
	h := NewHealthHandler(&spyStore)

	t.Run("readiness should work", func(t *testing.T) {
		spyStore.Reset()
		spyStore.WhenIsReady(func() error {
			return nil
		})
		request := _test.GetRequest(h, readinessUrl)

		_test.AssertResponseError(t, request, resperr.None)
		got := spyStore.IsReadyWasCall
		want := true
		if got != want {
			t.Fatalf("error in health response got %t, want %t", got, want)
		}
	})

	t.Run("readiness should fail", func(t *testing.T) {
		spyStore.Reset()
		spyStore.WhenIsReady(func() error {
			return mockError
		})
		request := _test.GetRequest(h, readinessUrl)

		_test.AssertResponseError(t, request, resperr.FromError(mockError))
		got := spyStore.IsReadyWasCall
		want := true
		if got != want {
			t.Fatalf("error in health response got %t, want %t", got, want)
		}
	})

	t.Run("liveness should work", func(t *testing.T) {
		request := _test.GetRequest(h, livenessUrl)

		_test.AssertResponseError(t, request, resperr.None)
	})

	t.Run("invalid url should fail", func(t *testing.T) {
		request := _test.GetRequest(h, "/invalid")

		_test.AssertResponseError(t, request, resperr.NotFound)
	})

}

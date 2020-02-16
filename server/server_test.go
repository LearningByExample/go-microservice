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
	"github.com/LearningByExample/go-microservice/store/memory"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer(t *testing.T) {
	store := memory.NewInMemoryPetStore()
	store.AddPet("pelusa", "dog", "happy")

	handler := NewServer(8080, store).(server)

	type testCase struct {
		name string
		path string
		want int
	}

	var cases = []testCase{
		{
			name: "must return not found",
			path: "/badurl",
			want: http.StatusNotFound,
		},
		{
			name: "must return ok",
			path: "/pet/1",
			want: http.StatusOK,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			got := response.Code

			if got != tt.want {
				t.Fatalf("error got %v, want %v", got, tt.want)
			}
		})
	}
}

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
	"github.com/LearningByExample/go-microservice/internal/app/config"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"math/rand"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func createServerRandomPort(st store.PetStore) *server {
	port := rand.Intn(7000-6000) + 6000
	cfg := config.CfgData{
		Server: config.ServerCfg{
			Port: port,
		},
		Store: config.StoreCfg{},
	}
	srv := NewServer(cfg, st).(*server)

	return srv
}

func TestServerRoutes(t *testing.T) {
	st := _test.NewSpyStore()
	srv := createServerRandomPort(&st)

	type testCase struct {
		name string
		path string
		want int
	}

	var cases = []testCase{
		{
			name: "must return not found",
			path: "/bad-url",
			want: http.StatusNotFound,
		},
		{
			name: "must return ok",
			path: "/pets/1",
			want: http.StatusOK,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			response := _test.GetRequest(srv, tt.path)
			got := response.Code

			if got != tt.want {
				t.Fatalf("error got %v, want %v", got, tt.want)
			}
		})
	}

}

func TestServerNoError(t *testing.T) {
	st := _test.NewSpyStore()
	srv := createServerRandomPort(&st)

	st.Reset()
	errs := make([]error, 0)
	go func() {
		errs = srv.Start()
	}()

	for srv.isListening() != true {
	}

	srv.quit()

	time.Sleep(100 * time.Millisecond)

	want := make([]error, 0)

	if reflect.DeepEqual(errs, want) != true {
		t.Fatalf("want %v errors, got %v", want, errs)
	}

	if !st.OpenWasCall {
		t.Fatal("open was not called")
	}

	if !st.CloseWasCall {
		t.Fatal("close was not called")
	}
}

func TestServerStoreError(t *testing.T) {
	st := _test.NewSpyStore()
	srv := createServerRandomPort(&st)

	oErr := errors.New("nasty open error")
	cErr := errors.New("nasty close error")

	st.Reset()
	st.WhenOpen(func() error {
		return oErr
	})
	st.WhenClose(func() error {
		return cErr
	})

	got := srv.Start()
	want := []error{
		oErr, cErr,
	}

	if reflect.DeepEqual(got, want) != true {
		t.Fatalf("want %v errors, got %v", want, got)
	}

	if !st.OpenWasCall {
		t.Fatal("open was not called")
	}

	if !st.CloseWasCall {
		t.Fatal("close was not called")
	}
}

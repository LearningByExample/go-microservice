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
	"encoding/json"
	"github.com/LearningByExample/go-microservice/data"
	"github.com/LearningByExample/go-microservice/store"
	"github.com/LearningByExample/go-microservice/store/memory"
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
		bytes, _ := json.Marshal(i)
		body = strings.NewReader(string(bytes))
	}
	request, _ := http.NewRequest(method, url, body)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	return response
}

func postRequest(handler http.Handler, url string, i interface{}) *httptest.ResponseRecorder {
	return testRequest(handler, url, http.MethodPost, i)
}

func getRequest(handler http.Handler, url string) *httptest.ResponseRecorder {
	return testRequest(handler, url, http.MethodGet, nil)
}

func deleteRequest(handler http.Handler, url string) *httptest.ResponseRecorder {
	return testRequest(handler, url, http.MethodDelete, nil)
}

type SpyStore struct {
	deleteWasCall bool
	deleteId      int
	deleteFunc    func(id int) error
}

func (s *SpyStore) reset() {
	s.deleteWasCall = false
	s.deleteId = 0
	s.deleteFunc = func(id int) error {
		return nil
	}
}

func (s SpyStore) AddPet(_ string, _ string, _ string) int {
	return 1
}

func (s SpyStore) GetPet(_ int) (data.Pet, error) {
	return data.Pet{}, nil
}

func (s *SpyStore) DeletePet(id int) error {
	s.deleteWasCall = true
	s.deleteId = id
	return s.deleteFunc(id)
}

func (s *SpyStore) whenDeletePet(deleteFunc func(id int) error) {
	s.deleteFunc = deleteFunc
}

func TestNewPetHandler(t *testing.T) {
	petStore := memory.NewInMemoryPetStore()
	got := NewPetHandler(petStore)

	if got == nil {
		t.Fatalf("want new handler, got nil")
	}
}

func TestPetId(t *testing.T) {
	petStore := memory.NewInMemoryPetStore()
	h := NewPetHandler(petStore).(petHandler)

	type testCase struct {
		name  string
		path  string
		want  int
		error bool
	}
	var cases = []testCase{
		{
			name:  "must found a pet",
			path:  "/pet/1",
			want:  1,
			error: false,
		},
		{
			name:  "found another",
			path:  "/pet/2",
			want:  2,
			error: false,
		},
		{
			name:  "must error",
			path:  "/bad-url",
			want:  0,
			error: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.petID(tt.path)

			if (err != nil) != tt.error {
				t.Fatalf("got error %v", err)
			}

			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPetRequest(t *testing.T) {
	petStore := memory.NewInMemoryPetStore()
	petStore.AddPet("Fluff", "dog", "happy")
	petStore.AddPet("Lion", "cat", "brave")

	handler := NewPetHandler(petStore)

	response := getRequest(handler, "/pet/2")

	got := response.Code
	want := http.StatusOK
	if got != want {
		t.Fatalf("error got %v, want %v", got, want)
	}

	gotHeader := response.Header().Get("Content-Type")
	wantHeader := "application/json; charset=utf-8"

	if gotHeader != wantHeader {
		t.Fatalf("error got %q, want %q", gotHeader, wantHeader)
	}

	wantPet := data.Pet{
		Id:   2,
		Name: "Lion",
		Race: "cat",
		Mod:  "brave",
	}

	decoder := json.NewDecoder(response.Body)
	gotPet := data.Pet{}
	err := decoder.Decode(&gotPet)

	if err != nil {
		t.Fatalf("got error, %v", err)
	}

	if reflect.DeepEqual(gotPet, wantPet) != true {
		t.Fatalf("got %v, want %v", gotPet, wantPet)
	}
}

func TestPetResponses(t *testing.T) {
	petStore := memory.NewInMemoryPetStore()
	petStore.AddPet("Fluff", "dog", "happy")
	petStore.AddPet("Lion", "cat", "sad")
	handler := NewPetHandler(petStore)

	type testCase struct {
		name string
		path string
		want int
	}
	var cases = []testCase{
		{
			name: "must found a pet",
			path: "/pet/1",
			want: http.StatusOK,
		},
		{
			name: "must found another",
			path: "/pet/2",
			want: http.StatusOK,
		},
		{
			name: "must not found another",
			path: "/pet/3",
			want: http.StatusNotFound,
		},
		{
			name: "must error",
			path: "/bad-url",
			want: http.StatusBadRequest,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			response := getRequest(handler, tt.path)

			got := response.Code

			if got != tt.want {
				t.Fatalf("got %v, want %v on case %v", got, tt.want, tt.name)
			}
		})
	}
}

func TestPetEmptyPost(t *testing.T) {
	petStore := memory.NewInMemoryPetStore()
	handler := NewPetHandler(petStore)

	response := postRequest(handler, "/pet", nil)

	got := response.Code
	want := http.StatusBadRequest

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPetInvalidMethod(t *testing.T) {
	petStore := memory.NewInMemoryPetStore()
	handler := NewPetHandler(petStore)

	response := testRequest(handler, "/pet/1", http.MethodPatch, nil)

	got := response.Code
	want := http.StatusBadRequest

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPetPostInvalidJson(t *testing.T) {
	petStore := memory.NewInMemoryPetStore()
	handler := NewPetHandler(petStore)

	body := strings.NewReader("{")
	request, _ := http.NewRequest(http.MethodPost, "/pet", body)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	got := response.Code
	want := http.StatusBadRequest

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestValidPet(t *testing.T) {
	handler := petHandler{}

	type TestCase struct {
		name string
		pet  data.Pet
		want bool
	}
	var cases = []TestCase{
		{
			name: "everything empty",
			pet: data.Pet{
				Id:   0,
				Name: "",
				Race: "",
				Mod:  "",
			},
			want: false,
		},
		{
			name: "name empty",
			pet: data.Pet{
				Id:   0,
				Name: "",
				Race: "aaa",
				Mod:  "aaa",
			},
			want: false,
		},
		{
			name: "race empty",
			pet: data.Pet{
				Id:   0,
				Name: "aaa",
				Race: "",
				Mod:  "aaa",
			},
			want: false,
		},
		{
			name: "mod empty",
			pet: data.Pet{
				Id:   0,
				Name: "aaa",
				Race: "aaa",
				Mod:  "",
			},
			want: false,
		},
		{
			name: "no empty",
			pet: data.Pet{
				Id:   0,
				Name: "aaa",
				Race: "aaa",
				Mod:  "aa",
			},
			want: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.validPet(tt.pet)
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPetPostValidJsonNoPet(t *testing.T) {
	petStore := memory.NewInMemoryPetStore()
	handler := NewPetHandler(petStore)

	body := strings.NewReader("{}")
	request, _ := http.NewRequest(http.MethodPost, "/pet", body)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	got := response.Code
	want := http.StatusBadRequest

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPetPost(t *testing.T) {
	petStore := memory.NewInMemoryPetStore()
	handler := NewPetHandler(petStore)

	postPet := data.Pet{
		Name: "Lion",
		Race: "cat",
		Mod:  "brave",
	}
	response := postRequest(handler, "/pet", postPet)

	got := response.Code
	want := http.StatusOK

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	wantPet := data.Pet{
		Id:   1,
		Name: postPet.Name,
		Race: postPet.Race,
		Mod:  postPet.Mod,
	}

	gotPet, _ := petStore.GetPet(1)

	if reflect.DeepEqual(wantPet, gotPet) != true {
		t.Fatalf("got %v, want %v", gotPet, wantPet)
	}

	gotLocation := response.Header().Get(location)
	wantLocation := "/pet/1"

	if gotLocation != wantLocation {
		t.Fatalf("got %v, want %v", gotLocation, wantLocation)
	}
}

func TestDeletePet(t *testing.T) {
	spyStore := SpyStore{}
	handler := NewPetHandler(&spyStore)

	t.Run("we could delete a existing pet", func(t *testing.T) {
		spyStore.reset()
		spyStore.whenDeletePet(func(id int) error {
			return nil
		})

		response := deleteRequest(handler, "/pet/2")

		got := response.Code
		want := http.StatusOK

		if got != want {
			t.Fatalf("got %v, want %v", got, want)
		}

		if spyStore.deleteWasCall != true {
			t.Fatalf("delete was not called")
		}

		gotId := spyStore.deleteId
		wantId := 2

		if gotId != wantId {
			t.Fatalf("we didn't delete the right pet, got %v, want %v", gotId, wantId)
		}
	})

	t.Run("we couldn't delete a non existing pet", func(t *testing.T) {
		spyStore.reset()
		spyStore.whenDeletePet(func(id int) error {
			return store.PetNotFound
		})

		response := deleteRequest(handler, "/pet/2")

		got := response.Code
		want := http.StatusNotFound

		if got != want {
			t.Fatalf("got %v, want %v", got, want)
		}

		if spyStore.deleteWasCall != true {
			t.Fatalf("delete was not called")
		}

		gotId := spyStore.deleteId
		wantId := 2

		if gotId != wantId {
			t.Fatalf("we didn't delete the right pet, got %v, want %v", gotId, wantId)
		}
	})
}

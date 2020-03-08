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
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/LearningByExample/go-microservice/internal/_test"
	"github.com/LearningByExample/go-microservice/internal/app/constants"
	"github.com/LearningByExample/go-microservice/internal/app/data"
	"github.com/LearningByExample/go-microservice/internal/app/resperr"
	"github.com/LearningByExample/go-microservice/internal/app/store"
)

func TestNewPetHandler(t *testing.T) {
	spyStore := _test.NewSpyStore()
	got := NewPetHandler(&spyStore)

	if got == nil {
		t.Fatalf("want new handler, got nil")
	}
}

func TestPetId(t *testing.T) {
	spyStore := _test.NewSpyStore()
	h := NewPetHandler(&spyStore).(petHandler)

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

func TestGetPetRequest(t *testing.T) {
	spyStore := _test.NewSpyStore()
	handler := NewPetHandler(&spyStore)

	mockPet := data.Pet{
		Id:   2,
		Name: "Filipino",
		Race: "dog",
		Mod:  "happy",
	}

	spyStore.WhenGetPet(func(id int) (data.Pet, error) {
		return mockPet, nil
	})
	response := _test.GetRequest(handler, "/pet/2")

	assertPetResponseEquals(t, response, mockPet)

	if spyStore.GetWasCall != true {
		t.Fatalf("get was not called")
	}

	gotId := spyStore.Id
	wantId := 2

	if gotId != wantId {
		t.Fatalf("we didn't get the right pet, got %v, want %v", gotId, wantId)
	}
}

func assertResponseError(t *testing.T, response *httptest.ResponseRecorder, error resperr.ResponseError) {
	t.Helper()

	got := response.Code
	want := error.Status()

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	if error != resperr.None {
		decoder := json.NewDecoder(response.Body)
		gotErrorResponse := resperr.ResponseError{}

		err := decoder.Decode(&gotErrorResponse)

		if err != nil {
			t.Fatalf("got error, %v", err)
		}

		if gotErrorResponse.ErrorStr != error.ErrorStr {
			t.Fatalf("got %q, want %q", gotErrorResponse.ErrorStr, error.ErrorStr)
		}
	}
}

func assertPetResponseEquals(t *testing.T, response *httptest.ResponseRecorder, pet data.Pet) {
	t.Helper()

	got := response.Code
	want := http.StatusOK

	if got != want {
		t.Fatalf("error got %v, want %v", got, want)
	}

	gotHeader := response.Header().Get(constants.ContentType)
	wantHeader := constants.ApplicationJsonUtf8

	if gotHeader != wantHeader {
		t.Fatalf("error got %q, want %q", gotHeader, wantHeader)
	}

	decoder := json.NewDecoder(response.Body)
	gotPet := data.Pet{}
	err := decoder.Decode(&gotPet)

	if err != nil {
		t.Fatalf("got error, %v", err)
	}

	if reflect.DeepEqual(gotPet, pet) != true {
		t.Fatalf("got %v, want %v", gotPet, pet)
	}
}

func TestPetResponses(t *testing.T) {
	spyStore := _test.NewSpyStore()
	handler := NewPetHandler(&spyStore)

	type testCase struct {
		name              string
		path              string
		want              resperr.ResponseError
		getShouldBeCalled bool
		id                int
	}

	mockPet := data.Pet{
		Id:   1,
		Name: "Filipino",
		Race: "dog",
		Mod:  "happy",
	}

	funcGet := func(id int) (data.Pet, error) {
		switch id {
		case 1:
			return mockPet, nil
		case 3:
			return data.Pet{}, store.PetNotFound
		}
		return data.Pet{}, nil
	}

	var cases = []testCase{
		{
			name:              "must found a pet",
			path:              "/pet/1",
			id:                1,
			want:              resperr.None,
			getShouldBeCalled: true,
		},
		{
			name:              "must not found another",
			path:              "/pet/3",
			id:                3,
			want:              resperr.NotFound,
			getShouldBeCalled: true,
		},
		{
			name:              "must error",
			path:              "/bad-url",
			id:                0,
			want:              resperr.InvalidUrl,
			getShouldBeCalled: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			spyStore.Reset()
			spyStore.WhenGetPet(funcGet)

			response := _test.GetRequest(handler, tt.path)
			assertResponseError(t, response, tt.want)

			if spyStore.GetWasCall != tt.getShouldBeCalled {
				t.Fatalf("get was not called")
			}

			if tt.getShouldBeCalled && spyStore.Id != tt.id {
				t.Fatalf("we didn't get the right pet, got %v, want %v", spyStore.Id, tt.id)
			}
		})
	}
}

func TestPetEmptyPost(t *testing.T) {
	spyStore := _test.NewSpyStore()
	handler := NewPetHandler(&spyStore)

	response := _test.PostRequest(handler, "/pet", nil)
	assertResponseError(t, response, resperr.NotBodyProvided)
}

func TestPetInvalidMethod(t *testing.T) {
	spyStore := _test.NewSpyStore()
	handler := NewPetHandler(&spyStore)

	response := _test.PatchRequest(handler, "/pet/1", nil)
	assertResponseError(t, response, resperr.BadRequest)
}

func TestPetPostInvalidJson(t *testing.T) {
	spyStore := _test.NewSpyStore()
	handler := NewPetHandler(&spyStore)

	response := _test.PostRequest(handler, "/pet", "{")
	assertResponseError(t, response, resperr.InvalidResource)
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
	spyStore := _test.NewSpyStore()
	handler := NewPetHandler(&spyStore)

	response := _test.PostRequest(handler, "/pet", "{}")
	assertResponseError(t, response, resperr.InvalidResource)
}

func TestPetPost(t *testing.T) {
	spyStore := _test.NewSpyStore()
	handler := NewPetHandler(&spyStore)

	postPet := data.Pet{
		Name: "Lion",
		Race: "cat",
		Mod:  "brave",
	}

	spyStore.WhenAddPet(func(name, race, mod string) int {
		return 5
	})
	response := _test.PostRequest(handler, "/pet", postPet)

	got := response.Code
	want := http.StatusOK

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	if spyStore.AddWasCall != true {
		t.Fatalf("add was not called")
	}

	wantPet := data.Pet{
		Id:   5,
		Name: postPet.Name,
		Race: postPet.Race,
		Mod:  postPet.Mod,
	}

	if reflect.DeepEqual(wantPet, spyStore.PetParameters) != true {
		t.Fatalf("got %v, want %v", spyStore.PetParameters, wantPet)
	}

	gotLocation := response.Header().Get(constants.Location)
	wantLocation := "/pet/5"

	if gotLocation != wantLocation {
		t.Fatalf("got %v, want %v", gotLocation, wantLocation)
	}
}

func TestDeletePet(t *testing.T) {
	spyStore := _test.NewSpyStore()
	handler := NewPetHandler(&spyStore)

	t.Run("we could delete a existing pet", func(t *testing.T) {
		spyStore.Reset()
		spyStore.WhenDeletePet(func(id int) error {
			return nil
		})

		response := _test.DeleteRequest(handler, "/pet/2")
		assertResponseError(t, response, resperr.None)

		if spyStore.DeleteWasCall != true {
			t.Fatalf("delete was not called")
		}

		gotId := spyStore.Id
		wantId := 2

		if gotId != wantId {
			t.Fatalf("we didn't delete the right pet, got %v, want %v", gotId, wantId)
		}
	})

	t.Run("we couldn't delete a non existing pet", func(t *testing.T) {
		spyStore.Reset()
		spyStore.WhenDeletePet(func(id int) error {
			return store.PetNotFound
		})

		response := _test.DeleteRequest(handler, "/pet/2")
		assertResponseError(t, response, resperr.NotFound)

		if spyStore.DeleteWasCall != true {
			t.Fatalf("delete was not called")
		}

		gotId := spyStore.Id
		wantId := 2

		if gotId != wantId {
			t.Fatalf("we didn't delete the right pet, got %v, want %v", gotId, wantId)
		}
	})
}

func TestPetPut(t *testing.T) {
	spyStore := _test.NewSpyStore()
	handler := NewPetHandler(&spyStore)

	putPet := data.Pet{
		Name: "Lion",
		Race: "cat",
		Mod:  "coward",
	}

	type Want struct {
		id          int
		status      int
		storeCalled bool
	}

	type TestCase struct {
		name   string
		url    string
		pet    data.Pet
		update bool
		err    error
		want   Want
	}

	var cases = []TestCase{
		{
			name: "modify a pet",
			url:  "/pet/1",
			pet: data.Pet{
				Name: "Lion",
				Race: "cat",
				Mod:  "coward",
			},
			update: true,
			err:    nil,
			want: Want{
				id:          1,
				status:      http.StatusOK,
				storeCalled: true,
			},
		},
		{
			name: "not modify a pet",
			url:  "/pet/1",
			pet: data.Pet{
				Name: "Lion",
				Race: "cat",
				Mod:  "coward",
			},
			update: false,
			err:    nil,
			want: Want{
				id:          1,
				status:      http.StatusNotModified,
				storeCalled: true,
			},
		},
		{
			name: "pet not found",
			url:  "/pet/1",
			pet: data.Pet{
				Name: "Lion",
				Race: "cat",
				Mod:  "coward",
			},
			update: true,
			err:    store.PetNotFound,
			want: Want{
				id:          1,
				status:      http.StatusNotFound,
				storeCalled: true,
			},
		},
		{
			name: "bad pet",
			url:  "/pet/1",
			pet: data.Pet{
				Name: "",
				Race: "cat",
				Mod:  "coward",
			},
			update: true,
			err:    nil,
			want: Want{
				id:          1,
				status:      http.StatusUnprocessableEntity,
				storeCalled: false,
			},
		},
		{
			name:   "bad url",
			url:    "/pet/zz",
			pet:    data.Pet{},
			update: true,
			err:    nil,
			want: Want{
				id:          1,
				status:      http.StatusBadRequest,
				storeCalled: false,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			spyStore.Reset()
			spyStore.WhenUpdatePet(func(id int, name string, race string, mod string) (b bool, err error) {
				return tt.update, tt.err
			})

			response := _test.PutRequest(handler, tt.url, tt.pet)

			got := response.Code

			if got != tt.want.status {
				t.Fatalf("got %v, want %v", got, tt.want.status)
			}

			if spyStore.UpdateWasCall != tt.want.storeCalled {
				t.Fatalf("want store call %v, got %v", tt.want.storeCalled, spyStore.UpdateWasCall)
			}

			if tt.want.storeCalled {
				gotId := spyStore.Id
				wantId := tt.want.id

				if gotId != wantId {
					t.Fatalf("we didn't update the right pet, got %v, want %v", gotId, wantId)
				}

				gotPet := spyStore.PetParameters
				if reflect.DeepEqual(gotPet, putPet) != true {
					t.Fatalf("got %v, want %v", gotPet, putPet)
				}
			}
		})
	}
}

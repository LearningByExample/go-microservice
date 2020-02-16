package server

import (
	"encoding/json"
	"github.com/LearningByExample/go-microservice/data"
	"github.com/LearningByExample/go-microservice/store/memory"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func testRequest(url string, method string, i interface{}) (*http.Request, *httptest.ResponseRecorder) {
	var body io.Reader = nil
	if i != nil {
		bytes, _ := json.Marshal(i)
		body = strings.NewReader(string(bytes))
	}
	request, _ := http.NewRequest(method, url, body)
	response := httptest.NewRecorder()

	return request, response
}

func postRequest(url string, i interface{}) (*http.Request, *httptest.ResponseRecorder) {
	return testRequest(url, http.MethodPost, i)
}

func getRequest(url string) (*http.Request, *httptest.ResponseRecorder) {
	return testRequest(url, http.MethodGet, nil)
}

func deleteRequest(url string) (*http.Request, *httptest.ResponseRecorder) {
	return testRequest(url, http.MethodDelete, nil)
}

type SpyStore struct {
	deleteWasCall bool
	deleteId      int
}

func NewSpyStore() SpyStore {
	store := SpyStore{
		deleteWasCall: false,
		deleteId:      0,
	}
	return store
}
func (s SpyStore) AddPet(name string, race string, mod string) int {
	return 1
}

func (s SpyStore) GetPet(id int) (data.Pet, error) {
	return data.Pet{}, nil
}

func (s SpyStore) DeletePet(id int) error {
	s.deleteWasCall = true
	s.deleteId = id
	return nil
}

func TestNewPetHandler(t *testing.T) {
	got := NewPetHandler(memory.NewInMemoryPetStore())

	if got == nil {
		t.Fatalf("want new handler, got nil")
	}
}

func TestPetId(t *testing.T) {
	h := NewPetHandler(memory.NewInMemoryPetStore()).(petHandler)

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
			path:  "/abcd",
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
	store := memory.NewInMemoryPetStore()
	store.AddPet("pelusa", "dog", "happy")
	store.AddPet("bola", "cat", "sad")

	handler := NewPetHandler(store)

	request, response := getRequest("/pet/2")

	handler.ServeHTTP(response, request)

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
		Name: "bola",
		Race: "cat",
		Mod:  "sad",
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
	store := memory.NewInMemoryPetStore()
	store.AddPet("pelusa", "dog", "happy")
	store.AddPet("bola", "cat", "sad")
	handler := NewPetHandler(store)

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
			path: "/abcd",
			want: http.StatusBadRequest,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			request, response := getRequest(tt.path)

			handler.ServeHTTP(response, request)

			got := response.Code

			if got != tt.want {
				t.Fatalf("got %v, want %v on case %v", got, tt.want, tt.name)
			}
		})
	}
}

func TestPetEmptyPost(t *testing.T) {
	store := memory.NewInMemoryPetStore()
	handler := NewPetHandler(store)

	request, response := postRequest("/pet", nil)

	handler.ServeHTTP(response, request)

	got := response.Code
	want := http.StatusBadRequest

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPetInvalidMethod(t *testing.T) {
	store := memory.NewInMemoryPetStore()
	handler := NewPetHandler(store)

	request, response := testRequest("/pet/1", http.MethodPatch, nil)

	handler.ServeHTTP(response, request)

	got := response.Code
	want := http.StatusBadRequest

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPetPostInvalidJson(t *testing.T) {
	store := memory.NewInMemoryPetStore()
	handler := NewPetHandler(store)

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
	store := memory.NewInMemoryPetStore()
	handler := NewPetHandler(store)

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
	store := memory.NewInMemoryPetStore()
	handler := NewPetHandler(store)

	postPet := data.Pet{
		Name: "leon",
		Race: "cat",
		Mod:  "brave",
	}
	request, response := postRequest("/pet", postPet)
	handler.ServeHTTP(response, request)

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

	gotPet, _ := store.GetPet(1)

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
	store := NewSpyStore()
	handler := NewPetHandler(store)

	request, response := deleteRequest("/pet/2")
	handler.ServeHTTP(response, request)

	got := response.Code
	want := http.StatusOK

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	if store.deleteWasCall == true {
		t.Fatalf("delete was not called")
	}

	if store.deleteId == 2 {
		t.Fatalf("we didn't delete the right pet")
	}
}

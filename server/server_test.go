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

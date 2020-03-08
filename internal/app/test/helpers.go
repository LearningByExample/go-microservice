package test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

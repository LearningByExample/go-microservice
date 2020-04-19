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
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"net/http"
)

type healthHandler struct {
	ps store.PetStore
}

const (
	readinessUrl = "/health/readiness"
	livenessUrl  = "/health/liveness"
)

func (h healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := http.StatusInternalServerError
	switch r.URL.Path {
	case livenessUrl:
		if h.isAlive() {
			status = http.StatusOK
		}
		break
	case readinessUrl:
		if h.isReady() {
			status = http.StatusOK
		}
		break
	default:
		status = http.StatusNotFound
	}

	w.WriteHeader(status)
}

func (h healthHandler) isAlive() bool {
	return true
}

func (h healthHandler) isReady() bool {
	_, err := h.ps.GetAllPets()
	return err == nil
}

func NewHealthHandler(ps store.PetStore) http.Handler {
	h := healthHandler{ps: ps}
	return &h
}

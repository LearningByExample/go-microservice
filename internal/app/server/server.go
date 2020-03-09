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
	"fmt"
	"github.com/LearningByExample/go-microservice/internal/app/resperr"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"log"
	"net/http"
)

const (
	rootPath     = "/"
	petPath      = "/pets"
	petWithSlash = "/pets/"
)

type Server interface {
	Serve() error
}

type server struct {
	port int
	mux  *http.ServeMux
}

func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s server) notFound(w http.ResponseWriter, _ *http.Request) {
	resperr.NotFound.Write(w)
}

func (s server) Serve() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting server at %s", addr)

	return http.ListenAndServe(addr, s)
}

func NewServer(port int, store store.PetStore) Server {
	mux := http.NewServeMux()

	srv := server{
		port: port,
		mux:  mux,
	}

	mux.HandleFunc(rootPath, srv.notFound)
	mux.Handle(petPath, NewPetHandler(store))
	mux.Handle(petWithSlash, NewPetHandler(store))

	return srv
}

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
	"context"
	"fmt"
	"github.com/LearningByExample/go-microservice/internal/app/resperr"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	rootPath     = "/"
	petPath      = "/pets"
	petWithSlash = "/pets/"
)

type Server interface {
	Start() []error
}

type server struct {
	h *http.Server
}

func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.h.Handler.ServeHTTP(w, r)
}

func (s server) notFound(w http.ResponseWriter, _ *http.Request) {
	resperr.NotFound.Write(w)
}

func (s server) Start() []error {
	errs := make([]error, 0)
	log.Printf("Starting server at %s ...", s.h.Addr)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.h.ListenAndServe(); err != nil {
			errs = append(errs, err)
			interrupt <- syscall.SIGIO
		}
	}()

	if len(errs) == 0 {
		log.Print("The service is ready to listen and serve.")

		killSignal := <-interrupt
		switch killSignal {
		case os.Interrupt:
			log.Print("Got SIGINT...")
		case syscall.SIGTERM:
			log.Print("Got SIGTERM...")
		case syscall.SIGIO:
			log.Print("Got SIGIO...")
		}

		log.Print("The service is shutting down...")
		err := s.h.Shutdown(context.Background())
		if err != nil {
			errs = append(errs, err)
		}
		log.Print("Server shutdown.")
	}

	return errs
}

func NewServer(port int, store store.PetStore) Server {
	mux := http.NewServeMux()

	addr := fmt.Sprintf(":%d", port)

	srv := server{
		h: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}

	petHandler := NewPetHandler(store)
	mux.HandleFunc(rootPath, srv.notFound)
	mux.Handle(petPath, petHandler)
	mux.Handle(petWithSlash, petHandler)

	return srv
}

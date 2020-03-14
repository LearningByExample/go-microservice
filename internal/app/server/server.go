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
	"github.com/LearningByExample/go-microservice/internal/app/config"
	"github.com/LearningByExample/go-microservice/internal/app/resperr"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
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
	hs  *http.Server
	ps  store.PetStore
	ch  chan os.Signal
	lnf int32
}

func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.hs.Handler.ServeHTTP(w, r)
}

func (s server) notFound(w http.ResponseWriter, _ *http.Request) {
	resperr.NotFound.Write(w)
}

const (
	quitSignal = syscall.SIGQUIT
)

func (s *server) isListening() bool {
	if atomic.LoadInt32(&(s.lnf)) != 0 {
		return true
	}
	return false
}

func (s *server) setListening(v bool) {
	var i int32 = 0
	if v {
		i = 1
	}
	atomic.StoreInt32(&(s.lnf), i)
}

func (s *server) quit() {
	s.ch <- quitSignal
}

func (s *server) Start() []error {
	log.Print("Starting server ...")
	errs := make([]error, 0)

	log.Print("Opening data store ...")
	if err := s.ps.Open(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		s.setListening(false)
		signal.Notify(s.ch, os.Interrupt, syscall.SIGTERM)

		log.Printf("Opening HTTP server at %s ...", s.hs.Addr)
		go func() {
			s.setListening(true)
			if err := s.hs.ListenAndServe(); err != nil {
				errs = append(errs, err)
				s.quit()
				s.setListening(false)
			}
		}()
		if len(errs) == 0 {
			log.Print("HTTP server listening ...")

			killSignal := <-s.ch
			switch killSignal {
			case os.Interrupt:
				log.Print("Got interrupt signal closing ...")
			case syscall.SIGTERM:
				log.Print("Got termination signal closing ...")
			case quitSignal:
				log.Print("Got quit signal closing ...")
			}

			if s.isListening() {
				log.Print("Closing HTTP server ...")
				err := s.hs.Shutdown(context.Background())
				if err != nil {
					errs = append(errs, err)
				}
				log.Print("HTTP server closed.")
				s.setListening(false)
			}

		}
	}

	log.Print("Closing data store ...")
	if err := s.ps.Close(); err != nil {
		errs = append(errs, err)
	}

	log.Print("Server stopped.")

	return errs
}

func NewServer(cfg config.CfgData, store store.PetStore) Server {
	mux := http.NewServeMux()

	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	srv := server{
		hs: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		ps: store,
		ch: make(chan os.Signal, 1),
	}

	srv.setListening(false)

	petHandler := NewPetHandler(srv.ps)
	mux.HandleFunc(rootPath, srv.notFound)
	mux.Handle(petPath, petHandler)
	mux.Handle(petWithSlash, petHandler)

	return &srv
}

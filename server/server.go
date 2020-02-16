package server

import (
	"fmt"
	"github.com/LearningByExample/go-microservice/store"
	"log"
	"net/http"
)

type Server interface {
	Serve()
}

type server struct {
	port int
	mux  *http.ServeMux
}

func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s server) notFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func (s server) Serve() {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting server at %s", addr)

	http.ListenAndServe(addr, s)
}

func NewServer(port int, store store.PetStore) Server {

	mux := http.NewServeMux()

	srv := server{
		port: port,
		mux:  mux,
	}

	mux.HandleFunc("/", srv.notFound)
	mux.Handle("/pet/", NewPetHandler(store))

	return srv
}

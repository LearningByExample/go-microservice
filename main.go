package main

import (
	"github.com/LearningByExample/go-microservice/server"
	"github.com/LearningByExample/go-microservice/store/memory"
)

func main() {
	store := memory.NewInMemoryPetStore()
	store.AddPet("pelusa", "dog", "happy")

	srv := server.NewServer(8080, store)
	srv.Serve()
}

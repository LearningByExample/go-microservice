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

package main

import (
	"errors"
	"flag"
	"github.com/LearningByExample/go-microservice/internal/app/config"
	"github.com/LearningByExample/go-microservice/internal/app/server"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"github.com/LearningByExample/go-microservice/internal/app/store/memory"
	"log"
)

var (
	logFatal            = log.Fatal
	errorStartingServer = errors.New("error starting server")
)

const (
	dog = `
   __
o-''|\_____/)
 \_/|_)     )
    \  __  /
    (_/ (_/    Pet Store
`
)

func addProviders() {
	store.AddProvider(memory.StoreName, memory.NewInMemoryPetStore)
}

func run(cfgPath string) error {
	log.Printf("Loading config from %q ...", cfgPath)
	cfg, err := config.GetConfig(cfgPath)
	if err == nil {
		log.Println("Config loaded.")
		addProviders()
		var st store.PetStore
		st, err = store.GetStoreFromProvider(cfg)
		if err == nil {
			srv := server.NewServer(cfg, st)
			if errs := srv.Start(); len(errs) != 0 {
				for _, err := range errs {
					log.Printf("Error %v.", err)
				}
				err = errorStartingServer
			}
		}
	}

	return err
}

func main() {
	print(dog)
	cfgPath := flag.String("config", "config/default.json", "configuration file path")
	flag.Parse()
	if err := run(*cfgPath); err != nil {
		logFatal(err)
	}
}

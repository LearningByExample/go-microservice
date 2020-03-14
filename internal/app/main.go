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
	"github.com/LearningByExample/go-microservice/internal/app/server"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"github.com/LearningByExample/go-microservice/internal/app/store/memory"
	"log"
)

var (
	logFatal            = log.Fatal
	errorStartingServer = errors.New("error starting server")
)

func run(port int, storeName string) error {
	store.AddStore(memory.StoreName, memory.NewInMemoryPetStore)

	st, err := store.GetStore(storeName)
	if err == nil {
		srv := server.NewServer(port, st)
		if errs := srv.Start(); len(errs) != 0 {
			for _, err := range errs {
				log.Printf("Error %v.", err)
			}
			err = errorStartingServer
		}
	}
	return err
}

func main() {
	port := flag.Int("port", 8080, "HTTP port")
	storeName := flag.String("store", "in-memory", "HTTP port")
	flag.Parse()
	if err := run(*port, *storeName); err != nil {
		logFatal(err)
	}
}

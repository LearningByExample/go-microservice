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
	"github.com/LearningByExample/go-microservice/internal/app/config"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"os"
	"path/filepath"
	"testing"
)

const (
	invalidPort    = "bad-port.json"
	invalidStore   = "bad-store.json"
	testDataFolder = "testdata"
)

func TestRun(t *testing.T) {
	t.Run("should fail with invalid port", func(t *testing.T) {
		path := filepath.Join(testDataFolder, invalidPort)
		err := run(path)
		if err == nil {
			t.Fatalf("expect error got nil")
		}
		want := errorStartingServer
		if err != want {
			t.Fatalf("expect error %v, got %v", want, err)
		}
	})

	t.Run("should fail with invalid store", func(t *testing.T) {
		path := filepath.Join(testDataFolder, invalidStore)
		err := run(path)
		if err == nil {
			t.Fatalf("expect error got nil")
		}
		want := store.ProviderNotFound
		if err != want {
			t.Fatalf("expect error %v, got %v", want, err)
		}
	})

}

func TestMainWithInvalidPort(t *testing.T) {
	var err error = nil
	savedLogFatal := logFatal
	logFatal = func(v ...interface{}) {
		if len(v) == 1 {
			switch x := v[0].(type) {
			case error:
				err = x
				break
			}
		}
	}

	oldArgs := os.Args
	os.Args = []string{"cmd", "-config", "bad"}
	main()
	os.Args = oldArgs
	logFatal = savedLogFatal

	if err == config.InvalidCfg {
		t.Fatalf("we should got invalid config, got %v", err)
	}
}

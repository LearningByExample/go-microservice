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

package config

import (
	"path/filepath"
	"testing"
)

const (
	testDataFolder = "testdata"
	cfgFile        = "cfg.json"
	badFile        = "bad.json"
	invalidFile    = "invalid.json"
	wrongPath      = "wrong"
)

func TestGetConfig(t *testing.T) {
	t.Run("should get config", func(t *testing.T) {
		path := filepath.Join(testDataFolder, cfgFile)
		_, err := GetConfig(path)

		if err != nil {
			t.Fatalf("wan't not error got %v", err)
		}
	})

	t.Run("should get an error on wrong path", func(t *testing.T) {
		path := filepath.Join(testDataFolder, wrongPath)
		_, err := GetConfig(path)

		if err == nil {
			t.Fatal("wan't error got nil")
		}
	})

	t.Run("should get an error on bad file", func(t *testing.T) {
		path := filepath.Join(testDataFolder, badFile)
		_, err := GetConfig(path)

		if err == nil {
			t.Fatal("wan't error got nil")
		}
	})

	t.Run("should get an error on invalid configuration", func(t *testing.T) {
		path := filepath.Join(testDataFolder, invalidFile)
		_, err := GetConfig(path)

		if err == nil {
			t.Fatal("wan't error got nil")
		}
	})

}

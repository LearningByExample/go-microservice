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

package data

import (
	"reflect"
	"testing"
)

func TestPet(t *testing.T) {
	pet := Pet{
		Id:   0,
		Name: "a",
		Race: "b",
		Mod:  "c",
	}

	got := pet.String()
	want := "{ Id: 0, Name: \"a\", Race: \"b\", Mod: \"c\" }"

	if got != want {
		t.Fatalf("error get pet string got %v, want %v", got, want)
	}
}

func TestSortValues(t *testing.T) {
	pm := make(PetMap)

	p1 := Pet{Id: 1}
	p2 := Pet{Id: 2}
	p3 := Pet{Id: 3}

	pm[2] = p2
	pm[3] = p3
	pm[1] = p1

	got := pm.Values()
	want := []Pet{p1, p2, p3}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

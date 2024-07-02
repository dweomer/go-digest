// Copyright 2019, 2020 OCI Contributors
// Copyright 2017 Docker, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package digest

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"strings"
	"testing"
)

func TestFlagInterface(t *testing.T) {
	var (
		alg     Algorithm
		flagSet flag.FlagSet
	)

	flagSet.Var(&alg, "algorithm", "set the digest algorithm")
	for _, testcase := range []struct {
		Name     string
		Args     []string
		Err      error
		Expected Algorithm
	}{
		{
			Name: "Invalid",
			Args: []string{"-algorithm", "bean"},
			Err:  ErrDigestUnsupported,
		},
		{
			Name:     "Default",
			Args:     []string{"unrelated"},
			Expected: "sha256",
		},
		{
			Name:     "Other",
			Args:     []string{"-algorithm", "sha512"},
			Expected: "sha512",
		},
	} {
		t.Run(testcase.Name, func(t *testing.T) {
			alg = Canonical
			if err := flagSet.Parse(testcase.Args); !errors.Is(err, testcase.Err) {
				if testcase.Err == nil {
					t.Fatal("unexpected error", err)
				}

				// check that flag package returns correct error
				if !strings.Contains(err.Error(), testcase.Err.Error()) {
					t.Fatalf("unexpected error: %v != %v", err, testcase.Err)
				}
				return
			}

			if alg != testcase.Expected {
				t.Fatalf("unexpected algorithm: %v != %v", alg, testcase.Expected)
			}
		})
	}
}

func TestFroms(t *testing.T) {
	p := make([]byte, 1<<20)
	_, err := rand.Read(p)
	if err != nil {
		t.Fatal(err)
	}

	for alg := range algorithms {
		t.Run(string(alg), func(t *testing.T) {
			h := alg.Hash()
			h.Write(p)
			expected := Digest(fmt.Sprintf("%s:%x", alg, h.Sum(nil)))

			var readerDgst Digest
			readerDgst, err = alg.FromReader(bytes.NewReader(p))
			if err != nil {
				t.Fatalf("error calculating hash from reader: %v", err)
			}

			dgsts := []Digest{
				alg.FromBytes(p),
				alg.FromString(string(p)),
				readerDgst,
			}

			if alg == Canonical {
				readerDgst, err = FromReader(bytes.NewReader(p))
				if err != nil {
					t.Fatalf("error calculating hash from reader: %v", err)
				}

				dgsts = append(dgsts,
					FromBytes(p),
					FromString(string(p)),
					readerDgst)
			}
			for _, dgst := range dgsts {
				if dgst != expected {
					t.Errorf("unexpected digest %v != %v", dgst, expected)
				}
			}
		})
	}
}

func TestBadAlgorithmNameRegistration(t *testing.T) {
	expectPanic := func(algorithm string) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("Expected panic and did not find one")
			}
			t.Logf("Captured panic: %v", r)
		}()
		// We just use SHA256 here as a test / stand-in
		RegisterAlgorithm(Algorithm(algorithm), crypto.SHA256)
	}

	expectPanic("sha256-")
	expectPanic("-")
	expectPanic("SHA256")
	expectPanic("sha25*")
}

func TestGoodAlgorithmNameRegistration(t *testing.T) {
	expectNoPanic := func(algorithm string) {
		defer func() {
			r := recover()
			if r != nil {
				t.Fatalf("Expected panic and found one: %v", r)
			}
		}()

		// We just use SHA256 here as a test / stand-in
		RegisterAlgorithm(Algorithm(algorithm), crypto.SHA256)
	}

	expectNoPanic("sha256-test")
	expectNoPanic("sha256_384")
}

package stemmers

import (
	"testing"
)

func TestConsonant(t *testing.T) {
	s := PorterStemmerEnglish{}

	s.bytes = []byte{'b'}
	if !s.consonant(0) {
		t.Errorf("Error: failed to detect that 'b' is a consonant")
	}

	s.bytes = []byte{'e'}
	if s.consonant(0) {
		t.Errorf("Error: thinks that 'e' is a consonant")
	}

	s.bytes = []byte("cya")
	if s.consonant(1) {
		t.Errorf("Error: thinks that 'y' is a vowel when a consonant")
	}

	s.bytes = []byte("yes")
	if !s.consonant(0) {
		t.Errorf("Error: thinks that 'y' is a consonant when its a vowel")
	}

	s.bytes = []byte("apple")
	if !s.consonant(3) {
		t.Errorf("Error: failed to detect that 'l' is a consonant")
	}
}

func TestMeasure(t *testing.T) {
	s := PorterStemmerEnglish{}
	s.bytes = []byte("")

	if s.measure() != 0 {
		t.Errorf("Error: measure returned wrong value")
	}

	s.bytes = []byte("fffaf")
	s.j = 4
	v := s.measure()
	if v != 1 {
		t.Errorf("Error: measure returned wrong value, got: %v\n", v)
	}

	s.bytes = []byte("fafaf")
	s.j = 4
	v = s.measure()
	if v != 2 {
		t.Errorf("Error: measure returned wrong value, got: %v\n", v)
	}
}

func TestStemmer(t *testing.T) {
	s := PorterStemmerEnglish{}

	v := s.Stem("cat")
	if v != "cat" {
		t.Errorf("Error: Stem on 'cat' returned wrong value, got: %v\n", v)
	}

	v = s.Stem("cats")
	if v != "cat" {
		t.Errorf("Error: Stem on 'cats' returned wrong value, got: %v\n", v)
	}

	v = s.Stem("running")
	if v != "run" {
		t.Errorf("Error: Stem on 'running' returned wrong value, got: %v\n", v)
	}

	v = s.Stem("running")
	if v != "run" {
		t.Errorf("Error: Stem on 'running' returned wrong value, got: %v\n", v)
	}
}

package stemmers

// EnglishPorterStemmer stems English tokens to thier common root forms.
// Eg: `foxes` becomes 'fox'
// The byte array passed in should represet a single token (word) that has been lower cased.
//
// This implementation supports the English langauge only.
// This is an implemntation is based on the original C implementation
// by Martin Porter. See: tartarus.org/martin/PorterStemmer
type PorterStemmerEnglish struct {
	j     int
	k     int
	bytes []byte
}

func (s *PorterStemmerEnglish) Stem(str string) string {
	s.bytes = []byte(str)
	s.j = 0
	s.k = 0

	i := len(str) // current offset
	end := 0

	s.k = i - 1
	if s.k > 1 {
		s.step1()
		s.step2()
		s.step3()
		s.step4()
		s.step5()
		s.step6()
	}

	end = s.k + 1
	return string(s.bytes[0:end])
}

// Consonant tests if the character at `i` in `bytes` is a constant
func (s *PorterStemmerEnglish) consonant(i int) bool {
	b := s.bytes[i]

	switch b {
	case 'a', 'e', 'i', 'o', 'u':
		return false
	}

	if b != 'y' || i == 0 {
		return true
	}

	return !s.consonant(i - 1)
}

// Measure measures the number of consonant sequences between 0 and j
func (s *PorterStemmerEnglish) measure() int {
	if len(s.bytes) == 0 {
		return 0
	}

	n := 0
	i := 0

	for {
		if i > s.j {
			return n
		}
		if !s.consonant(i) {
			break
		}
		i++
	}

	i++
	for {
		for {
			if i > s.j {
				return n
			}
			if s.consonant(i) {
				break
			}
			i++
		}

		i++
		n++
		for {
			if i > s.j {
				return n
			}
			if !s.consonant(i) {
				break
			}
			i++
		}
		i++
	}
}

// VowelInStem returns true if 0...j contains a vowel
func (s *PorterStemmerEnglish) vowelInStem() bool {
	for i := 0; i <= s.j; i++ {
		if !s.consonant(i) {
			return true
		}
	}
	return false
}

// DoubleConsonant tests for a double consonant (ie: bb) between j...j-1
func (s *PorterStemmerEnglish) doubleConsonant(i int) bool {
	if i < 1 {
		return false
	}
	if s.bytes[i] != s.bytes[i-1] {
		return false
	}
	return s.consonant(i)
}

// Cvc tests for a consonant-vowel-consonant pattern starting from i and working backwards
// This is used to restore `e` at the end of words
// such as: cav(e), lov(e), hop(e), crim(e)
// but not: snow, box, tray.
func (s *PorterStemmerEnglish) cvc(i int) bool {
	if i < 2 || !s.consonant(i) || s.consonant(i-1) || !s.consonant(i-2) {
		return false
	}

	switch s.bytes[i] {
	case 'w', 'x', 'y':
		return false
	}
	return true
}

func (s *PorterStemmerEnglish) ends(str string) bool {
	l := len(str)
	o := s.k - l + 1

	if o < 0 {
		return false
	}

	for i := 0; i < l; i++ {
		if s.bytes[o+i] != str[i] {
			return false
		}
		s.j = s.k - l
	}
	return true
}

// Setto sets (j+1),...k to the characters in the string s, readjusting
func (s *PorterStemmerEnglish) setto(str string) {
	l := len(str)
	o := s.j + 1

	for i := 0; i < l; i++ {
		s.bytes[o+i] = str[i]
	}
	s.k = s.j + l
}

func (s *PorterStemmerEnglish) r(str string) {
	if s.measure() > 0 {
		s.setto(str)
	}
}

func (s *PorterStemmerEnglish) step1() {
	if s.bytes[s.k] == 's' {
		if s.ends("sses") {
			s.k -= 2
		} else if s.ends("ies") {
			s.setto("i")
		} else if s.bytes[s.k-1] != 's' {
			s.k--
		}
	}

	if s.ends("eed") {
		if s.measure() > 0 {
			s.k--
		}
	} else if (s.ends("ed") || s.ends("ing")) && s.vowelInStem() {
		s.k = s.j

		if s.ends("at") {
			s.setto("ate")
		} else if s.ends("bl") {
			s.setto("ble")
		} else if s.ends("iz") {
			s.setto("ize")
		} else if s.doubleConsonant(s.k) {
			s.k--
			switch s.bytes[s.k] {
			case 'l', 's', 'z':
				s.k++
			}
		}
	} else if s.measure() == 1 && s.cvc(s.k) {
		s.setto("e")
	}
}

func (s *PorterStemmerEnglish) step2() {
	if s.ends("y") && s.vowelInStem() {
		s.bytes[s.k] = 'i'
	}
}

func (s *PorterStemmerEnglish) step3() {
	if s.k == 0 {
		return
	}

	switch s.bytes[s.k-1] {
	case 'a':
		if s.ends("ational") {
			s.r("ate")
			break
		}
		if s.ends("tional") {
			s.r("tion")
			break
		}
		break
	case 'c':
		if s.ends("enci") {
			s.r("ence")
			break
		}
		if s.ends("anci") {
			s.r("ance")
			break
		}
		break
	case 'e':
		if s.ends("izer") {
			s.r("ize")
			break
		}
		break
	case 'l':
		if s.ends("bli") {
			s.r("ble")
			break
		}
		if s.ends("alli") {
			s.r("al")
			break
		}
		if s.ends("entli") {
			s.r("ent")
			break
		}
		if s.ends("eli") {
			s.r("e")
			break
		}
		if s.ends("ousli") {
			s.r("ous")
			break
		}
		break
	case 'o':
		if s.ends("ization") {
			s.r("ize")
			break
		}
		if s.ends("ation") {
			s.r("ate")
			break
		}
		if s.ends("ator") {
			s.r("ate")
			break
		}
		break
	case 's':
		if s.ends("alism") {
			s.r("al")
			break
		}
		if s.ends("iveness") {
			s.r("ive")
			break
		}
		if s.ends("fulness") {
			s.r("ful")
			break
		}
		if s.ends("ousness") {
			s.r("ous")
			break
		}
		break
	case 't':
		if s.ends("aliti") {
			s.r("al")
			break
		}
		if s.ends("iviti") {
			s.r("ive")
			break
		}
		if s.ends("biliti") {
			s.r("ble")
			break
		}
		break
	case 'g':
		if s.ends("logi") {
			s.r("log")
			break
		}
	}
}

func (s *PorterStemmerEnglish) step4() {
	switch s.bytes[s.k] {
	case 'e':
		if s.ends("icate") {
			s.r("ic")
			break
		}
		if s.ends("ative") {
			s.r("")
			break
		}
		if s.ends("alize") {
			s.r("al")
			break
		}
		break
	case 'i':
		if s.ends("iciti") {
			s.r("ic")
			break
		}
		break
	case 'l':
		if s.ends("ical") {
			s.r("ic")
			break
		}
		if s.ends("ful") {
			s.r("")
			break
		}
		break
	case 's':
		if s.ends("ness") {
			s.r("")
			break
		}
		break
	}
}

func (s *PorterStemmerEnglish) step5() {
	if s.k == 0 {
		return
	}

	switch s.bytes[s.k-1] {
	case 'a':
		if s.ends("al") {
			break
		}
		return
	case 'c':
		if s.ends("ance") {
			break
		}
		if s.ends("ence") {
			break
		}
		return
	case 'e':
		if s.ends("er") {
			break
		}
		return
	case 'i':
		if s.ends("ic") {
			break
		}
		return
	case 'l':
		if s.ends("able") {
			break
		}
		if s.ends("ible") {
			break
		}
		return
	case 'n':
		if s.ends("ant") {
			break
		}
		if s.ends("ement") {
			break
		}
		if s.ends("ment") {
			break
		}
		/* element etc. not stripped before the m */
		if s.ends("ent") {
			break
		}
		return
	case 'o':
		if s.ends("ion") && s.j >= 0 && (s.bytes[s.j] == 's' || s.bytes[s.j] == 't') {
			break
		}
		if s.ends("ou") {
			break
		}
		return
	case 's':
		if s.ends("ism") {
			break
		}
		return
	case 't':
		if s.ends("ate") {
			break
		}
		if s.ends("iti") {
			break
		}
		return
	case 'u':
		if s.ends("ous") {
			break
		}
		return
	case 'v':
		if s.ends("ive") {
			break
		}
		return
	case 'z':
		if s.ends("ize") {
			break
		}
		return
	default:
		return
	}

	if s.measure() > 1 {
		s.k = s.j
	}
}

func (s *PorterStemmerEnglish) step6() {
	s.j = s.k
	if s.bytes[s.k] == 'e' {
		a := s.measure()
		if a > 1 || a == 1 && !s.cvc(s.k-1) {
			s.k--
		}
	}
	if s.bytes[s.k] == 'l' && s.doubleConsonant(s.k) && s.measure() > 1 {
		s.k--
	}
}

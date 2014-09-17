package search

import (
	"strings"
	"te/search/stemmers"
)

type (
	Token string

	Tokenizer interface {
		Tokenize(text string, stripStopWords bool) []Token
		// Clean strips and unescapes HTML
		CleanAndSplit(text string) []string
		// Stems a single word, example: `runs` becomes `run`
		Stem(word string) string
		// IsStopWord tests if the passed in word is a stop word, example: `to` or `and`
		IsStopWord(word string) bool
	}

	SimpleTokenizer struct {
		stemmer stemmers.PorterStemmerEnglish
	}
)

func NewSimpleTokenizer() SimpleTokenizer {
	return SimpleTokenizer{stemmers.PorterStemmerEnglish{}}
}

func (t *SimpleTokenizer) Tokenize(text string, stripStopWords bool) []Token {
	list := t.CleanAndSplit(text)

	seen := map[string]bool{}
	tokens := make([]Token, len(list))
	n := 0

	for _, v := range list {
		if v == "" {
			continue
		}

		// remove stop words
		if stripStopWords && isStopWord(v) {
			continue
		}

		// apply stemming
		v = t.Stem(v)

		// ignore duplicates, add to reuslts if unique
		_, ok := seen[v]
		if !ok {
			seen[v] = true
			// add token to list
			tokens[n] = Token(v)
			n++
		}
	}

	return tokens[0:n]
}

func (t *SimpleTokenizer) TokenizeWithPositions(text string, startPos int) (map[Token][]int, int) {
	list := t.CleanAndSplit(text)
	tokens := make(map[Token][]int, 0)
	pos := startPos

	for _, v := range list {
		if v == "" {
			continue
		}

		pos++

		// apply stemming
		tok := Token(t.Stem(v))

		// ignore duplicates, add to reuslts if unique
		positions, ok := tokens[tok]
		if !ok {
			positions = []int{pos}
			tokens[tok] = positions
		} else {
			positions = append(positions, pos)
		}
	}

	return tokens, pos
}

func (t *SimpleTokenizer) IsStopWord(word string) bool {
	return isStopWord(word)
}

func (t *SimpleTokenizer) Stem(word string) string {
	return t.stemmer.Stem(word)
}

func (t *SimpleTokenizer) CleanAndSplit(text string) []string {
	text = stripHtml(text)
	text = strings.ToLower(text)
	list := strings.Fields(text)
	// this must come after split because it can increase the length of list
	return cleanPunctuation(list)
}

// takes a list of tokens and returns the unique values
func uniqueTokenMap(list []Token) map[Token]bool {
	m := map[Token]bool{}
	for _, t := range list {
		m[t] = true
	}
	return m
}

// If `a` is a new list, and `b` is an older version of a
// this func returns two lists:
// 1) all tokens are in `a` that were not in `b`
// 2) all tokens were in `b` that are not in `a`
func sortNewAndOldTokens(a []Token, b map[Token]bool) ([]Token, []Token) {
	added := []Token{}
	removed := []Token{}

	ua := uniqueTokenMap(a)
	ub := b

	for t, _ := range ua {
		_, inB := ub[t]
		if !inB {
			added = append(added, t)
		}
	}

	for t, _ := range ub {
		_, inA := ua[t]
		if !inA {
			removed = append(removed, t)
		}
	}

	return added, removed
}

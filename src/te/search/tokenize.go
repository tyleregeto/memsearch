package search

import (
	"strings"
	"te/search/stemmers"
)

type Token string

type Tokenizer interface {
	Tokenize(text string) []Token
}

type SimpleTokenizer struct {
}

func (t *SimpleTokenizer) Tokenize(text string) []Token {
	text = stripHtml(text)
	list := strings.Fields(text)
	list = cleanPunctuation(list)

	tokens := []Token{}
	stemmer := stemmers.PorterStemmerEnglish{}
	seen := map[string]bool{}

	for _, v := range list {
		// lower case all tokens
		v = strings.ToLower(v)

		if v == "" {
			continue
		}

		// remove stop words
		if isStopWord(v) {
			continue
		}

		// apply stemming
		v = stemmer.Stem(v)

		// ignore duplicates, add to reuslts if unique
		_, ok := seen[v]
		if !ok {
			seen[v] = true
			// add token to list
			tokens = append(tokens, Token(v))
		}
	}

	return tokens
}

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

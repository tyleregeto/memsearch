package search

import (
	"strings"
)

// k-gram index
// [$ap] = [apple, apart, ...]
// [ap] = [apple, map, apart, , ...]
// [pp] = [apple, mapping, , ...]

// Note: we only support xyx* quries, so the index lacks *xyz indexing. Easy
// add if we need it the future
type KGramIndexTable struct {
	table map[string][]string
}

func NewKGramIndexTable() KGramIndexTable {
	return KGramIndexTable{map[string][]string{}}
}

func (i *KGramIndexTable) Add(term string, token string) {
	if term == "" {
		return
	}

	var last string
	for j, v := range term {
		var k string
		s := string(v)

		if j == 0 {
			k = "$" + s
		} else {
			k = last + s
		}

		last = s
		i.append(k, token)
	}
}

// returns a list of tokens that match term*
func (i *KGramIndexTable) Get(partialTerm Token) []Token {
	list := []Token{}
	term := string(partialTerm)

	if term == "" {
		return list
	}

	// a map[term]bool
	matching := map[string]bool{}

	var last string
	for j, v := range term {
		var k string
		s := string(v)

		if j == 0 {
			k = "$" + s
		} else {
			k = last + s
		}

		last = s
		terms, ok := i.table[k]
		if !ok {
			continue
		}

		clean := map[string]bool{}
		for _, t := range terms {
			if j == 0 {
				clean[t] = true
			} else if matching[t] {
				clean[t] = true
			}
		}
		matching = clean
	}

	// iterate over matching and populate list. This makes sure that the found
	// Token actually starts with the term to avoid false positives
	for t, _ := range matching {
		if strings.HasPrefix(t, term) || strings.HasPrefix(term, t) {
			list = append(list, Token(t))
		}
	}

	return list
}

func (i *KGramIndexTable) append(kgram string, term string) {
	table, ok := i.table[kgram]
	if !ok {
		table = []string{}
		i.table[kgram] = table
	}

	// only add unique items
	// TODO if we use a sorted list, we can search much faster
	for _, v := range table {
		if v == term {
			return
		}
	}

	// add the new item
	i.table[kgram] = append(table, term)
}

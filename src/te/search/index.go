package search

import (
	"sort"
	"strings"
)

// k-gram index
// [$ap] = [apple, apart, ...]
// [ap] = [apple, map, apart, , ...]
// [pp] = [apple, mapping, , ...]

type (
	// used for term to doc
	IndexTable struct {
		table     map[Token][]int
		nextIndex int
	}

	// used for k-gram to term
	// Note: we only support xyx* quries, so the index lacks *xyz indexing. Easy
	// add if we need it the future
	KGramIndexTable struct {
		table map[string][]string
	}
)

func NewIndexTable() IndexTable {
	return IndexTable{map[Token][]int{}, 0}
}

func (i *IndexTable) NextIndex() int {
	// TODO should have a lock around this
	i.nextIndex++
	return i.nextIndex
}

func (i *IndexTable) Add(t Token, docid int) {
	table, ok := i.table[t]
	if !ok {
		table = []int{}
		i.table[t] = table
	}

	// if the new item is smaller than the last item in the list
	// then we need to resort it. Tables must be sorted for searching
	// to work
	length := len(table)
	needsSort := length > 0 && table[length-1] > docid

	// add the new item
	i.table[t] = append(table, docid)

	// sort if needed
	if needsSort {
		sort.Ints(i.table[t])
	}
}

func (i *IndexTable) Remove(t Token, docid int) {
	// TODO should have locks around this operation
	table := i.table[t]
	idx := sort.SearchInts(table, docid)
	if idx != len(table) && table[idx] == docid {
		i.table[t] = append(table[:idx], table[idx+1:]...)
	}
}

func (i *IndexTable) Get(t Token) []int {
	row, ok := i.table[t]
	if ok {
		return row
	}
	return []int{}
}

func NewKGramIndexTable() KGramIndexTable {
	return KGramIndexTable{map[string][]string{}}
}

func (i *KGramIndexTable) Add(term string) {
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
		i.append(k, term)
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

	// iterate over matching and poulate list. This makes sure that the found
	// Token actually starts with the term to avoid false positives
	for t, _ := range matching {
		if strings.HasPrefix(t, term) {
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

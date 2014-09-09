package search

import (
	"sort"
)

type IndexTable struct {
	table     map[Token][]int
	nextIndex int
}

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

package search

import (
	"sort"
	"sync"
)

type (
	IndexTable struct {
		table     map[Token]IndexRow
		nextIndex int
		lock      sync.Mutex
	}

	IndexRow struct {
		// number of time it appears in all documents
		Frequency int
		// a list of all documents this term appear sin
		Docs []IndexDoc
	}

	// TODO This could be replresented by a single int slice [docid, freq, positions...] (is it worth it?)
	IndexDoc struct {
		Doc int
		// how many times the parent term appears in this doc
		Frequency int
		// the position of each occurance in the document, lowest to highest
		Positions []int
	}

	docSorter struct {
		Docs []IndexDoc
	}
)

func NewIndexTable() IndexTable {
	return IndexTable{table: map[Token]IndexRow{}}
}

func (i *IndexTable) NextIndex() int {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.nextIndex++
	return i.nextIndex
}

// TODO should have locks around this
func (i *IndexTable) Add(t Token, docid int, positions []int) {
	row, ok := i.table[t]
	if !ok {
		row = IndexRow{Docs: []IndexDoc{}}
	}

	freq := len(positions)
	row.Frequency += freq

	// if the new item is smaller than the last item in the list then we
	// need to re-sort it. Tables must be sorted for searching to work
	length := len(row.Docs)
	needsSort := length > 0 && row.Docs[length-1].Doc > docid

	// add the new item
	doc := IndexDoc{Doc: docid, Frequency: freq, Positions: positions}

	row.Docs = append(row.Docs, doc)

	// sort if needed
	if needsSort {
		sort.Sort(&docSorter{row.Docs})
	}

	i.table[t] = row
}

func (i *IndexTable) Remove(t Token, docid int) {
	// TODO should have locks around this operation
	row := i.table[t]
	docs := row.Docs
	idx := sort.Search(len(docs), func(i int) bool { return docs[i].Doc >= docid })

	if idx != len(docs) && docs[idx].Doc == docid {
		d := docs[idx]
		row.Frequency -= d.Frequency
		row.Docs = append(docs[:idx], docs[idx+1:]...)
	}

	i.table[t] = row
}

func (i *IndexTable) Get(t Token) []IndexDoc {
	row, ok := i.table[t]
	if ok {
		return row.Docs
	}
	return []IndexDoc{}
}

func (d *docSorter) Len() int {
	return len(d.Docs)
}

func (d *docSorter) Less(a int, b int) bool {
	return d.Docs[a].Doc < d.Docs[b].Doc
}

func (d *docSorter) Swap(a int, b int) {
	d.Docs[a], d.Docs[b] = d.Docs[b], d.Docs[a]
}

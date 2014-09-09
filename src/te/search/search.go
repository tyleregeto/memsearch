package search

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var DefaultPageSize int = 20

const (
	defaultSavePath string = "/te_search_data/%v"
	defaultSaveName string = "__default_collection"
	indexFileName   string = "_index"
)

type (
	SearchEngine struct {
		// name is used for saving
		savePath   string
		persistent bool
		index      IndexTable
		kIndex     KGramIndexTable
		// docid to doc
		documents            map[int]Document
		externalToInternalId map[string]int
		// wild card quries can be disabled on an engine level. If disabled, the index
		// never gets created, resulting in less memory usage.
		SupportWildCardQuries bool
	}

	SearchResult struct {
		Hits      int         `json:"hits"`
		Page      int         `json:"page"`
		PageSize  int         `json:"pageSize"`
		Documents []DocResult `json:"documents"`
	}

	DocResult struct {
		Id     string            `json:"id"`
		Fields map[string]string `json:"fields,omitempty"`
	}

	Document struct {
		// internal id
		Uid int
		// external id
		Id          string            `json:"id"`
		Fields      map[string]*Field `json:"fields"`
		DateAdded   time.Time         `json:"dateAdded"`
		DateUpdated time.Time         `json:"dateUpdated"`
	}

	Field struct {
		Value  string
		Tokens map[Token]bool
	}

	Query struct {
		// a query string of terms
		Terms string
		// Fields to search, can be `` to mean all or `field1|field2|field3`
		// SearchFields string
		// Fields to return, can be `` to mean all or `field1|field2|field3`
		ReturnFields string
		PageSize     int
		Page         int
		PartialMatch bool
	}

	/*
		type ComplexSearchQuery struct {
			// multiple queries are always treated as "AND" queries
			List []SearchQuery
		}
	*/

	// wrapper struct for exporting private fields to JSON
	engineJsonExport struct {
		ExternalToInternalId map[string]int
		Index                map[Token][]int
		KIndex               map[string][]string
		NextIndex            int
	}
)

func newSearchResult() SearchResult {
	return SearchResult{Documents: []DocResult{}}
}

func NewDocument() Document {
	return Document{Fields: map[string]*Field{}}
}

func NewSearchEngine() *SearchEngine {
	s := &SearchEngine{}
	s.index = NewIndexTable()
	s.kIndex = NewKGramIndexTable()
	s.documents = map[int]Document{}
	s.externalToInternalId = map[string]int{}
	s.SupportWildCardQuries = true
	return s
}

func NewPersistentSearchEngine(savePath string) *SearchEngine {
	if savePath == "" {
		savePath = fmt.Sprintf(defaultSavePath, defaultSaveName)
	}

	s := NewSearchEngine()
	s.SetPersistent(true, savePath)
	return s
}

// savePath _must_ be unique per database. If not, multiple databases
// we restore from and write to the same files, over-writing each other.
func (s *SearchEngine) SetPersistent(persistent bool, savePath string) {
	s.persistent = persistent
	s.savePath = savePath

	if persistent {
		// make sure pathh exists
		os.MkdirAll(savePath, 0770)
		// load any previous data
		s.readIndexFromDisk()
	}
}

func (s *SearchEngine) Query(query Query) SearchResult {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = DefaultPageSize
	}

	tokenizer := NewSimpleTokenizer()

	// If its not a partial match query, remove stop words
	tokens := tokenizer.Tokenize(query.Terms, !query.PartialMatch)
	docs := s._all(tokens, query.PartialMatch)

	// lookup documents
	results := newSearchResult()
	results.PageSize = query.PageSize
	results.Page = query.Page
	results.Hits = len(docs)

	start := (query.Page - 1) * query.PageSize
	if start >= results.Hits {
		return results
	}

	end := start + query.PageSize
	if end > results.Hits {
		end = results.Hits
	}

	count := end - start
	results.Documents = make([]DocResult, count)

	returnFields := map[string]bool{}
	for _, f := range strings.Split(query.ReturnFields, "|") {
		returnFields[f] = true
	}

	for i := 0; i < count; i++ {
		docid := docs[start+i]
		doc := s.documents[docid]
		res := DocResult{Id: doc.Id, Fields: map[string]string{}}

		// only return fields explicitly asked for. By default only id is returned.
		if query.ReturnFields != "" {
			// return all fields
			for k, v := range doc.Fields {
				if _, ok := returnFields[k]; ok {
					res.Fields[k] = v.Value
				}
			}
		}

		results.Documents[i] = res
	}

	return results
}

func (s *SearchEngine) QueryField(field string, query string) SearchResult {
	tokenizer := NewSimpleTokenizer()
	tokens := tokenizer.Tokenize(query, false)
	docs := s._all(tokens, false)

	// lookup documents, filter to only include matching fields
	results := newSearchResult()
	results.Page = 1
	results.PageSize = DefaultPageSize

	for _, docid := range docs {
		d := s.documents[docid]
		// We are basically ndexing per-field here, why not just make the indexing
		// global, and we can skip the initial `all` query. It will probably be smaller than
		// what we are currently doing, lots of duplication now
		f, ok := d.Fields[field]
		if !ok {
			continue
		}

		for _, t := range tokens {
			_, ok := f.Tokens[t]
			if ok {
				doc := s.documents[docid]
				res := DocResult{Id: doc.Id, Fields: map[string]string{}}

				for k, v := range doc.Fields {
					res.Fields[k] = v.Value
				}

				results.Documents = append(results.Documents, res)
			}
		}
	}

	results.Hits = len(results.Documents)
	return results
}

// Remove purges the given document from the index
func (s *SearchEngine) Remove(docid string) {
	uid, ok := s.externalToInternalId[docid]
	if ok {
		d := s.documents[uid]
		// remove the document from all tokens
		for _, f := range d.Fields {
			for t, _ := range f.Tokens {
				s.index.Remove(t, uid)
			}
		}
		delete(s.documents, uid)
	}
}

// Index adds a document to the index based on the `terms`.
// If `docid` already exists in the index, it is updated.
// `data` is the value returned when searching.
func (s *SearchEngine) Index(doc Document) {
	// get/set the documentes internal id
	uid, exists := s.externalToInternalId[doc.Id]
	if !exists {
		uid = s.index.NextIndex()
		doc.Uid = uid
		doc.DateAdded = time.Now()
		s.externalToInternalId[doc.Id] = uid
	}

	doc.Uid = uid
	doc.DateUpdated = time.Now()

	// add to the inverse index
	s.addToInverseIndex(doc, !exists)

	// add the document to the kgram index. This one is
	// opt in because it results in a large memory increase
	if s.SupportWildCardQuries {
		s.addToKgramIndex(doc)
	}

	// save the document for later retrieval
	s.documents[uid] = doc

	// write the document to disk
	if s.persistent {
		docJson, err := json.Marshal(doc)
		if err != nil {
			panic(fmt.Sprintf("Failed to save indexed file to disk, err: %v", err.Error()))
		}
		ioutil.WriteFile(fmt.Sprintf("%v/%v", s.savePath, uid), docJson, 0770)
	}

	// write the updated index to disk
	s.writeIndexToDisk()
}

func (s *SearchEngine) addToInverseIndex(doc Document, isNew bool) {
	// all tokens in the document
	tokens := []Token{}
	uniqueTokens := map[Token]bool{}
	tokenizer := NewSimpleTokenizer()

	// TODO this should be more efficient. Here we remove all tokens just
	// to re add them in the next step. We should calculate all the
	// added/removed ones and only update those.
	// This was implemented (but had bugs) so we've gone with the simple approach
	// for now, speed has been affected.

	var prevVersion Document
	if !isNew {
		prevVersion = s.documents[doc.Uid]

		for _, f := range prevVersion.Fields {
			for k := range f.Tokens {
				s.index.Remove(k, doc.Uid)
			}
		}
	}

	for _, f := range doc.Fields {
		fieldTokens := tokenizer.Tokenize(f.Value, false)
		tokens = append(tokens, fieldTokens...)
		// Store tokens on each field for further indexing
		f.Tokens = uniqueTokenMap(fieldTokens)
	}

	// index the document under all tokens
	for _, t := range tokens {
		_, seen := uniqueTokens[t]
		if !seen {
			uniqueTokens[t] = true
			s.index.Add(t, doc.Uid)
		}
	}
}

func (s *SearchEngine) addToKgramIndex(doc Document) {
	tokenizer := NewSimpleTokenizer()
	words := []string{}

	// combine all the words into one big list
	for _, f := range doc.Fields {
		words = append(words, tokenizer.CleanAndSplit(f.Value)...)
	}

	// add each word to the kgram index, passing in the tokenized value of each.
	// We index under the original word and the stemmed word, but we always reference
	// back to the tokenized value
	for _, w := range words {
		if tokenizer.IsStopWord(w) {
			continue
		}

		t := tokenizer.Stem(w)
		s.kIndex.Add(t, t)
		s.kIndex.Add(w, t)
	}
}

// returns a list of docids
func (s *SearchEngine) _all(tokens []Token, partialMatches bool) []int {
	docs := []int{}
	found := map[int]bool{}

	for _, t := range tokens {
		r := s.index.Get(t)

		// If no results found on the exact term, and partial matching is enabled
		// perform the partial matching
		if partialMatches && len(r) == 0 {
			// TODO we should include partial matches here anyways, by weight them differently
			r = s._all(s.kIndex.Get(t), false)
		}

		for _, docid := range r {
			// remove duplicate ids
			// TODO count the number of times a doc is returned for relevence ranking
			if _, ok := found[docid]; !ok {
				found[docid] = true
				docs = append(docs, docid)
			}
		}
	}

	// TODO sort docs by relevence

	return docs
}

func (s *SearchEngine) writeIndexToDisk() {
	// wrap fields we want exported in an exportable struct
	json, err := json.Marshal(engineJsonExport{
		ExternalToInternalId: s.externalToInternalId,
		Index:                s.index.table,
		KIndex:               s.kIndex.table,
		NextIndex:            s.index.nextIndex,
	})

	if err != nil {
		panic(fmt.Sprintf("Failed to save search engine disk, err: %v", err.Error()))
	}

	ioutil.WriteFile(fmt.Sprintf("%v/%v", s.savePath, indexFileName), json, 0770)

	// TODO rather than writing the whole thing everytime, we should use some type of
	// update system
}

func (s *SearchEngine) readIndexFromDisk() {
	bytes, err := ioutil.ReadFile(fmt.Sprintf("%v/%v", s.savePath, indexFileName))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var savedIndex engineJsonExport

	err = json.Unmarshal(bytes, &savedIndex)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// load docs
	// TODO probably don't want to load these all the time (unless we are targeting
	// fairly small datasets) We should load on demand as needed. First we need to
	// have better in-memory indexing
	files, err := ioutil.ReadDir(s.savePath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, f := range files {
		if f.Name() == indexFileName {
			continue
		}

		bytes, err := ioutil.ReadFile(fmt.Sprintf("%v/%v", s.savePath, f.Name()))
		if err != nil {
			continue
		}

		var d Document
		err = json.Unmarshal(bytes, &d)
		if err != nil {
			continue
		}

		s.documents[d.Uid] = d
	}

	// once all docs and the index are loaded, restore state
	s.externalToInternalId = savedIndex.ExternalToInternalId
	s.index.nextIndex = savedIndex.NextIndex
	s.index.table = savedIndex.Index
	s.kIndex.table = savedIndex.KIndex
}

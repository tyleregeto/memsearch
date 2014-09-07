package search

import (
	"fmt"
	"os"
)

// SearchServer is an interface for creating and accessing multiple named search engines
type SearchServer struct {
	savePath      string
	persistent    bool
	searchEngines map[string]*SearchEngine
}

func NewSearchServer() *SearchServer {
	s := &SearchServer{}
	s.searchEngines = map[string]*SearchEngine{}
	return s
}

func NewPersistentSearchServer(savepath string) *SearchServer {
	s := NewSearchServer()
	s.persistent = true
	s.savePath = savepath

	if savepath == "" {
		s.savePath = defaultSavePath
	}

	return s
}

func (s *SearchServer) Create(name string) bool {
	// TODO enforce alpha-numeric, return error if not

	// don't replace an existing engine
	if _, ok := s.searchEngines[name]; ok {
		return false
	}

	if s.persistent {
		savePath := fmt.Sprintf(defaultSavePath, name)
		s.searchEngines[name] = NewPersistentSearchEngine(savePath)
	} else {
		s.searchEngines[name] = NewSearchEngine()
	}

	return true
}

func (s *SearchServer) Destroy(name string) {
	if s.persistent {
		// destory persistent data
		savepath := fmt.Sprintf(defaultSavePath, name)
		os.RemoveAll(savepath)
	}
	delete(s.searchEngines, name)
}

func (s *SearchServer) Exists(name string) bool {
	_, ok := s.searchEngines[name]
	return ok
}

func (s *SearchServer) Query(engine string, query Query) SearchResult {
	e, ok := s.searchEngines[engine]
	if !ok {
		return newSearchResult()
	}
	return e.Query(query)
}

func (s *SearchServer) Index(engine string, doc Document) {
	e, ok := s.searchEngines[engine]
	if !ok {
		return
	}
	e.Index(doc)
}

// Remove purges the given document from the index
func (s *SearchServer) Remove(engine string, docid string) {
	e, ok := s.searchEngines[engine]
	if !ok {
		return
	}
	e.Remove(docid)
}

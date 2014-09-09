package search

import (
	"testing"
)

const testDataDir = "../../../test_data"

func TestSearch(t *testing.T) {
	s := NewSearchEngine()
	s.Index(Document{Id: "1", Fields: map[string]*Field{
		"title": &Field{Value: "dog fish cat"},
	},
	})

	res := s.Query(Query{Terms: "dog"})
	if res.Hits != 1 || res.Documents[0].Id != "1" {
		t.Errorf("Search failed.")
	}

	if s.Query(Query{Terms: "cat"}).Hits != 1 {
		t.Errorf("Search failed.")
	}

	if s.Query(Query{Terms: "apple"}).Hits != 0 {
		t.Errorf("Search failed.")
	}

	if s.Query(Query{Terms: "apple cat"}).Hits != 1 {
		t.Errorf("Search failed.")
	}

	if s.Query(Query{Terms: "dog cat"}).Hits != 1 {
		t.Errorf("Search failed.")
	}
}

func TestSearchMultipleDocs(t *testing.T) {
	s := NewSearchEngine()
	s.Index(Document{Id: "1", Fields: map[string]*Field{
		"title": &Field{Value: "dog fish cat"},
	},
	})
	s.Index(Document{Id: "2", Fields: map[string]*Field{
		"title": &Field{Value: "fish rat brat"},
	},
	})

	if s.Query(Query{Terms: "dog"}).Hits != 1 {
		t.Errorf("Search failed.")
	}

	if s.Query(Query{Terms: "rat"}).Hits != 1 {
		t.Errorf("Search failed.")
	}

	if s.Query(Query{Terms: "apple"}).Hits != 0 {
		t.Errorf("Search failed.")
	}

	if s.Query(Query{Terms: "fish"}).Hits != 2 {
		t.Errorf("Search failed.")
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	s := NewSearchEngine()
	s.Index(Document{Id: "1", Fields: map[string]*Field{
		"title": &Field{Value: "Dog Fish Cat"},
	},
	})
	s.Index(Document{Id: "2", Fields: map[string]*Field{
		"title": &Field{Value: "fish rat brat"},
	},
	})
	if s.Query(Query{Terms: "dog"}).Hits != 1 {
		t.Errorf("Search failed.")
	}

	if s.Query(Query{Terms: "fish"}).Hits != 2 {
		t.Errorf("Search failed.")
	}
}

func TestSearchStopWords(t *testing.T) {
	s := NewSearchEngine()
	s.Index(Document{Id: "1", Fields: map[string]*Field{
		"title": &Field{Value: "Dog Fish Cat"},
	},
	})
	s.Index(Document{Id: "2", Fields: map[string]*Field{
		"title": &Field{Value: "fish rat brat"},
	},
	})

	if s.Query(Query{Terms: "dog"}).Hits != 1 {
		t.Errorf("Search failed expected word not found amongst stop words.")
	}

	if s.Query(Query{Terms: "and"}).Hits != 0 {
		t.Errorf("Search failed, stop word found")
	}
}

func TestSearchRemove(t *testing.T) {
	s := NewSearchEngine()
	s.Index(Document{Id: "1", Fields: map[string]*Field{
		"title": &Field{Value: "Dog Fish Cat"},
	},
	})

	if s.Query(Query{Terms: "dog"}).Hits != 1 {
		t.Errorf("Search failed expected word not found amongst stop words.")
	}

	s.Remove("1")
	if s.Query(Query{Terms: "dog"}).Hits != 0 {
		t.Errorf("Failed to remove document")
	}
}

func TestFieldSearch(t *testing.T) {
	s := NewSearchEngine()
	s.Index(Document{Id: "1", Fields: map[string]*Field{
		"title": &Field{Value: "Dog Fish Cat"},
		"body":  &Field{Value: "Plane Car Truck"},
	},
	})

	if s.QueryField("title", "dog").Hits != 1 {
		t.Errorf("Search field failed, expected one result")
	}

	if s.QueryField("title", "plane").Hits != 0 {
		t.Errorf("Search field failed, expected zero result")
	}

	if s.QueryField("body", "plane").Hits != 1 {
		t.Errorf("Search field failed, expected one result")
	}
}

func TestPuncuation(t *testing.T) {
	s := NewSearchEngine()
	s.Index(Document{Id: "1", Fields: map[string]*Field{
		"title": &Field{Value: "Dogs? bears' Cat's turbo-snail"},
		"body":  &Field{Value: "Planes, trains, automobiles!, O'Niel"},
		"xyz":   &Field{Value: ""},
	},
	})

	if s.Query(Query{Terms: "dog"}).Hits != 1 {
		t.Errorf("Search failed, expected one result")
	}

	if s.Query(Query{Terms: "plane"}).Hits != 1 {
		t.Errorf("Search failed, expected one result")
	}

	if s.Query(Query{Terms: "automobiles"}).Hits != 1 {
		t.Errorf("Search failed, expected one result")
	}

	if s.Query(Query{Terms: "bear"}).Hits != 1 {
		t.Errorf("Search failed, expected one result")
	}

	if s.Query(Query{Terms: "cat"}).Hits != 1 {
		t.Errorf("Search failed, expected one result")
	}

	if s.Query(Query{Terms: "turbo"}).Hits != 1 || s.Query(Query{Terms: "snail"}).Hits != 1 || s.Query(Query{Terms: "turbo-snail"}).Hits != 1 {
		t.Errorf("Search failed, compound words expected one result")
	}

	if s.Query(Query{Terms: "ONiel"}).Hits != 1 || s.Query(Query{Terms: "O'Niel"}).Hits != 1 {
		t.Errorf("Search failed")
	}
}

func TestPersistenace(t *testing.T) {
	// create a search engine and index a document
	s := NewPersistentSearchEngine(testDataDir)
	s.Index(Document{Id: "1", Fields: map[string]*Field{
		"title": &Field{Value: "Dogs? bears' Cat's turbo-snail"},
		"body":  &Field{Value: "Planes, trains, automobiles!, O'Niel"},
		"xyz":   &Field{Value: ""},
	},
	})

	s.Index(Document{Id: "1", Fields: map[string]*Field{
		"title": &Field{Value: "Dogs? bears' Cat's turbo-snail"},
		"body":  &Field{Value: "Planes, trains, automobiles!, O'Niel"},
		"xyz":   &Field{Value: ""},
	},
	})

	// replace the search engine with a new instance, then test for the document
	// that was indexed on the previous engine
	s = NewPersistentSearchEngine(testDataDir)
	if s.Query(Query{Terms: "bear"}).Hits == 0 {
		t.Errorf("Search failed to restore index")
	}

	if len(s.index.table["automobil"]) != 1 {
		t.Errorf("Item indexed more than once")
	}
}

// A document should never be indexed more than one for a given term.
// This was happening at one point, this tests ensures it doesn't regress
func TestDuplicateIndexValues(t *testing.T) {
	s := NewSearchEngine()
	s.Index(Document{Id: "55", Fields: map[string]*Field{
		"title": &Field{Value: "Dogs? bears' Cat's turbo-snail"},
		"body":  &Field{Value: "Planes, trains, automobiles!, O'Niel"},
	},
	})

	if len(s.index.table["dog"]) != 1 {
		t.Errorf("Item indexed more than once")
	}

	s.Index(Document{Id: "55", Fields: map[string]*Field{
		"title": &Field{Value: "Dogs? bears' Cat's turbo-snail"},
		"body":  &Field{Value: "Planes, trains, automobiles!, O'Niel"},
		"xyz":   &Field{Value: "Planes, trains, automobiles!, O'Niel"},
	},
	})

	if len(s.index.table["plane"]) != 1 {
		t.Errorf("Item indexed more than once")
	}
}

func TestPartialMatching(t *testing.T) {
	s := NewSearchEngine()
	s.SupportWildCardQuries = true

	s.Index(Document{Id: "1", Fields: map[string]*Field{
		"title": &Field{Value: "program? programming, progress!"},
	},
	})

	s.Index(Document{Id: "2", Fields: map[string]*Field{
		"title": &Field{Value: "its progress!"},
	},
	})

	if s.Query(Query{Terms: "pro", PartialMatch: false}).Hits != 0 {
		t.Errorf("Search failed, expected zero result")
	}

	res := s.Query(Query{Terms: "pro", PartialMatch: true})
	if res.Hits != 2 {
		t.Errorf("Search failed, expected 2 matches, got: %v", res)
	}
}

func TestPartialMatchingFromStopWord(t *testing.T) {
	s := NewSearchEngine()
	s.SupportWildCardQuries = true

	s.Index(Document{Id: "1", Fields: map[string]*Field{
		"title": &Field{Value: "tomb program? programming, progress!"},
	},
	})

	res := s.Query(Query{Terms: "to", PartialMatch: true})
	if res.Hits == 0 {
		t.Errorf("Search failed, expected matches")
	}
}

func TestPartialMatchingAfterStemming(t *testing.T) {
	s := NewPersistentSearchEngine(testDataDir)
	s.SupportWildCardQuries = true

	s.Index(Document{Id: "1", Fields: map[string]*Field{
		"title": &Field{Value: "program? programming, progress!"},
	},
	})

	res := s.Query(Query{Terms: "programmin", PartialMatch: true})
	if res.Hits == 0 {
		t.Errorf("Search failed, expected matches")
	}
}

// new test case index a document
// then index the same document minus a field
// then index the same document with that field added in again
// you now have a duplicate entry

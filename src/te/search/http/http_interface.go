// A HTTP interface for accessing a search server instance
// Responses are returned as JSON.
//
// Example usage:
//
// import (
//     "net/http"
//     "te/search"
//     sh "te/search/http"
// )
//
// http.HandleFunc("/my/search/end/point", sh.HandlerFunc(search.NewSearchServer()))
// http.ListenAndServe(":80", nil)

package http

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"te/search"
	"time"
)

// format JSON documents are expected to be when coming through the HTTP interface
type document struct {
	Id     string            `json:"id"`
	Fields map[string]string `json:"fields"`
}

// Returns a HTTP handler function that will pass requests through
// to the specified SearchServer
func HandlerFunc(s *search.SearchServer, authToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// ifs its a get request, it can only be a query request, so hand it off
		if r.Method == "GET" {
			queryHandler(s, w, r)
			return
		}

		if r.Method != "POST" {
			respondWithError(w, r, "unsupport HTTP method")
			return
		}

		params := r.URL.Query()

		if authToken != "" {
			token := params.Get("authtoken")
			if token != authToken {
				respondWithError(w, r, "Auth token invalid")
				return
			}
		}

		action := params.Get("action")
		switch action {
		case "create":
			createHandler(s, w, r)
		case "destroy":
			destroyHandler(s, w, r)
		case "index":
			indexHandler(s, w, r)
		case "remove":
			removeHandler(s, w, r)
		default:
			respondWithError(w, r, "Unknown action specified")
		}
	}
}

// create a search engine
func createHandler(s *search.SearchServer, w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	collection := params.Get("collection")

	if collection == "" {
		respondWithError(w, r, "Collection query parameter is required")
		return
	}

	if s.Exists(collection) {
		respondWithError(w, r, "Collection already exists")
		return
	}

	s.Create(collection)
	respondWithSuccess(w, r, "collection created")
}

// destroy a search engine
func destroyHandler(s *search.SearchServer, w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	collection := params.Get("collection")

	if collection == "" {
		respondWithError(w, r, "Collection query parameter is required")
		return
	}

	if !s.Exists(collection) {
		respondWithError(w, r, "Collection does not exist")
		return
	}

	s.Destroy(collection)
	respondWithSuccess(w, r, "collection destroyed")
}

// add a document to the search engine
func indexHandler(s *search.SearchServer, w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	collection := params.Get("collection")

	if collection == "" {
		respondWithError(w, r, "Collection query parameter is required")
		return
	}

	if !s.Exists(collection) {
		respondWithError(w, r, "Collection does not exist")
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)

	if err != nil {
		respondWithError(w, r, "Error reading body")
		return
	}

	if len(bytes) == 0 {
		respondWithError(w, r, "Error document missing")
		return
	}

	var doc document
	err = json.Unmarshal(bytes, &doc)
	if err != nil {
		respondWithError(w, r, "Error parsing document JSON")
		return
	}

	if len(doc.Id) == 0 {
		respondWithError(w, r, fmt.Sprintf("Error document id is required, not found in: %v", string(bytes)))
		return
	}

	if len(doc.Fields) == 0 {
		respondWithError(w, r, "Error document is missing fields")
		return
	}

	d := search.NewDocument()
	d.Id = doc.Id
	for k, v := range doc.Fields {
		d.Fields[k] = &search.Field{Value: v}
	}

	s.Index(collection, d)
	respondWithSuccess(w, r, "Success, document indexed")
}

// remove a document from the search engine
func removeHandler(s *search.SearchServer, w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	collection := params.Get("collection")

	if collection == "" {
		respondWithError(w, r, "Collection query parameter is required")
		return
	}

	if !s.Exists(collection) {
		respondWithError(w, r, "Collection does not exist")
		return
	}

	docid := params.Get("docid")
	if docid == "" {
		respondWithError(w, r, "docid query parameter is required")
		return
	}

	s.Remove(collection, docid)
	respondWithSuccess(w, r, "Document removed")
}

// query a search engine
//
// Query a data set named 'foo' for the term 'xyz'
// ?collection=foo&query=xyz
//
// Query a data set named 'foo' for the term 'xyz' in the field 'title'
// ?collection=foo&title=xyz
//
// Query a data set named 'foo' for the term 'xyz' AND query the field 'tag' for the term 'bar'
// Return results that match both queries
// ?collection=foo&query=xyz&tag=bar
func queryHandler(s *search.SearchServer, w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	collection := params.Get("collection")

	if collection == "" {
		respondWithError(w, r, "Collection query parameter is required")
		return
	}

	if !s.Exists(collection) {
		respondWithError(w, r, "Specified collection does not exist")
		return
	}

	// TODO support more complex queries, eg: AND, OR

	query := params.Get("query")
	partialMatch := params.Get("partial") == "1"

	var err error
	var count int
	var page int

	countStr := params.Get("count")
	if countStr == "" {
		count = search.DefaultPageSize
	} else if count, err = strconv.Atoi(countStr); err != nil {
		count = search.DefaultPageSize
	}

	pageStr := params.Get("page")
	if pageStr == "" {
		page = 1
	} else if page, err = strconv.Atoi(pageStr); err != nil {
		page = 1
	}

	fields := params.Get("fields")

	res := s.Query(collection, search.Query{Terms: query, Page: page, PageSize: count, ReturnFields: fields, PartialMatch: partialMatch})
	resp := map[string]interface{}{}
	resp["success"] = true
	bytes, _ := json.Marshal(res)
	respondWithBody(w, r, string(bytes))
}

func respondWithError(w http.ResponseWriter, r *http.Request, msg string) {
	resp := map[string]interface{}{}
	resp["success"] = false
	resp["msg"] = msg
	bytes, _ := json.Marshal(resp)

	http.Error(w, string(bytes), 400)
}

func respondWithSuccess(w http.ResponseWriter, r *http.Request, msg string) {
	resp := map[string]interface{}{}
	resp["success"] = true
	resp["msg"] = msg
	bytes, _ := json.Marshal(resp)
	respondWithBody(w, r, string(bytes))
}

func respondWithBody(w http.ResponseWriter, r *http.Request, body string) {
	t := time.Now()
	t = t.Add(time.Minute * 30)
	w.Header().Set("Expires", t.Format(time.RFC1123))

	// Flag the content type. If GZipped, the browser needs this information
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	// Check if the client supports gzip, if so we'll update the response
	// writer to one that gzip's Write() calls.
	g := w
	shouldGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
	if shouldGzip {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		gz := gzip.NewWriter(w)
		g = GZipWriter{Writer: gz, ResponseWriter: w}
		defer gz.Close()
	}

	io.WriteString(g, body)
}

// GZipWriter wraps http.ResponseWriter adding GZip compression
type GZipWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w GZipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

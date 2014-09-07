package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"te/search"
	"testing"
)

var collectionName string = "c1"
var fishingDoc string = `{
	"id": "doc1",
	"fields": {
		"title": "Fishing guide",
		"body": "This is a guide to fishing. Trout, Bass, Turtles."
	}
}`

var computerDoc string = `{
	"id": "doc2",
	"fields": {
		"title": "All about computers: Windows 95 edition",
		"body": "This is an introduction guide to windows computers."
	}
}`

func TestAuth(t *testing.T) {
	server := search.NewSearchServer()
	ln := startHttpServer(":10245", server, "password")
	defer ln.Close()

	resp, err := http.Post("http://localhost:10245?action=create&collection="+collectionName+"&authtoken=password", "text/plain", nil)

	if err != nil {
		t.Error(err.Error())
	}

	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error(string(b))
	}

	if !server.Exists("c1") {
		t.Error("Failed to create collection")
	}
}

func TestAuthFail(t *testing.T) {
	server := search.NewSearchServer()
	ln := startHttpServer(":10245", server, "password")
	defer ln.Close()

	resp, err := http.Post("http://localhost:10245?action=create&collection="+collectionName+"&authtoken=wrong", "text/plain", nil)

	if err != nil {
		t.Error(err.Error())
	}

	if resp.StatusCode == 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error(string(b))
	}

	if server.Exists("c1") {
		t.Error("Create collection with wrong auth token")
	}
}

func TestCreate(t *testing.T) {
	server := search.NewSearchServer()
	ln := startHttpServer(":10245", server, "")
	defer ln.Close()

	resp, err := http.Post("http://localhost:10245?action=create&collection="+collectionName, "text/plain", nil)

	if err != nil {
		t.Error(err.Error())
	}

	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error(string(b))
	}

	if !server.Exists("c1") {
		t.Error("Failed to create collection")
	}
}

func TestDestroy(t *testing.T) {
	server := search.NewSearchServer()
	ln := startHttpServer(":10245", server, "")
	defer ln.Close()

	http.Post("http://localhost:10245?action=create&collection="+collectionName, "text/plain", nil)
	if !server.Exists("c1") {
		t.Error("Failed to create collection, can't test destory")
		return
	}

	http.Post("http://localhost:10245?action=destroy&collection="+collectionName, "text/plain", nil)
	if server.Exists("c1") {
		t.Error("Failed to destory collection")
		return
	}
}

func TestIndex(t *testing.T) {
	server := search.NewSearchServer()
	server.Create(collectionName)

	ln := startHttpServer(":10245", server, "")
	defer ln.Close()

	// insert doc over http
	resp, err := http.Post("http://localhost:10245?action=index&collection="+collectionName, "text/json", strings.NewReader(fishingDoc))

	if err != nil {
		t.Error(err.Error())
	}

	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error(string(b))
	}

	// query server
	res := server.Query(collectionName, search.Query{Terms: "fish"})
	if res.Hits != 1 {
		t.Error("Http insert failed, could not query for doc after")
	}
}

func TestQuery(t *testing.T) {
	server := search.NewSearchServer()
	server.Create(collectionName)

	ln := startHttpServer(":10245", server, "")
	defer ln.Close()

	http.Post("http://localhost:10245?action=index&collection="+collectionName, "text/json", strings.NewReader(fishingDoc))
	res, err := http.Get("http://localhost:10245?query=fish&collection=" + collectionName)

	var results search.SearchResult
	bytes, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(bytes, &results)

	if err != nil {
		t.Error(err.Error())
	}

	if results.Hits != 1 {
		t.Error("Http query failed")
	}
}

func TestRemove(t *testing.T) {
	server := search.NewSearchServer()
	server.Create(collectionName)

	ln := startHttpServer(":10246", server, "")
	defer ln.Close()

	resp, err := http.Post("http://localhost:10246?action=index&collection="+collectionName, "text/json", strings.NewReader(fishingDoc))

	res := server.Query(collectionName, search.Query{Terms: "fishing"})
	if res.Hits == 0 {
		t.Errorf("Http insert failed, could not query for doc after: %v", res)
	}

	q := fmt.Sprintf("http://localhost:10246?action=remove&collection=%v&docid=%v", collectionName, "doc1")
	resp, err = http.Post(q, "text/json", nil)

	if err != nil {
		t.Error(err.Error())
	}

	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error(string(b))
	}

	res = server.Query(collectionName, search.Query{Terms: "fish"})
	if res.Hits > 0 {
		t.Error("Http remove failed, fish doc still exists")
	}
}

// use this to start http servers over the default methods when testing
// because it allows us to shut down the server. The SDL implementation
// does not expose the listener. We need to close it at the end of each test
// to avoid port conflicts
func startHttpServer(addr string, searchServer *search.SearchServer, authToken string) net.Listener {
	httpServer := &http.Server{Addr: addr, Handler: HandlerFunc(searchServer, authToken)}
	ln, _ := net.Listen("tcp", addr)
	go httpServer.Serve(ln)
	return ln
}

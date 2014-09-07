// A standalone HTTP search server
package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"te/search"
	sh "te/search/http"
)

func main() {
	var addr string
	var collection string
	var authtoken string
	flag.StringVar(&addr, "a", ":10600", "The addres to listen on, eg: `:80`, defaults to: `:10600`")
	flag.StringVar(&collection, "c", "", "A comma seperated list of collections to ensure exist on startup.")
	flag.StringVar(&authtoken, "t", "", "The authtoken for non-query actions.")
	flag.Parse()

	s := search.NewPersistentSearchServer("./search_data")

	// ensure any default collections exist
	if collection != "" {
		list := strings.Split(collection, ",")
		for _, name := range list {
			s.Create(name)
		}
	}

	err := http.ListenAndServe(addr, sh.HandlerFunc(s, authtoken))
	fmt.Println(err.Error())
}

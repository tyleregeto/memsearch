Overview:
------------------------------

A full text, in-memory, search engine.

Cons
	- slow (compared to Lucene)
	- not distributed
	- not good for realy large data sets
	- tokenization optimized for English only

Pros
	- easy to deploy (standalone server with HTTP API)
	- easy to embed (in Go apps)
	- low memory foot print
	- no dependencies. zero.
	- works well for small to medium sized data sets
	- support for instant search style quries
	
	
Example HTTP query: http://example.com/?collection=mycollection&count=10&fields=title|description&query=dogs
Demo: http://tyleregeto.com

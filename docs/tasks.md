
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


References:
------------------------------
http://nlp.stanford.edu/IR-book/information-retrieval-book.html
http://www.elasticsearch.org/guide/en/elasticsearch/guide/current/index.html
http://www.ardendertat.com/2011/05/30/how-to-implement-a-search-engine-part-1-create-index/
http://en.wikipedia.org/wiki/Inverted_index

Issues:
https://github.com/tyleregeto/memsearch/issues/1


Tasks:
------------------------------

tyleregeto.com (where demo sits)
	- On article publish, index with search (needs to go live)
	- On article unpublish, remove from search

Extra ideas:
--------------------------------

JavaScript query interface that works offline
	- loads JSON index
	- saves for offline accessibility
	- JavaScript lib that makes search accessible offline
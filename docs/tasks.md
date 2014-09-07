
Overview:
------------------------------

A free text, in-memory, search engine.

Cons
	- slow (compared to Lucene)
	- not distributed
	- not good for large data sets (>10000)
	- tokenization optimized for English only (I'd love some commits here)

Pros
	- easy to deploy
	- easy to embed (in Golang apps)
	- HTTP API
	- low memory fot print
	- no dependencies. zero.
	- JavaScript lib that makes search accessible offline
	- ideal for small data sets (<10000)


References:
------------------------------
http://nlp.stanford.edu/IR-book/information-retrieval-book.html
http://www.elasticsearch.org/guide/en/elasticsearch/guide/current/index.html
http://www.ardendertat.com/2011/05/30/how-to-implement-a-search-engine-part-1-create-index/
http://en.wikipedia.org/wiki/Inverted_index


Tasks:
------------------------------

Query
	- don't stem tokens for keys in the kgram index
	- don't remove stop words on query tokenization, for partial matching

Relevence sorting
	- num of tokens in query that match document
	- number of times a term appears in the document

Tokenization
	- strip all HTML from strings before tokenization
	- support for programming "words", eg: c++, c#, c. Right now thee all become "c"

Stats tracking
	- track how many times each term is searched for (instant search is going to clobber this?)
	- we can use this for "did you mean?" later
	- save as json to _stats

tyleregeto.com
	- On article publish, index with search (needs to go live)
	- On article unpublish, remove from search

Query
	- Multi-field querying at one time (eg: query all for 'foo' where tagged 'bar')
	- Date based result filtering >, <, =

Examine live _index data, look for oddities
	- tokens have the same doc entered multiple times

Extra tasks:
--------------------------------
Due to stemming
 - programmin does not match anything when using partial search
 - programming matches lots
Phrase search
Support for the query: “to be or not to be”
	- this would remove stop word filtering
	- puncuation filtering should convert `they're` to `they are`
	- stop words would still get filtered on query, but:
		- if no results found, search again with them
		- if using phrase search, include them
		- if doing a partial match, include them
Spelling correction / Did you mean?
proximity matching?
fuzzy matching?
Query caching
	- to ease the performance strain on common quries and pagination changes
	- cache for 5 minutes, or similar
	- when deciding which ones to cache, we could look at cost to calculate
Merge words of similar meaning, ie: "leap" to "jump" during tokenization
Fields can have a index_mode value
	- tokenize: standard full text analize
	- as_is: no tokenization, indexed "as is" as a single value
	- ignore: ignore the field completely, only return with results
Fields can have a field_type value (string or date) this affects indexing
For the http server, consider making collection an optional parameter, default to "__default_collection" or something like that
JavaScript query interface that works offline
	- loads JSON index
	- saves for offline accessibility
When partial matching, and many terms are found for a partial, should we reduce the lsit to commonly searched terms?
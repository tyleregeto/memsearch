
References:
------------------------------
http://nlp.stanford.edu/IR-book/information-retrieval-book.html  
http://www.elasticsearch.org/guide/en/elasticsearch/guide/current/index.html  
http://en.wikipedia.org/wiki/Inverted_index  

Issues:
- insertIntoInverseIndex is slow, doing excessive work
- create isuue for index compression, its too big
- alpha sort kgram index table rows, improve insert time
- field quries have kind of fallen by the way side, need to update and support partial matches aand http quries

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

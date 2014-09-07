package search

import (
	"testing"
)

func TestHtmlStrip(t *testing.T) {
	var s string
	var e string
	var r string

	s = "<p>cat</p>"
	e = "cat"
	r = stripHtml(s)

	if stripHtml(s) != e {
		t.Errorf("expected: %v\ngot: %v", e, r)
	}

	s = "<img src=\"some/path/img.jpg\">cat</p>"
	e = "cat"
	r = stripHtml(s)

	if r != e {
		t.Errorf("expected: %v\ngot: %v", e, r)
	}

	s = `<code>&quot;use strict&quot;;</code>`
	e = `"use strict";`
	r = stripHtml(s)

	if r != e {
		t.Errorf("expected: %v\ngot: %v", e, r)
	}
}

package search

import (
	"html"
)

func stripHtml(val string) string {
	res := make([]rune, len(val))
	n := 0
	inTag := false

	// TODO remove all start tags, replace with a single empty space
	// TODO remove all end tags, replace with a single empty space
	// TODO remove attributes
	for _, r := range val {
		// if we are not in a tag, and this rune does not open one,
		// add the run and continue
		if !inTag && r != '<' {
			res[n] = r
			n++
			continue
		}

		// if we are in a tag, and this closes it. Mark tag end and continue
		if inTag && r == '>' {
			inTag = false
			continue
		}

		// if we are not in a tag, and this opens one, replace the tag with an
		// empty space, mark tag open, and continue
		if r == '<' {
			inTag = true
			continue
		}

	}

	s := string(res[0:n])
	// remove `&apos;` and things like that
	return html.UnescapeString(s)
}

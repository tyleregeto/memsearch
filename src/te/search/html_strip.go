package search

import (
	"html"
	"unicode"
)

func stripHtml(val string) string {
	length := len(val)
	res := make([]rune, length)
	n := 0
	inTag := false

	for i, r := range val {
		// if we are not in a tag, and this opens one, replace the tag with an
		// empty space, mark tag open, and continue
		if !inTag && r == '<' && i+1 < length {
			// the rune following it must be a letter (a-z) or the start of
			// a closing tag
			next := rune(val[i+1])
			if next == '/' || unicode.IsLetter(next) {
				inTag = true
				continue
			}
		}

		// if we are not in a tag, add the rune and continue
		if !inTag {
			res[n] = r
			n++
			continue
		}

		// if we are in a tag, and this closes it. Mark tag end and continue
		if inTag && r == '>' {
			inTag = false
			continue
		}
	}

	s := string(res[0:n])
	// remove `&apos;` and things like that
	return html.UnescapeString(s)
}

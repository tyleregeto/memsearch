package search

import (
	"strings"
)

// Normailizes puncuation for tokenization. Examples:
// `car's` becomes `car`
// `cars'` becomes cars
// `co-sleep` becomes `sleep` AND `co-sleep`
// `right-of-way` becomes `right-of-way` AND `right` AND `of` AND `way`
// Things like !?,. get stripped
func cleanPunctuation(list []string) []string {
	var clean []string
	var compoundWords []string

	for _, t := range list {
		var compoundWord bool
		runes := make([]rune, len(t))
		n := 0

		// TODO if is contraction, replace word with other words
		// We don't do this now because these are all stripped out as stop words
		// Once we start supporting stop word searches, this needs to work

		for i, r := range t {
			// remove standard puncuation
			if isPunc(r) {
				continue
			}

			if isHyphon(r) {
				compoundWord = true
			}

			if r == '\'' {
				// If its the first or second character we don't skip, jsut remove. This catches two cases,
				// 1) its a qouted word, and the first character is a `'`
				// 2) Its a name, such as O'Niel
				// TODO are missing cases here? Its strips everything in all other cases, eg: car's becomes car, abcd'efg becomes abcd
				if i > 1 {
					break
				}
				continue
			}

			runes[n] = r
			n++
		}

		if compoundWord {
			parts := strings.Split(string(runes[0:n]), "-")
			if !isCommonPrefix(parts[0]) {
				compoundWords = append(compoundWords, parts...)
			} else {
				compoundWords = append(compoundWords, parts[1:]...)
			}
		}

		if n > 0 {
			clean = append(clean, string(runes[0:n]))
		}
	}

	return append(clean, compoundWords...)
}

func isPunc(c rune) bool {
	switch c {
	case '.', ',', ':', ';', '{', '}', '[', ']', '?', '/', '!', '%', '&', '(', ')', '<', '>', '\\', '|', '`', '~', '+', '*', '$', '#', '"', '—':
		return true
	}
	return false
}

func isHyphon(c rune) bool {
	return c == '-'
}

func isContraction() bool {
	return false
}

// if its a commom hyphonated word, such as `co-sleep`, we don't want to index
// `co` so we just ingnore it.
func isCommonPrefix(s string) bool {
	switch s {
	case "co", "re":
		return true
	}
	return false
}

const contractions string = `aren't,are,not
can't,cannot
couldn't,could,not
didn't,did,not
doesn't,does,not
don't,do,not
hadn't,had,not
hasn't,has,not
haven't,have,not
he'd,he,would
he'll,he,will
he's,he,is
I'd,I,had
I'll,I,will
I'm,I,am
I've,I,have
isn't,is,not
it's,it,is
let's,let,us
mightn't,might,not
mustn't,must,not
shan't,shall,not
she'd,she,would
she'll,she,will
she's,she,is
shouldn't,should,not
that's,that,is
there's,there,is
they'd,they,would
they'll,they,will
they're,they,are
they've,they,have
we'd,we,would
we're,we,are
we've,we,have
weren't,were,not
what'll,what,will
what're,what,are
what's,what,is
what've,what,have
where's,where,is
who'd,who,would
who'll,who,will
who're,who,are
who's,who,is
who've,who,have
won't,will,not
wouldn't,would,not
you'd,you,would
you'll,you,will
you're,you,are
you've,you,have
`

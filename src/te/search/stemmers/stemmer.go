package stemmers

type Stemmer interface {
	Stem(s string) string
}

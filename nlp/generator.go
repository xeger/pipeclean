package nlp

type Generator interface {
	Generate(seed string) string
}

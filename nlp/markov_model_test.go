package nlp_test

import (
	"testing"

	"github.com/xeger/pipeclean/nlp"
)

func TestDeterminism(t *testing.T) {
	m := nlp.NewMarkovModel(2, " ")
	m.Train("i like pizza")
	m.Train("i like tacos")
	m.Train("i want to go to the beach")
	m.Train("i want to go to the moon")

	for i := 0; i < 10; i++ {
		if s := m.Generate("same seed every time"); s != "i like pizza" {
			t.Errorf("Variance detected (%q)", s)
			break
		}
	}
}

package nlp

type modelStats struct {
	FreqN map[int]int `json:"freq_n"`
	MaxN  int
	MinN  int
}

func (ms *modelStats) Add(s string) {
	n := len(s)
	ms.FreqN[n]++
	if n > ms.MaxN {
		ms.MaxN = n
	}
	if n < ms.MinN {
		ms.MinN = n
	}
}

func (ms *modelStats) Derive() {
	for n := range ms.FreqN {
		if n > ms.MaxN {
			ms.MaxN = n
		}
		if n < ms.MinN {
			ms.MinN = n
		}
	}
}

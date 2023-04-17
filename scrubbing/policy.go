package scrubbing

type HeuristicPolicy struct {
	In         string
	Confidence float64
	Out        string
}

// Policy captures human decisionmaking about which fields should be scrubbed, and how.
type Policy struct {
	// FieldName ensures that certain fields are always scrubbed based on their name.
	//     Keys: ield-name substring to match e.g. "email", "smtp_addr"
	//   Values: how to scrub matching fields; either "erase" or "mask"
	FieldName map[string]string          `json:"fieldname"`
	Heuristic map[string]HeuristicPolicy `json:"heuristic"`
}

func DefaultPolicy() *Policy {
	return &Policy{
		FieldName: map[string]string{
			"email":      "mask",
			"phone":      "mask",
			"postcode":   "mask",
			"postalcode": "mask",
			"zip":        "mask",
		},
		Heuristic: map[string]HeuristicPolicy{},
	}
}

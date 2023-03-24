package main

import (
	"fmt"
	"hash/fnv"
	"net/mail"
	"regexp"
	"strings"

	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/test_driver"
	_ "github.com/pingcap/tidb/parser/test_driver"

	"gonum.org/v1/gonum/mathext/prng"
)

var reEIN = regexp.MustCompile(`\d{2}-?\d{7}`)

var reSSN = regexp.MustCompile(`\d{3}-?\d{2}-?\d{4}`)

var reTelUS = regexp.MustCompile(`\(?\d{3}\)?[ -]?\d{3}-?\d{4}`)

type scrubber struct {
	source *prng.MT19937
}

func NewScrubber() *scrubber {
	return &scrubber{
		source: prng.NewMT19937(),
	}
}

func (v *scrubber) Enter(in ast.Node) (ast.Node, bool) {
	switch st := in.(type) {
	case *test_driver.ValueExpr:
		switch st.Kind() {
		case test_driver.KindString:
			scrubbed := test_driver.Datum{}
			scrubbed.SetString(v.ScrubString(st.Datum.GetString()))
			return &test_driver.ValueExpr{Datum: scrubbed}, true
		}
	}
	return in, false
}

func (v *scrubber) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

// Scrubs recognized well-formed PII from a string, preserving all other values.
// Recognizes the following:
// - email addresses
// - YAML Ruby hashes
func (v *scrubber) ScrubString(s string) string {
	if a, _ := mail.ParseAddress(s); a != nil {
		at := strings.Index(a.Address, "@")
		local, domain := a.Address[:at], a.Address[at+1:]
		local = v.mask(local)
		domain = v.mask(domain)
		return fmt.Sprintf("%s@%s", local, domain)
	}

	if reTelUS.MatchString(s) {
		dash := strings.Index(s, "-")
		if dash < 0 {
			return v.mask(s)
		}
		area, num := s[:dash], s[dash+1:]
		area = v.mask(area)
		num = v.mask(num)
		return fmt.Sprintf("%s-%s", area, num)
	} else if reEIN.MatchString(s) || reSSN.MatchString(s) {
		return v.mask(s)
	}

	if strings.Index(s, "--- !ruby/hash") == 0 {
		return "{}"
	}

	return s
}

// Scrambles letters and numbers, preserving case sensitivity.
// As a special case, preserves 0 (and thus the distribution of zero to nonzero).
// Always returns the same output for a given input.
func (v *scrubber) mask(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	v.source.Seed(h.Sum64())

	sb := []byte(s)
	for i, b := range sb {
		if b > 'a' && b < 'z' {
			sb[i] = 'a' + byte(v.source.Uint32()%26)
		} else if b > 'A' && b < 'Z' {
			sb[i] = 'A' + byte(v.source.Uint32()%26)
		} else if b > '1' && b < '9' {
			sb[i] = '1' + byte(v.source.Uint32()%9)
		}
	}

	return string(sb)
}

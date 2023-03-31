package scrubbing

import (
	"fmt"
	"hash/fnv"
	"net/mail"
	"regexp"
	"strings"

	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/test_driver"
	_ "github.com/pingcap/tidb/parser/test_driver"
	"github.com/xeger/sqlstream/nlp"

	"gonum.org/v1/gonum/mathext/prng"
)

// Preserves non-parseable lines (assuming they are comments).
const doComments = false

// Preserves INSERT statements (disable to make debug printfs readable).
const doInserts = true

// Preserves non-insert lines (LOCK/UNLOCK/SET/...).
const doMisc = false

var reEIN = regexp.MustCompile(`\d{2}-?\d{7}`)

var reSSN = regexp.MustCompile(`\d{3}-?\d{2}-?\d{4}`)

var reTelUS = regexp.MustCompile(`\(?\d{3}\)?[ -]?\d{3}-?\d{4}`)

var reZip = regexp.MustCompile(`\d{5}(-\d{4})?`)

type scrubber struct {
	source *prng.MT19937
	models []*nlp.Model
}

func NewScrubber(models []*nlp.Model) *scrubber {
	return &scrubber{
		models: models,
		source: prng.NewMT19937(),
	}
}

func (sc *scrubber) Enter(in ast.Node) (ast.Node, bool) {
	switch st := in.(type) {
	case *test_driver.ValueExpr:
		switch st.Kind() {
		case test_driver.KindString:
			scrubbed := test_driver.Datum{}
			scrubbed.SetString(sc.scrubString(st.Datum.GetString()))
			return &test_driver.ValueExpr{Datum: scrubbed}, true
		}
	}
	return in, false
}

func (sc *scrubber) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

// Removes sensitive data from an SQL statement AST.
// May modify the AST in-place (and return it), or may return a derived AST.
// Returns nil if the entire statement should be dropped.
func (sc *scrubber) Scrub(stmt ast.StmtNode) (ast.StmtNode, bool) {
	switch st := stmt.(type) {
	// for table name: st.Table.TableRefs.Left.(*ast.TableSource).Source.(*ast.TableName).Name
	// for raw values: st.Lists[0][0], etc...
	case *ast.InsertStmt:
		if doInserts {
			st.Accept(sc)
			return st, true
		} else {
			return nil, true
		}
	default:
		if doMisc {
			return stmt, false
		}
		return nil, true
	}
}

// Scrubs recognized well-formed PII from a string, preserving all other values.
// Recognizes the following:
// - email addresses
// - YAML Ruby hashes
func (sc *scrubber) scrubString(s string) string {
	if len(s) < 1024 {
		if a, _ := mail.ParseAddress(s); a != nil {
			at := strings.Index(a.Address, "@")
			local, domain := a.Address[:at], a.Address[at+1:]
			local = sc.mask(local)
			domain = sc.mask(domain)
			return fmt.Sprintf("%s@%s", local, domain)
		}
	}

	if reTelUS.MatchString(s) {
		dash := strings.Index(s, "-")
		if dash < 0 {
			return sc.mask(s)
		}
		area, num := s[:dash], s[dash+1:]
		area = sc.mask(area)
		num = sc.mask(num)
		return fmt.Sprintf("%s-%s", area, num)
	} else if reEIN.MatchString(s) || reSSN.MatchString(s) || reZip.MatchString(s) {
		return sc.mask(s)
	}

	if strings.Index(s, "--- !ruby/hash") == 0 {
		return "{}"
	}

	for _, model := range sc.models {
		// TODO: normalize spacing of s; apply only word or sentence models depending on number of spaces
		if model.Recognize(s) > 0.8 {
			// TODO -- determinism
			return model.Generate()
		}
	}

	return s
}

// Scrambles letters and numbers; preserves case, punctuation, and special characters.
// As a special case, preserves 0 (and thus the distribution of zero to nonzero).
// Always returns the same output for a given input.
func (sc *scrubber) mask(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	sc.source.Seed(h.Sum64())

	sb := []byte(s)
	for i, b := range sb {
		if b > 'a' && b < 'z' {
			sb[i] = 'a' + byte(sc.source.Uint32()%26)
		} else if b > 'A' && b < 'Z' {
			sb[i] = 'A' + byte(sc.source.Uint32()%26)
		} else if b > '1' && b < '9' {
			sb[i] = '1' + byte(sc.source.Uint32()%9)
		}
	}

	return string(sb)
}

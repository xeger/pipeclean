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

var reStreetNum = regexp.MustCompile(`^#?\d{1,5}$`)

var reStreetSuffixUS = regexp.MustCompile(`^(?i)(Ave?n?u?e?|Dri?v?e?|Str?e?e?t|Wa?y)[.]?$`)

var reTelUS = regexp.MustCompile(`^\(?\d{3}\)?[ -]?\d{3}-?\d{4}$`)

var reZip = regexp.MustCompile(`^\d{5}(-\d{4})?$`)

type scrubber struct {
	source     *prng.MT19937
	models     []nlp.Model
	confidence float64
}

func NewScrubber(models []nlp.Model, confidence float64) *scrubber {
	return &scrubber{
		models:     models,
		source:     prng.NewMT19937(),
		confidence: confidence,
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
func (sc *scrubber) scrubString(s string) string {
	// Mask email addresses w/ consistent local and domain parts.
	if len(s) < 1024 && strings.Index(s, " ") == -1 {
		if a, _ := mail.ParseAddress(s); a != nil {
			at := strings.Index(a.Address, "@")
			local, domain := a.Address[:at], a.Address[at+1:]
			local = sc.mask(local)
			domain = sc.mask(domain)
			return fmt.Sprintf("%s@%s", local, domain)
		}
	}

	// Empty serialized Ruby YAML hashes.
	if strings.Index(s, "--- !ruby/hash") == 0 {
		return "{}"
	}

	// Mask well-known numeric formats and abbreviations.
	if reTelUS.MatchString(s) {
		dash := strings.Index(s, "-")
		if dash < 0 {
			return sc.mask(s)
		}
		area, num := s[:dash], s[dash+1:]
		area = sc.mask(area)
		num = sc.mask(num)
		return fmt.Sprintf("%s-%s", area, num)
	} else if reStreetNum.MatchString(s) || reZip.MatchString(s) {
		return sc.mask(s)
	} else if reStreetSuffixUS.MatchString(s) {
		return s
	}

	// Mask each part of short phrases of 2-4 words (i.e. addresses and names).
	spaces := strings.Count(s, " ")
	if spaces > 1 && spaces < 4 {
		words := strings.Fields(s)
		for i, w := range words {
			words[i] = sc.scrubString(w)
		}
		return strings.Join(words, " ")
	}

	for _, model := range sc.models {
		if model.Recognize(s) >= sc.confidence {
			if generator, ok := model.(nlp.Generator); ok {
				return nlp.ToSameCase(generator.Generate(s), s)
			} else {
				return sc.mask(s)
			}
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

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
	"github.com/xeger/sqlstream/rand"
)

// Preserves non-parseable lines (assuming they are comments).
const doComments = false

// Preserves INSERT statements (disable to make debug printfs readable).
const doInserts = true

// Preserves non-insert lines (LOCK/UNLOCK/SET/...).
const doMisc = false

var reBase64 = regexp.MustCompile(`^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$`)

// Numeric sequence (e.g. street address) that may be part of a larger phrase.
var reContainsNum = regexp.MustCompile(`#?\d{1,5}`)

// Numeric sequence (e.g. street address).
var reNum = regexp.MustCompile(`^#?\d{1,5}$`)

var reTelUS = regexp.MustCompile(`^\(?\d{3}\)?[ -]?\d{3}-?\d{4}$`)

var reZip = regexp.MustCompile(`^\d{5}(-\d{4})?$`)

type Scrubber struct {
	salt       string
	models     []nlp.Model
	confidence float64
}

func NewScrubber(salt string, models []nlp.Model, confidence float64) *Scrubber {
	return &Scrubber{
		salt:       salt,
		models:     models,
		confidence: confidence,
	}
}

func (sc *Scrubber) Enter(in ast.Node) (ast.Node, bool) {
	switch st := in.(type) {
	case *test_driver.ValueExpr:
		switch st.Kind() {
		case test_driver.KindString:
			datum := test_driver.Datum{}
			s := st.Datum.GetString()
			if sc.EraseString(s) {
				datum.SetNull()
			} else {
				datum.SetString(sc.ScrubString(s))
			}
			return &test_driver.ValueExpr{Datum: datum}, true
		}
	}
	return in, false
}

func (sc *Scrubber) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

// Removes sensitive data from an SQL statement AST.
// May modify the AST in-place (and return it), or may return a derived AST.
// Returns nil if the entire statement should be dropped.
func (sc *Scrubber) ScrubSQL(stmt ast.StmtNode) (ast.StmtNode, bool) {
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

// Signals to remove a string entirely from the input stream and replace it
// with a format-specific empty-value.
func (sc *Scrubber) EraseString(s string) bool {
	return reBase64.MatchString(s)
}

// Masks recognized PII in a string, preserving other values.
func (sc *Scrubber) ScrubString(s string) string {
	// Mask email addresses w/ consistent local and domain parts.
	if len(s) < 1024 && strings.Index(s, " ") == -1 {
		if a, _ := mail.ParseAddress(s); a != nil {
			at := strings.Index(a.Address, "@")
			local, domain := a.Address[:at], a.Address[at+1:]
			dot := strings.LastIndex(domain, ".")
			tld := domain[dot+1:]
			prefix := domain[0:dot]

			return fmt.Sprintf("%s@%s.%s", sc.mask(local), sc.mask(prefix), tld)
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
	} else if reNum.MatchString(s) || reZip.MatchString(s) {
		return sc.mask(s)
	}

	// Mask each part of short phrases of 2-10 words that contain a numeric component.
	if reContainsNum.MatchString(s) {
		spaces := strings.Count(s, " ")
		if spaces > 0 && spaces < 10 {
			words := strings.Fields(s)
			for i, w := range words {
				words[i] = sc.ScrubString(w)
			}
			return strings.Join(words, " ")
		}
	}

	// Match against all models.
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
func (sc *Scrubber) mask(s string) string {
	rand := rand.NewRand(nlp.Clean(s))
	h := fnv.New64a()
	if sc.salt != "" {
		h.Write([]byte(sc.salt))
		h.Write([]byte{0})
	}
	h.Write([]byte(s))

	sb := []byte(s)
	for i, b := range sb {
		if b >= 'a' && b <= 'z' {
			sb[i] = 'a' + byte(rand.Uint32()%26)
		} else if b >= 'A' && b <= 'Z' {
			sb[i] = 'A' + byte(rand.Uint32()%26)
		} else if b >= '1' && b <= '9' {
			sb[i] = '1' + byte(rand.Uint32()%9)
		}
	}

	return string(sb)
}

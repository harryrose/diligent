package requirements

import (
	"fmt"
	"regexp"
)

const (
	tokStr = "str"
	tokOp = "op"
	tokNewline = "nl"
	tokEOF = "eof"
)

func Unmarshal(in []byte, out *[]Requirement) error {
	t := &tokenizer{
		in: in,
	}

	tokens := t.Tokenize()
	var outTmp []Requirement
	for {
		req, cont, err := parseLine(tokens)
		if err != nil {
			return err
		}

		if !cont {
			break
		}

		outTmp = append(outTmp, req)
	}

	*out = outTmp
	return nil
}

type Token struct {
	offset int
	typ string
	value string
}

type tokenizer struct {
	in []byte
	lastResult []byte
	pos int
}

type TokenResult struct{
	Token
	error
}

type lineState int
const (
	lsExpectPkg lineState = iota
	lsExpectOp
	lsExpectVer
	lsEndLine
	lsSuccess
)

func parseLine(tokens <-chan TokenResult) (Requirement, bool, error) {
	out := Requirement{}
	state := lsExpectPkg

	handleExpectPkgState := func(tok Token) error {
		if tok.typ == tokNewline || tok.typ == tokEOF {
			return nil
		}

		if tok.typ != tokStr {
			return fmt.Errorf("unexpected token: '%s'. expected a package name", tok.value)
		}

		out.PackageName = tok.value
		state = lsExpectOp
		return nil
	}

	handleExpectOpState := func(tok Token) error {
		if tok.typ == tokEOF || tok.typ == tokNewline {
			state = lsSuccess
			return nil
		}

		if tok.typ != tokOp {
			return fmt.Errorf("unexpected token: '%s'. expected a comparison operator", tok.value)
		}

		out.VersionOperator = tok.value
		state = lsExpectVer
		return nil
	}

	handleExpectVersionState := func(tok Token) error {
		if tok.typ != tokStr {
			return fmt.Errorf("unexpected token: '%s'. expected a version string", tok.value)
		}

		out.PackageVersion = tok.value
		state = lsEndLine
		return nil
	}

	handleEndLineState := func(tok Token) error {
		if tok.typ != tokNewline && tok.typ != tokEOF {
			return fmt.Errorf("unexpected token: '%s'. expected end of line or end of file", tok.value)
		}
		state = lsSuccess
		return nil
	}

	for tok := range tokens {
		if tok.error != nil {
			return out, false, fmt.Errorf("tokenizer error: %s", tok.error)
		}

		var err error
		switch state {
		case lsExpectPkg:
			err = handleExpectPkgState(tok.Token)

		case lsExpectOp:
			err = handleExpectOpState(tok.Token)

		case lsExpectVer:
			err = handleExpectVersionState(tok.Token)

		case lsEndLine:
			err = handleEndLineState(tok.Token)

		}

		if err != nil {
			return out, false, err
		}

		if state == lsSuccess {
			return out, true, err
		}
	}

	return out, state == lsSuccess, nil
}

func (d *tokenizer) Tokenize() (<-chan TokenResult) {
	out := make(chan TokenResult)
	go func() {
		defer close(out)
		for d.pos <= len(d.in) {
			initialPos := d.pos
			d.consumeWhile(isWhitespace)

			if d.pos >= len(d.in) {
				out <- TokenResult{
					Token: Token{
						offset: initialPos,
						typ: tokEOF,
						value: tokEOF,
					},
				}

				return
			}

			switch {
			case d.consumeWhile(isString):
				out <- TokenResult{
					Token: Token{
						offset: initialPos,
						typ:    tokStr,
						value:  string(d.lastResult),
					},
				}

			case d.consumeWhile(isOp):
				out <- TokenResult{
					Token: Token{
						offset: initialPos,
						typ:    tokOp,
						value:  string(d.lastResult),
					},
				}

			case d.consumeWhile(isNewline):
				out <- TokenResult{
					Token: Token{
						offset: initialPos,
						typ:    tokNewline,
						value:  string(d.lastResult),
					},
				}

			default:
				out <- TokenResult{
					error: fmt.Errorf("unexpected character, '%s'", string(d.in[d.pos])),
				}
			}
		}

	}()

	return out
}

var whitespace = []byte{' ','\t'}
var newline = []byte{'\r','\n'}
var strChars = regexp.MustCompile("[a-zA-Z0-9_.]")
var versionChars = regexp.MustCompile("[0-9.]")
var opChars = []byte{'=','>','<','~'}

func byteIn(needle byte, haystack []byte) bool {
	for _, h := range haystack {
		if needle == h  {
			return true
		}
	}

	return false
}


func isNewline(b byte) bool {
	return byteIn(b, newline)
}

func isWhitespace(b byte) bool {
	return byteIn(b, whitespace)
}

func isString(b byte) bool {
	return strChars.Match([]byte{b})
}

func isOp(b byte) bool {
	return byteIn(b, opChars)
}

func (d *tokenizer) consumeWhile(pred func(b byte) bool) bool {
	startIdx := d.pos

	for d.pos < len(d.in) && pred(d.in[d.pos]) {
		d.pos ++
	}


	d.lastResult = d.in[startIdx:d.pos]
	return startIdx != d.pos
}
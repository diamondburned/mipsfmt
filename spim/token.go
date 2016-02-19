package spim

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type IndentOpts struct {
	Instruction int
	Comment     int
}

var DefaultIndentOpts = IndentOpts{
	Instruction: 8,
	Comment:     40,
}

// Lines consists of multiple lines.
type Lines []Line

func (ls Lines) String() string {
	var b strings.Builder
	for _, l := range ls {
		b.WriteString(l.String())
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// Line consists of multiple tokens.
type Line struct {
	Token   Token
	Comment CommentToken
}

func (l Line) IsEmpty() bool {
	return l.Token == nil && l.Comment == (CommentToken{})
}

// String formats a line.
func (l Line) String() string {
	var b strings.Builder
	if l.Token != nil {
		b.WriteString(l.Token.String())
		b.WriteByte('\t')
	}
	if l.Comment != (CommentToken{}) {
		b.WriteString(l.Comment.String())
	}
	return strings.TrimSuffix(b.String(), "\t")
}

type Token interface {
	fmt.Stringer
	token()
}

func (CommentToken) token()     {}
func (LabelToken) token()       {}
func (InstructionToken) token() {}

type TokenParser func(*Parser, string) (Token, string)

var TokenParsers = []TokenParser{
	ParseCommentToken, // trims end of line
	ParseLabelToken,
	ParseInstructionToken, // matches whole line
}

type LabelToken struct {
	Label string
}

func ParseLabelToken(parser *Parser, line string) (Token, string) {
	noq := NoQuotes(line, "x")

	idx := strings.Index(noq, ":")
	if idx == -1 {
		return nil, line
	}

	label := strings.TrimSpace(line[:idx])
	rest := strings.TrimSpace(line[idx+1:])
	if strings.Contains(rest, "[") || strings.Contains(rest, "]") {
		return nil, line
	}

	line = rest
	return LabelToken{label}, rest
}

func (t LabelToken) String() string {
	return t.Label + ":"
}

type InstructionToken struct {
	Instr string
	Args  []string
}

var instrRe = regexp.MustCompile(`\s*(\S+)`)

func ParseInstructionToken(parser *Parser, line string) (Token, string) {
	line = strings.TrimLeftFunc(line, unicode.IsSpace)
	noq := NoQuotes(line, "x")

	instrIdx := instrRe.FindStringSubmatchIndex(noq)
	if instrIdx == nil {
		return nil, line
	}

	instr := line[instrIdx[2]:instrIdx[3]]
	token := InstructionToken{Instr: instr}

	rest := line[instrIdx[3]:]
	args := strings.Split(rest, ",")

	for i, arg := range args {
		args[i] = strings.TrimSpace(arg)
	}

	token.Args = args
	return token, ""
}

func (t InstructionToken) String() string {
	s := t.Instr
	if len(t.Args) > 0 {
		s += "\t"
		s += strings.Join(t.Args, ", ")
	}
	return s
}

type CommentToken struct {
	Comment string
	Inline  bool
}

func ParseCommentToken(parser *Parser, line string) (Token, string) {
	noq := NoQuotes(line, "x")

	idx := strings.Index(noq, "#")
	if idx == -1 {
		return nil, line
	}

	cmt := line[idx+1:]
	cmt = strings.TrimLeft(cmt, "#")
	cmt = strings.TrimLeft(cmt, " ")

	return CommentToken{
		Comment: cmt,
		Inline:  idx > 0,
	}, line[:idx]
}

func (t CommentToken) String() string {
	if t.Inline {
		return "# " + t.Comment
	}
	return "## " + t.Comment
}

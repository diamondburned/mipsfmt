package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"libdb.so/mipsfmt/spim"
)

var (
	insIndent     int
	commentIndent int
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [params] [files...]\nParameters:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.IntVar(&insIndent, "ii", 8, "Indentation for instructions in spaces")
	flag.IntVar(&commentIndent, "ci", 40, "Indentation for comments in spaces")
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	for _, file := range flag.Args() {
		if err := formatFile(file); err != nil {
			log.Fatalf("cannot format file %q: %v", file, err)
		}
	}
}

func formatFile(file string) error {
	if file == "-" {
		return format(os.Stdout, os.Stdin)
	}

	src, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("cannot open: %w", err)
	}
	defer src.Close()

	dst, err := os.CreateTemp(filepath.Dir(file), ".~*"+filepath.Ext(file))
	if err != nil {
		return fmt.Errorf("cannot create temp: %w", err)
	}
	defer os.Remove(dst.Name())
	defer dst.Close()

	dstbuf := bufio.NewWriter(dst)
	defer dstbuf.Flush()

	if err := format(dstbuf, src); err != nil {
		return err
	}

	if err := dstbuf.Flush(); err != nil {
		return fmt.Errorf("cannot flush write buffer: %w", err)
	}

	if err := dst.Close(); err != nil {
		return fmt.Errorf("cannot close written temp: %w", err)
	}

	if err := os.Rename(dst.Name(), file); err != nil {
		return fmt.Errorf("cannot mv to commit write: %w", err)
	}

	return nil
}

func format(dst io.Writer, src io.Reader) error {
	lines, err := spim.Parse(src)
	if err != nil {
		return err
	}

	blocks := []spim.Lines{nil}

	newBlock := func() {
		if blocks[len(blocks)-1] != nil {
			blocks = append(blocks, nil)
		}
	}

	addToBlockN := func(line spim.Line, n int) {
		blocks[len(blocks)-n] = append(blocks[len(blocks)-n], line)
	}

	addToBlock := func(line spim.Line) {
		addToBlockN(line, 1)
	}

	iter := spim.NewLineIterator(lines)
	for iter.Next() {
		line := iter.Current()

		if line.IsEmpty() {
			newBlock()
			continue
		}

		if _, ok := line.Token.(spim.LabelToken); ok {
			newBlock()
			addToBlock(line)
			continue
		}

		addToBlock(line)
	}

	for _, block := range blocks {
		if err := writeBlock(dst, block); err != nil {
			return err
		}
	}

	return nil
}

func writeBlock(dst io.Writer, block spim.Lines) error {
	lines := writeLinesNoComment(block)

	// Vertical align the lines.
	lines = strings.Split(valign(lines), "\n")

	// Ugly hack to add comments after we tab-align the columns before the
	// comments are added. We're only doing this for the sake of keeping a fixed
	// indentation before inline comments.
	//
	// I actually hate this so much.
	for i, s := range lines {
		if i >= len(block) {
			break
		}

		line := block[i]
		if line.Comment == (spim.CommentToken{}) {
			continue
		}

		if line.Token != nil {
			switch line.Token.(type) {
			case spim.InstructionToken:
				indent := commentIndent - (len(s) + 1)
				if indent < 1 {
					indent = 1
				}
				s += strings.Repeat(" ", indent)
			default:
				s += "\t"
			}
		} else if i > 0 && lines[i-1] != "" {
			commentIx := strings.Index(spim.NoQuotes(lines[i-1], "x"), "#")
			if commentIx != -1 {
				s += strings.Repeat(" ", commentIx)
			}
		}

		s += line.Comment.String()
		lines[i] = s
	}

	// Re-vertically align the lines.
	out := valign(lines)

	_, err := dst.Write([]byte(out))
	return err
}

func writeLinesNoComment(lines spim.Lines) []string {
	strs := make([]string, len(lines))

	iter := spim.NewLineIterator(lines)
	for iter.Next() {
		line := iter.Current()
		if line.IsEmpty() {
			continue
		}

		var s strings.Builder

		if line.Token != nil {
			_, instr := line.Token.(spim.InstructionToken)
			if instr {
				s.WriteString(strings.Repeat(" ", insIndent))
			}

			s.WriteString(line.Token.String())
		}

		strs[iter.LineNum()] = s.String()
	}

	return strs
}

func valign(lines []string) string {
	var buf strings.Builder
	tabw := tabwriter.NewWriter(&buf, 1, 0, 1, ' ', 0)

	for _, line := range lines {
		tabw.Write([]byte(line))
		tabw.Write([]byte("\n"))
	}

	tabw.Flush()

	return buf.String()
}

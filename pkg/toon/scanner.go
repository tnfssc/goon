package toon

import (
	"fmt"
	"strings"
)

// ParsedLine represents a single line with its metadata
type ParsedLine struct {
	Raw        string
	Content    string
	Indent     int
	Depth      int
	LineNumber int
}

// BlankLineInfo represents a blank line
type BlankLineInfo struct {
	LineNumber int
	Indent     int
	Depth      int
}

// ScanResult contains the result of scanning
type ScanResult struct {
	Lines      []ParsedLine
	BlankLines []BlankLineInfo
}

// LineCursor helps iterate over parsed lines
type LineCursor struct {
	lines      []ParsedLine
	blankLines []BlankLineInfo
	index      int
}

func NewLineCursor(lines []ParsedLine, blankLines []BlankLineInfo) *LineCursor {
	return &LineCursor{
		lines:      lines,
		blankLines: blankLines,
		index:      0,
	}
}

func (c *LineCursor) Peek() *ParsedLine {
	if c.index >= len(c.lines) {
		return nil
	}
	return &c.lines[c.index]
}

func (c *LineCursor) Next() *ParsedLine {
	if c.index >= len(c.lines) {
		return nil
	}
	line := &c.lines[c.index]
	c.index++
	return line
}

func (c *LineCursor) Current() *ParsedLine {
	if c.index > 0 {
		return &c.lines[c.index-1]
	}
	return nil
}

func (c *LineCursor) Advance() {
	c.index++
}

func (c *LineCursor) AtEnd() bool {
	return c.index >= len(c.lines)
}

func (c *LineCursor) GetBlankLines() []BlankLineInfo {
	return c.blankLines
}

func (c *LineCursor) Length() int {
	return len(c.lines)
}

// ScanLines processes the input string into structured lines
func ScanLines(source string, indentSize int, strict bool) (*ScanResult, error) {
	if strings.TrimSpace(source) == "" {
		return &ScanResult{}, nil
	}

	lines := strings.Split(source, "\n")
	parsed := make([]ParsedLine, 0, len(lines))
	blankLines := make([]BlankLineInfo, 0)

	for i, raw := range lines {
		lineNumber := i + 1
		indent := 0

		// Calculate indentation
		for indent < len(raw) && raw[indent] == ' ' {
			indent++
		}

		content := raw[indent:]

		// Handle blank lines
		if strings.TrimSpace(content) == "" {
			depth := indent / indentSize
			blankLines = append(blankLines, BlankLineInfo{
				LineNumber: lineNumber,
				Indent:     indent,
				Depth:      depth,
			})
			continue
		}

		// Strict mode validation
		if strict {
			// Check for tabs in leading whitespace
			whitespaceEnd := 0
			for whitespaceEnd < len(raw) && (raw[whitespaceEnd] == ' ' || raw[whitespaceEnd] == '\t') {
				whitespaceEnd++
			}
			if strings.Contains(raw[:whitespaceEnd], "\t") {
				return nil, fmt.Errorf("line %d: tabs are not allowed in indentation in strict mode", lineNumber)
			}

			// Check indentation multiple
			if indent > 0 && indent%indentSize != 0 {
				return nil, fmt.Errorf("line %d: indentation must be exact multiple of %d, but found %d spaces", lineNumber, indentSize, indent)
			}
		}

		depth := indent / indentSize
		parsed = append(parsed, ParsedLine{
			Raw:        raw,
			Content:    content,
			Indent:     indent,
			Depth:      depth,
			LineNumber: lineNumber,
		})
	}

	return &ScanResult{
		Lines:      parsed,
		BlankLines: blankLines,
	}, nil
}

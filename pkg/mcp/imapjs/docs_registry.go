package imapjs

import (
	"bytes"
	"io/fs"
	"sort"
	"strings"
	"sync"

	"github.com/go-go-golems/go-go-goja/pkg/jsdoc/extract"
	"github.com/go-go-golems/go-go-goja/pkg/jsdoc/model"
	smailnaildocs "github.com/go-go-golems/smailnail/pkg/js/modules/smailnail/docs"
)

type docsRegistry struct {
	store *model.DocStore
}

var (
	defaultRegistryOnce sync.Once
	defaultRegistry     *docsRegistry
	defaultRegistryErr  error
)

func getDefaultDocsRegistry() (*docsRegistry, error) {
	defaultRegistryOnce.Do(func() {
		defaultRegistry, defaultRegistryErr = loadDocsRegistry()
	})
	return defaultRegistry, defaultRegistryErr
}

func loadDocsRegistry() (*docsRegistry, error) {
	entries, err := fs.ReadDir(smailnaildocs.Files, smailnaildocs.Dir)
	if err != nil {
		return nil, err
	}

	store := model.NewDocStore()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := entry.Name()
		fd, err := extract.ParseFSFile(smailnaildocs.Files, path)
		if err != nil {
			return nil, err
		}
		source, err := fs.ReadFile(smailnaildocs.Files, path)
		if err != nil {
			return nil, err
		}
		populateExampleBodies(fd, source)
		store.AddFile(fd)
	}

	return &docsRegistry{store: store}, nil
}

func (r *docsRegistry) sortedConcepts() []string {
	if r == nil || r.store == nil {
		return nil
	}
	concepts := make([]string, 0, len(r.store.ByConcept))
	for concept := range r.store.ByConcept {
		concepts = append(concepts, concept)
	}
	sort.Strings(concepts)
	return concepts
}

func populateExampleBodies(fd *model.FileDoc, source []byte) {
	if fd == nil || len(fd.Examples) == 0 || len(source) == 0 {
		return
	}

	for _, example := range fd.Examples {
		if example == nil || example.Body != "" {
			continue
		}
		offset := lineOffset(source, example.Line)
		if offset < 0 || offset >= len(source) {
			continue
		}
		if body, ok := extractNextFunction(source[offset:]); ok {
			example.Body = string(bytes.TrimSpace(body))
		}
	}
}

func lineOffset(source []byte, line int) int {
	if line <= 1 {
		return 0
	}
	currentLine := 1
	for i, b := range source {
		if b == '\n' {
			currentLine++
			if currentLine == line {
				return i + 1
			}
		}
	}
	return len(source)
}

func extractNextFunction(source []byte) ([]byte, bool) {
	text := string(source)
	idx := strings.Index(text, "function ")
	if idx == -1 {
		return nil, false
	}
	open := strings.Index(text[idx:], "{")
	if open == -1 {
		return nil, false
	}
	open += idx
	closeIdx, ok := findMatchingBrace(text, open)
	if !ok {
		return nil, false
	}
	return source[idx : closeIdx+1], true
}

func findMatchingBrace(text string, open int) (int, bool) {
	depth := 0
	inSingle := false
	inDouble := false
	inTemplate := false
	inLineComment := false
	inBlockComment := false
	escaped := false

	for i := open; i < len(text); i++ {
		ch := text[i]
		next := byte(0)
		if i+1 < len(text) {
			next = text[i+1]
		}

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment {
			if ch == '*' && next == '/' {
				inBlockComment = false
				i++
			}
			continue
		}
		if inSingle {
			if ch == '\\' && !escaped {
				escaped = true
				continue
			}
			if ch == '\'' && !escaped {
				inSingle = false
			}
			escaped = false
			continue
		}
		if inDouble {
			if ch == '\\' && !escaped {
				escaped = true
				continue
			}
			if ch == '"' && !escaped {
				inDouble = false
			}
			escaped = false
			continue
		}
		if inTemplate {
			if ch == '\\' && !escaped {
				escaped = true
				continue
			}
			if ch == '`' && !escaped {
				inTemplate = false
			}
			escaped = false
			continue
		}

		if ch == '/' && next == '/' {
			inLineComment = true
			i++
			continue
		}
		if ch == '/' && next == '*' {
			inBlockComment = true
			i++
			continue
		}
		switch ch {
		case '\'':
			inSingle = true
		case '"':
			inDouble = true
		case '`':
			inTemplate = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

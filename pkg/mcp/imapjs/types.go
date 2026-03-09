package imapjs

import "github.com/go-go-golems/go-go-goja/pkg/jsdoc/model"

type ExecuteIMAPJSRequest struct {
	Code string `json:"code"`
}

type DocumentationRequest struct {
	Mode        string `json:"mode"`
	Package     string `json:"package"`
	Symbol      string `json:"symbol"`
	Example     string `json:"example"`
	Concept     string `json:"concept"`
	Query       string `json:"query"`
	Limit       int    `json:"limit"`
	IncludeBody bool   `json:"includeBody"`
}

type ToolError struct {
	Message string `json:"message"`
	Kind    string `json:"kind,omitempty"`
}

type ExecuteIMAPJSResponse struct {
	Success bool       `json:"success"`
	Value   any        `json:"value,omitempty"`
	Console []string   `json:"console,omitempty"`
	Error   *ToolError `json:"error,omitempty"`
}

type DocumentationResponse struct {
	Mode             string             `json:"mode"`
	Package          *model.Package     `json:"package,omitempty"`
	Symbols          []*model.SymbolDoc `json:"symbols,omitempty"`
	Examples         []*model.Example   `json:"examples,omitempty"`
	Concepts         []string           `json:"concepts,omitempty"`
	Summary          string             `json:"summary,omitempty"`
	RenderedMarkdown string             `json:"renderedMarkdown,omitempty"`
}

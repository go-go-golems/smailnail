package imapjs

type ExecuteIMAPJSRequest struct {
	Code string `json:"code"`
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

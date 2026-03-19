package mailruntime

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// SieveOptions holds connection parameters for a ManageSieve server.
type SieveOptions struct {
	Host     string
	Port     int
	Username string
	Password string
}

// ScriptInfo describes a Sieve script on the server.
type ScriptInfo struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// SieveCapabilities holds server capabilities.
type SieveCapabilities struct {
	Implementation string   `json:"implementation"`
	Sieve          []string `json:"sieve"`
	Notify         []string `json:"notify"`
	SASL           []string `json:"sasl"`
	StartTLS       bool     `json:"starttls"`
	Version        string   `json:"version"`
}

// SieveClient is a small ManageSieve protocol client.
type SieveClient struct {
	conn net.Conn
	r    *bufio.Reader
	caps SieveCapabilities
}

// ConnectSieve opens a ManageSieve connection and authenticates with PLAIN.
func ConnectSieve(opts SieveOptions) (*SieveClient, error) {
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	log.Debug().Str("addr", addr).Msg("connecting to ManageSieve")

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "dial ManageSieve")
	}

	sc := &SieveClient{
		conn: conn,
		r:    bufio.NewReader(conn),
	}

	if err := sc.readCapabilities(); err != nil {
		_ = conn.Close()
		return nil, errors.Wrap(err, "reading greeting")
	}

	authStr := base64.StdEncoding.EncodeToString([]byte("\x00" + opts.Username + "\x00" + opts.Password))
	if err := sc.sendLine(fmt.Sprintf("AUTHENTICATE \"PLAIN\" %q", authStr)); err != nil {
		_ = conn.Close()
		return nil, errors.Wrap(err, "send AUTHENTICATE")
	}
	if err := sc.expectOK(); err != nil {
		_ = conn.Close()
		return nil, &MailError{Name: "AuthError", Message: err.Error(), Source: "sieve"}
	}

	return sc, nil
}

func (sc *SieveClient) Capabilities() SieveCapabilities {
	return sc.caps
}

func (sc *SieveClient) ListScripts() ([]ScriptInfo, error) {
	if err := sc.sendLine("LISTSCRIPTS"); err != nil {
		return nil, errors.Wrap(err, "LISTSCRIPTS")
	}
	var scripts []ScriptInfo
	for {
		line, err := sc.readLine()
		if err != nil {
			return nil, errors.Wrap(err, "reading LISTSCRIPTS response")
		}
		if isOK(line) {
			break
		}
		if isNO(line) {
			return nil, parseSieveError(line)
		}
		name, active := parseScriptLine(line)
		if name != "" {
			scripts = append(scripts, ScriptInfo{Name: name, Active: active})
		}
	}
	return scripts, nil
}

func (sc *SieveClient) GetScript(name string) (string, error) {
	if err := sc.sendLine(fmt.Sprintf("GETSCRIPT %q", name)); err != nil {
		return "", errors.Wrap(err, "GETSCRIPT")
	}
	content, err := sc.readLiteral()
	if err != nil {
		return "", errors.Wrap(err, "reading script literal")
	}
	if err := sc.expectOK(); err != nil {
		return "", errors.Wrap(err, "GETSCRIPT OK")
	}
	return content, nil
}

func (sc *SieveClient) PutScript(name, content string, activate bool) error {
	cmd := fmt.Sprintf("PUTSCRIPT %q {%d+}\r\n%s\r\n", name, len(content), content)
	if err := sc.sendRaw(cmd); err != nil {
		return errors.Wrap(err, "PUTSCRIPT")
	}
	if err := sc.expectOK(); err != nil {
		return errors.Wrap(err, "PUTSCRIPT response")
	}
	if activate {
		return sc.Activate(name)
	}
	return nil
}

func (sc *SieveClient) Activate(name string) error {
	if err := sc.sendLine(fmt.Sprintf("SETACTIVE %q", name)); err != nil {
		return errors.Wrap(err, "SETACTIVE")
	}
	return sc.expectOK()
}

func (sc *SieveClient) Deactivate() error {
	if err := sc.sendLine(`SETACTIVE ""`); err != nil {
		return errors.Wrap(err, "SETACTIVE (deactivate)")
	}
	return sc.expectOK()
}

func (sc *SieveClient) DeleteScript(name string) error {
	if err := sc.sendLine(fmt.Sprintf("DELETESCRIPT %q", name)); err != nil {
		return errors.Wrap(err, "DELETESCRIPT")
	}
	return sc.expectOK()
}

func (sc *SieveClient) RenameScript(oldName, newName string) error {
	if err := sc.sendLine(fmt.Sprintf("RENAMESCRIPT %q %q", oldName, newName)); err != nil {
		return errors.Wrap(err, "RENAMESCRIPT")
	}
	return sc.expectOK()
}

func (sc *SieveClient) CheckScript(content string) error {
	cmd := fmt.Sprintf("CHECKSCRIPT {%d+}\r\n%s\r\n", len(content), content)
	if err := sc.sendRaw(cmd); err != nil {
		return errors.Wrap(err, "CHECKSCRIPT")
	}
	line, err := sc.readLine()
	if err != nil {
		return errors.Wrap(err, "CHECKSCRIPT response")
	}
	if isOK(line) {
		return nil
	}
	return parseSieveError(line)
}

func (sc *SieveClient) HaveSpace(name string, sizeBytes int) (bool, error) {
	if err := sc.sendLine(fmt.Sprintf("HAVESPACE %q %d", name, sizeBytes)); err != nil {
		return false, errors.Wrap(err, "HAVESPACE")
	}
	line, err := sc.readLine()
	if err != nil {
		return false, errors.Wrap(err, "HAVESPACE response")
	}
	return isOK(line), nil
}

func (sc *SieveClient) Logout() error {
	_ = sc.sendLine("LOGOUT")
	return sc.conn.Close()
}

func (sc *SieveClient) sendLine(s string) error {
	return sc.sendRaw(s + "\r\n")
}

func (sc *SieveClient) sendRaw(s string) error {
	log.Trace().Str("cmd", s).Msg("sieve send")
	_, err := fmt.Fprint(sc.conn, s)
	return err
}

func (sc *SieveClient) readLine() (string, error) {
	line, err := sc.r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimRight(line, "\r\n")
	log.Trace().Str("line", line).Msg("sieve recv")
	return line, nil
}

func (sc *SieveClient) expectOK() error {
	line, err := sc.readLine()
	if err != nil {
		return err
	}
	if isOK(line) {
		return nil
	}
	return parseSieveError(line)
}

func isOK(line string) bool {
	return strings.HasPrefix(strings.ToUpper(line), "OK")
}

func isNO(line string) bool {
	return strings.HasPrefix(strings.ToUpper(line), "NO") || strings.HasPrefix(strings.ToUpper(line), "BYE")
}

func parseSieveError(line string) error {
	return &MailError{
		Name:    "SieveError",
		Message: line,
		Source:  "sieve",
	}
}

func (sc *SieveClient) readCapabilities() error {
	for {
		line, err := sc.readLine()
		if err != nil {
			return err
		}
		if isOK(line) {
			break
		}
		if isNO(line) {
			return fmt.Errorf("server error: %s", line)
		}
		sc.parseCapLine(line)
	}
	return nil
}

func (sc *SieveClient) parseCapLine(line string) {
	parts := splitQuoted(line)
	if len(parts) == 0 {
		return
	}
	key := strings.ToUpper(parts[0])
	val := ""
	if len(parts) > 1 {
		val = parts[1]
	}
	switch key {
	case "IMPLEMENTATION":
		sc.caps.Implementation = val
	case "SIEVE":
		sc.caps.Sieve = strings.Fields(val)
	case "NOTIFY":
		sc.caps.Notify = strings.Fields(val)
	case "SASL":
		sc.caps.SASL = strings.Fields(val)
	case "STARTTLS":
		sc.caps.StartTLS = true
	case "VERSION":
		sc.caps.Version = val
	}
}

func (sc *SieveClient) readLiteral() (string, error) {
	line, err := sc.readLine()
	if err != nil {
		return "", err
	}
	if isNO(line) {
		return "", parseSieveError(line)
	}
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "{") || !strings.HasSuffix(line, "}") {
		return "", fmt.Errorf("expected literal size, got: %s", line)
	}
	n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(line, "{"), "}"))
	if err != nil {
		return "", errors.Wrap(err, "parse literal size")
	}
	buf := make([]byte, n)
	if _, err := sc.r.Read(buf); err != nil {
		return "", errors.Wrap(err, "read literal bytes")
	}
	if _, err := sc.readLine(); err != nil {
		return "", errors.Wrap(err, "read literal terminator")
	}
	return string(buf), nil
}

func parseScriptLine(line string) (string, bool) {
	active := strings.HasSuffix(line, " ACTIVE")
	line = strings.TrimSuffix(line, " ACTIVE")
	parts := splitQuoted(line)
	if len(parts) == 0 {
		return "", active
	}
	return parts[0], active
}

func splitQuoted(line string) []string {
	var parts []string
	for len(line) > 0 {
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if line[0] == '"' {
			line = line[1:]
			idx := strings.IndexByte(line, '"')
			if idx < 0 {
				parts = append(parts, line)
				break
			}
			parts = append(parts, line[:idx])
			line = line[idx+1:]
			continue
		}
		idx := strings.IndexByte(line, ' ')
		if idx < 0 {
			parts = append(parts, line)
			break
		}
		parts = append(parts, line[:idx])
		line = line[idx+1:]
	}
	return parts
}

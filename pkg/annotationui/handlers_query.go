package annotationui

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type queryRequest struct {
	SQL string `json:"sql"`
}

type saveQueryRequest struct {
	Name        string `json:"name"`
	Folder      string `json:"folder"`
	Description string `json:"description"`
	SQL         string `json:"sql"`
}

type validatedQueryPath struct {
	rootDir      string
	relativePath string
	absolutePath string
}

var mutatingQueryPattern = regexp.MustCompile(`(?i)\b(insert|update|delete|drop|alter|create|replace|vacuum|attach|detach|reindex|truncate|grant|revoke|analyze)\b`)

func (h *appHandler) handleExecuteQuery(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	req := queryRequest{}
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.SQL) == "" {
		writeMessageError(w, http.StatusBadRequest, "sql is required")
		return
	}
	if err := validateReadOnlyQuery(req.SQL); err != nil {
		writeMessageError(w, http.StatusBadRequest, err.Error())
		return
	}

	rows, err := h.db.QueryxContext(r.Context(), req.SQL)
	if err != nil {
		writeMessageError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, errors.Wrap(err, "read query columns").Error())
		return
	}

	resultRows := make([]map[string]any, 0)
	for rows.Next() {
		row := map[string]any{}
		if err := rows.MapScan(row); err != nil {
			writeMessageError(w, http.StatusInternalServerError, errors.Wrap(err, "scan query row").Error())
			return
		}
		for key, value := range row {
			row[key] = normalizeQueryValue(value)
		}
		resultRows = append(resultRows, row)
	}
	if err := rows.Err(); err != nil {
		writeMessageError(w, http.StatusInternalServerError, errors.Wrap(err, "iterate query rows").Error())
		return
	}

	writeJSON(w, http.StatusOK, QueryResult{
		Columns:    columns,
		Rows:       resultRows,
		DurationMs: time.Since(start).Milliseconds(),
		RowCount:   len(resultRows),
	})
}

func (h *appHandler) handleGetPresets(w http.ResponseWriter, r *http.Request) {
	presets, err := loadEmbeddedQueries()
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, dir := range h.presetDirs {
		loaded, err := loadQueriesFromDir(dir)
		if err != nil {
			writeMessageError(w, http.StatusInternalServerError, err.Error())
			return
		}
		presets = mergeQueries(presets, loaded)
	}
	writeJSON(w, http.StatusOK, presets)
}

func (h *appHandler) handleGetSavedQueries(w http.ResponseWriter, r *http.Request) {
	queries := []SavedQuery{}
	for _, dir := range h.queryDirs {
		loaded, err := loadQueriesFromDir(dir)
		if err != nil {
			writeMessageError(w, http.StatusInternalServerError, err.Error())
			return
		}
		queries = mergeQueries(queries, loaded)
	}
	writeJSON(w, http.StatusOK, queries)
}

func (h *appHandler) handleSaveQuery(w http.ResponseWriter, r *http.Request) {
	req := saveQueryRequest{}
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.SQL) == "" {
		writeMessageError(w, http.StatusBadRequest, "sql is required")
		return
	}

	baseDir, err := firstQueryDir(h.queryDirs)
	if err != nil {
		writeMessageError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	queryPath, err := buildQueryCreatePath(baseDir, req.Folder, req.Name)
	if err != nil {
		writeMessageError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := queryPath.MkdirAllParent(); err != nil {
		writeMessageError(w, http.StatusInternalServerError, errors.Wrap(err, "create query directory").Error())
		return
	}
	if err := queryPath.WriteFile([]byte(buildSQLFileContent(req.Description, req.SQL))); err != nil {
		writeMessageError(w, http.StatusInternalServerError, errors.Wrap(err, "write query file").Error())
		return
	}

	saved, err := loadSingleQueryFile(baseDir, queryPath.relativePath)
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, saved)
}

func validateReadOnlyQuery(sqlText string) error {
	sqlText = strings.TrimSpace(sqlText)
	if sqlText == "" {
		return errors.New("sql is required")
	}

	trimmed := strings.TrimSuffix(sqlText, ";")
	if strings.Contains(trimmed, ";") {
		return errors.New("multiple SQL statements are not allowed")
	}

	leading := leadingSQLKeyword(sqlText)
	switch leading {
	case "SELECT", "WITH", "EXPLAIN":
	default:
		return errors.New("only read-only SELECT/WITH/EXPLAIN queries are allowed")
	}

	if mutatingQueryPattern.MatchString(sqlText) {
		return errors.New("query must be read-only")
	}
	return nil
}

func leadingSQLKeyword(sqlText string) string {
	for _, line := range strings.Split(sqlText, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		return strings.ToUpper(fields[0])
	}
	return ""
}

func normalizeQueryValue(value any) any {
	switch typed := value.(type) {
	case []byte:
		return string(typed)
	case time.Time:
		return typed.Format(time.RFC3339Nano)
	default:
		return value
	}
}

func loadEmbeddedQueries() ([]SavedQuery, error) {
	ret := make([]SavedQuery, 0)
	err := fs.WalkDir(embeddedQueries, "queries", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".sql") {
			return nil
		}
		content, err := fs.ReadFile(embeddedQueries, path)
		if err != nil {
			return err
		}
		folder := filepath.ToSlash(filepath.Dir(strings.TrimPrefix(path, "queries/")))
		if folder == "." {
			folder = ""
		}
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		ret = append(ret, SavedQuery{
			Name:        name,
			Folder:      folder,
			Description: extractSQLComment(string(content)),
			SQL:         string(content),
		})
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "walk embedded queries")
	}
	sortSavedQueries(ret)
	return ret, nil
}

func loadQueriesFromDir(dir string) ([]SavedQuery, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return []SavedQuery{}, nil
	}
	if _, err := os.Stat(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []SavedQuery{}, nil
		}
		return nil, errors.Wrap(err, "stat query directory")
	}

	ret := make([]SavedQuery, 0)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".sql") {
			return nil
		}
		relativePath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		query, err := loadSingleQueryFile(dir, filepath.ToSlash(relativePath))
		if err != nil {
			return err
		}
		ret = append(ret, query)
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "walk query directory")
	}
	sortSavedQueries(ret)
	return ret, nil
}

func loadSingleQueryFile(rootDir, relativePath string) (SavedQuery, error) {
	queryPath, err := newValidatedQueryPath(rootDir, relativePath)
	if err != nil {
		return SavedQuery{}, err
	}
	content, err := queryPath.ReadFile()
	if err != nil {
		return SavedQuery{}, errors.Wrap(err, "read query file")
	}

	folder := filepath.ToSlash(filepath.Dir(queryPath.relativePath))
	if folder == "." {
		folder = ""
	}
	return SavedQuery{
		Name:        strings.TrimSuffix(filepath.Base(queryPath.relativePath), filepath.Ext(queryPath.relativePath)),
		Folder:      folder,
		Description: extractSQLComment(string(content)),
		SQL:         string(content),
	}, nil
}

func mergeQueries(base []SavedQuery, overlay []SavedQuery) []SavedQuery {
	index := map[string]SavedQuery{}
	for _, query := range base {
		index[queryKey(query)] = query
	}
	for _, query := range overlay {
		index[queryKey(query)] = query
	}
	ret := make([]SavedQuery, 0, len(index))
	for _, query := range index {
		ret = append(ret, query)
	}
	sortSavedQueries(ret)
	return ret
}

func queryKey(query SavedQuery) string {
	return query.Folder + "/" + query.Name
}

func sortSavedQueries(queries []SavedQuery) {
	sort.Slice(queries, func(i, j int) bool {
		if queries[i].Folder == queries[j].Folder {
			return queries[i].Name < queries[j].Name
		}
		return queries[i].Folder < queries[j].Folder
	})
}

func extractSQLComment(sqlText string) string {
	for _, line := range strings.Split(sqlText, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-- ") {
			return strings.TrimPrefix(line, "-- ")
		}
		if line != "" && !strings.HasPrefix(line, "--") {
			break
		}
	}
	return ""
}

func buildSQLFileContent(description, sqlText string) string {
	description = strings.TrimSpace(description)
	sqlText = strings.TrimSpace(sqlText)
	if description == "" {
		return sqlText + "\n"
	}
	return "-- " + description + "\n" + sqlText + "\n"
}

func firstQueryDir(queryDirs []string) (string, error) {
	if len(queryDirs) == 0 {
		return "", errors.New("query-dir is not configured")
	}
	return queryDirs[0], nil
}

func buildQueryCreatePath(baseDir, folder, name string) (validatedQueryPath, error) {
	safeName := sanitizeFilename(name)
	if safeName == "" {
		return validatedQueryPath{}, errors.New("query name is required")
	}
	relativePath := safeName + ".sql"
	if strings.TrimSpace(folder) != "" {
		cleanFolder, err := cleanRelativePath(folder)
		if err != nil {
			return validatedQueryPath{}, err
		}
		relativePath = filepath.ToSlash(filepath.Join(cleanFolder, relativePath))
	}
	return newValidatedQueryPath(baseDir, relativePath)
}

func newValidatedQueryPath(baseDir, relativePath string) (validatedQueryPath, error) {
	cleanPath, err := cleanRelativePath(relativePath)
	if err != nil {
		return validatedQueryPath{}, err
	}

	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return validatedQueryPath{}, errors.Wrap(err, "resolve base directory")
	}
	absolutePath := filepath.Join(baseAbs, filepath.FromSlash(cleanPath))
	absolutePath, err = filepath.Abs(absolutePath)
	if err != nil {
		return validatedQueryPath{}, errors.Wrap(err, "resolve query path")
	}

	relativeToBase, err := filepath.Rel(baseAbs, absolutePath)
	if err != nil {
		return validatedQueryPath{}, errors.Wrap(err, "resolve query path relative to base")
	}
	relativeToBase = filepath.ToSlash(relativeToBase)
	if relativeToBase == "." || relativeToBase == "" {
		return validatedQueryPath{}, errors.New("query path must point to a file")
	}
	if strings.HasPrefix(relativeToBase, "../") || relativeToBase == ".." || filepath.IsAbs(relativeToBase) {
		return validatedQueryPath{}, errors.New("query path escapes query directory")
	}

	return validatedQueryPath{
		rootDir:      baseAbs,
		relativePath: relativeToBase,
		absolutePath: absolutePath,
	}, nil
}

func (p validatedQueryPath) ensureWithinRoot() error {
	if strings.TrimSpace(p.rootDir) == "" || strings.TrimSpace(p.relativePath) == "" || strings.TrimSpace(p.absolutePath) == "" {
		return errors.New("query path is not initialized")
	}
	relativeToRoot, err := filepath.Rel(p.rootDir, p.absolutePath)
	if err != nil {
		return errors.Wrap(err, "resolve query path relative to root")
	}
	relativeToRoot = filepath.ToSlash(relativeToRoot)
	if relativeToRoot == "." || relativeToRoot == "" {
		return errors.New("query path must point to a file")
	}
	if strings.HasPrefix(relativeToRoot, "../") || relativeToRoot == ".." || filepath.IsAbs(relativeToRoot) {
		return errors.New("query path escapes query directory")
	}
	return nil
}

func (p validatedQueryPath) MkdirAllParent() error {
	if err := p.ensureWithinRoot(); err != nil {
		return err
	}
	return os.MkdirAll(filepath.Dir(p.absolutePath), 0o755)
}

func (p validatedQueryPath) WriteFile(content []byte) error {
	if err := p.ensureWithinRoot(); err != nil {
		return err
	}
	return os.WriteFile(p.absolutePath, content, 0o644)
}

func (p validatedQueryPath) ReadFile() ([]byte, error) {
	if err := p.ensureWithinRoot(); err != nil {
		return nil, err
	}
	return os.ReadFile(p.absolutePath)
}

func cleanRelativePath(path string) (string, error) {
	path = strings.ReplaceAll(strings.TrimSpace(path), "\\", "/")
	if path == "" {
		return "", errors.New("path is required")
	}
	if strings.HasPrefix(path, "/") {
		return "", errors.New("absolute paths are not allowed")
	}
	cleaned := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if cleaned == "." || cleaned == "" {
		return "", errors.New("path is required")
	}
	if strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return "", errors.New("parent directory traversal is not allowed")
	}
	return cleaned, nil
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	var builder strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-', r == '_':
			builder.WriteRune(r)
		default:
			builder.WriteRune('-')
		}
	}
	return strings.Trim(strings.TrimSpace(builder.String()), "-")
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

func TestWriteGeneratedFileOverwritesModelOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample_model.go")
	require.NoError(t, os.WriteFile(path, []byte("old model"), 0o600))
	tmpl := template.Must(template.New("model").Parse("model {{.Domain}}"))

	written, err := writeGeneratedFile(path, tmpl, DomainData{Domain: "Sample"}, true)

	require.NoError(t, err)
	require.True(t, written)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "model Sample", string(content))
}

func TestWriteGeneratedFilePreservesExistingSupportingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample_raw.go")
	require.NoError(t, os.WriteFile(path, []byte("custom queries"), 0o600))
	tmpl := template.Must(template.New("raw").Parse("generated raw"))

	written, err := writeGeneratedFile(path, tmpl, DomainData{}, false)

	require.NoError(t, err)
	require.False(t, written)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "custom queries", string(content))
}

func TestWriteGeneratedFileCreatesMissingSupportingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample_domain.go")
	tmpl := template.Must(template.New("domain").Parse("domain {{.Domain}}"))

	written, err := writeGeneratedFile(path, tmpl, DomainData{Domain: "Sample"}, false)

	require.NoError(t, err)
	require.True(t, written)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "domain Sample", string(content))
}

func TestParseExtraFieldsIntegerNullabilityAndDefaults(t *testing.T) {
	tests := []struct {
		name        string
		declaration string
		field       string
		schema      string
		defaulted   bool
	}{
		{
			name:        "implicit nullable integer",
			declaration: "score INT",
			field:       "Score *int `json:\"score\"`",
			schema:      "\"score\": \"int\"",
		},
		{
			name:        "explicit nullable integer",
			declaration: "score INT NULL",
			field:       "Score *int `json:\"score\"`",
			schema:      "\"score\": \"int\"",
		},
		{
			name:        "nullable integer with explicit null default",
			declaration: "score INT DEFAULT NULL",
			field:       "Score *int `json:\"score\"`",
			schema:      "\"score\": \"int\"",
			defaulted:   true,
		},
		{
			name:        "required integer",
			declaration: "score INT NOT NULL",
			field:       "Score int `json:\"score\" validate:\"required\"`",
			schema:      "\"score\": \"int\"",
		},
		{
			name:        "required integer with default",
			declaration: "score INT NOT NULL DEFAULT 0",
			field:       "Score *int `json:\"score\" validate:\"required\"`",
			schema:      "\"score\": \"int\"",
			defaulted:   true,
		},
		{
			name:        "lowercase nullable integer",
			declaration: "score int null",
			field:       "Score *int `json:\"score\"`",
			schema:      "\"score\": \"int\"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ddl := fmt.Sprintf("CREATE TABLE sample (\n%s,\nPRIMARY KEY (id)\n)", test.declaration)
			fields, _, _, _, schema, defaults, _, _ := parseExtraFields(ddl)

			require.Equal(t, test.field, fields)
			require.Equal(t, test.schema, schema)
			require.Equal(t, test.defaulted, strings.Contains(defaults, "\"score\""))
		})
	}
}

func TestParseExtraFieldsSkipsMultilineColumnComment(t *testing.T) {
	ddl := `CREATE TABLE project_draw_disbursement (
		id varchar(26) NOT NULL,
		source_line_key varchar(255) NOT NULL
			COMMENT '{norm(party_name)}|{norm(work_type)}#{n}, built from extracted values',
		page_number int DEFAULT NULL,
		PRIMARY KEY (id)
	)`

	fields, columns, values, _, schema, defaults, _, _ := parseExtraFields(ddl)

	require.Contains(t, fields, "SourceLineKey string")
	require.Contains(t, fields, "PageNumber *int")
	require.NotContains(t, fields, "Comment")
	require.Equal(t, `"source_line_key", "page_number"`, columns)
	require.Equal(t, "m.SourceLineKey, m.PageNumber", values)
	require.NotContains(t, schema, `"COMMENT"`)
	require.Equal(t, `"page_number"`, defaults)
}

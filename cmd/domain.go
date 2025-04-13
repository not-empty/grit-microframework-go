package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

type DomainData struct {
	Domain      string
	DomainLower string
	TableName   string
	Fields      string
	Columns     string
	Values      string
	Sanitize    string
	Schema      string
	HasSanitize bool
}

func Capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func extractTableName(ddl string) string {
	ddl = strings.TrimSpace(ddl)
	ddlLower := strings.ToLower(ddl)
	if !strings.HasPrefix(ddlLower, "create table") {
		return ""
	}
	rest := strings.TrimSpace(ddl[12:])
	end := strings.IndexAny(rest, " (")
	if end == -1 {
		return strings.Trim(rest, "`\"")
	}
	return strings.Trim(rest[:end], "`\"")
}

func parseExtraFields(ddl string) (fields, columns, values, sanitize, schema string, hasSanitize bool) {
	lines := strings.Split(ddl, "\n")
	var fieldLines, colNames, valNames, sanitizeLines []string
	schemaMap := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		upperLine := strings.ToUpper(line)
		if line == "" || strings.HasPrefix(line, ")") ||
			strings.HasPrefix(upperLine, "PRIMARY") ||
			strings.HasPrefix(upperLine, "KEY") ||
			strings.HasPrefix(upperLine, "CREATE TABLE") {
			continue
		}

		line = strings.TrimRight(line, ",")
		tokens := strings.Fields(line)
		if len(tokens) < 2 {
			continue
		}

		colName := strings.Trim(tokens[0], "`")
		if colName == "id" || colName == "created_at" || colName == "updated_at" || colName == "deleted_at" {
			continue
		}

		sqlType := strings.ToLower(tokens[1])
		var goType string
		switch {
		case strings.Contains(sqlType, "int"):
			goType = "int"
		case strings.Contains(sqlType, "char"), strings.Contains(sqlType, "text"):
			goType = "string"
		case strings.Contains(sqlType, "bool"):
			goType = "bool"
		case strings.Contains(sqlType, "date"), strings.Contains(sqlType, "time"):
			goType = "*time.Time"
		default:
			goType = "string"
		}

		fieldName := strings.Title(colName)
		fieldLines = append(fieldLines, fmt.Sprintf("%s %s `json:\"%s\"`", fieldName, goType, colName))
		colNames = append(colNames, fmt.Sprintf("\"%s\"", colName))
		valNames = append(valNames, fmt.Sprintf("m.%s", fieldName))
		schemaMap = append(schemaMap, fmt.Sprintf("\"%s\": \"%s\"", colName, goType))

		if strings.Contains(line, "-- sanitize-html") {
			sanitizeLines = append(sanitizeLines, fmt.Sprintf("m.%s = policy.Sanitize(m.%s)", fieldName, fieldName))
			hasSanitize = true
		}
	}

	fields = strings.Join(fieldLines, "\n\t")
	columns = strings.Join(colNames, ", ")
	values = strings.Join(valNames, ", ")
	schema = strings.Join(schemaMap, ",\n\t\t")

	if hasSanitize {
		sanitize = "policy := bluemonday.UGCPolicy()\n\t" + strings.Join(sanitizeLines, "\n\t")
	} else {
		sanitize = "// no fields to sanitize"
	}

	return
}

func main() {
	domainPtr := flag.String("domain", "", "Name of the domain (e.g., user, role)")
	flag.Parse()
	if *domainPtr == "" {
		log.Fatal("Domain name is required. Use -domain flag.")
	}
	domainLower := strings.ToLower(*domainPtr)
	domainCap := Capitalize(domainLower)

	ddlPath := filepath.Join("sql", domainLower+".sql")
	ddlData, err := ioutil.ReadFile(ddlPath)
	if err != nil {
		log.Fatalf("Error reading DDL file %s: %v", ddlPath, err)
	}
	ddlContent := string(ddlData)
	tableName := extractTableName(ddlContent)
	if tableName == "" {
		log.Fatalf("Could not extract table name from DDL")
	}

	extraField, extraColumn, extraValue, sanitize, schema, hasSanitize := parseExtraFields(ddlContent)

	data := DomainData{
		Domain:      domainCap,
		DomainLower: domainLower,
		TableName:   tableName,
		Fields:      extraField,
		Columns:     extraColumn,
		Values:      extraValue,
		Sanitize:    sanitize,
		Schema:      schema,
		HasSanitize: hasSanitize,
	}

	modelStubPath := filepath.Join("stubs", "model.stub")
	routesStubPath := filepath.Join("stubs", "domain.stub")

	modelStubBytes, err := ioutil.ReadFile(modelStubPath)
	if err != nil {
		log.Fatalf("Error reading model stub: %v", err)
	}
	routesStubBytes, err := ioutil.ReadFile(routesStubPath)
	if err != nil {
		log.Fatalf("Error reading domain stub: %v", err)
	}

	modelTmpl, err := template.New("model").Parse(string(modelStubBytes))
	if err != nil {
		log.Fatalf("Error parsing model stub: %v", err)
	}
	routesTmpl, err := template.New("routes").Parse(string(routesStubBytes))
	if err != nil {
		log.Fatalf("Error parsing domain stub: %v", err)
	}

	modelOutPath := filepath.Join("..", "app", "repository", "models", domainLower+"_model.go")
	routesOutPath := filepath.Join("..", "app", "router", "domains", domainLower+"_domain.go")

	modelFile, err := os.Create(modelOutPath)
	if err != nil {
		log.Fatalf("Error creating model output file: %v", err)
	}
	defer modelFile.Close()

	routesFile, err := os.Create(routesOutPath)
	if err != nil {
		log.Fatalf("Error creating routes output file: %v", err)
	}
	defer routesFile.Close()

	if err := modelTmpl.Execute(modelFile, data); err != nil {
		log.Fatalf("Error executing model template: %v", err)
	}
	if err := routesTmpl.Execute(routesFile, data); err != nil {
		log.Fatalf("Error executing routes template: %v", err)
	}

	if err := os.Chown(modelOutPath, 1000, 1000); err != nil {
		log.Fatalf("Error changing file ownership for model: %v", err)
	}
	if err := os.Chown(routesOutPath, 1000, 1000); err != nil {
		log.Fatalf("Error changing file ownership for routes: %v", err)
	}

	fmt.Printf("Generated domain files:\n - %s\n - %s\n", modelOutPath, routesOutPath)
}

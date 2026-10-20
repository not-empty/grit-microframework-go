package nullable_integer

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/not-empty/grit-microframework-go/app/config"
	"github.com/not-empty/grit-microframework-go/app/controller"
	"github.com/not-empty/grit-microframework-go/app/database"
	"github.com/not-empty/grit-microframework-go/app/middleware"
	"github.com/not-empty/grit-microframework-go/app/repository"
	"github.com/stretchr/testify/require"
)

var featureTableName = fmt.Sprintf("nullable_integer_feature_%d", os.Getpid())

type nullableIntegerModel struct {
	ID        string     `json:"id"`
	Label     string     `json:"label"`
	Score     *int       `json:"score"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (m *nullableIntegerModel) Schema() map[string]string {
	return map[string]string{
		"id":         "string",
		"label":      "string",
		"score":      "int",
		"created_at": "*time.Time",
		"updated_at": "*time.Time",
		"deleted_at": "*time.Time",
	}
}

func (m *nullableIntegerModel) TableName() string { return "`" + featureTableName + "`" }

func (m *nullableIntegerModel) Columns() []string {
	return []string{"id", "label", "score", "created_at", "updated_at", "deleted_at"}
}

func (m *nullableIntegerModel) Values() []interface{} {
	return []interface{}{m.ID, m.Label, m.Score, m.CreatedAt, m.UpdatedAt, m.DeletedAt}
}

func (m *nullableIntegerModel) HasDefaultValue() []string { return nil }
func (m *nullableIntegerModel) PrimaryKey() string        { return "id" }
func (m *nullableIntegerModel) PrimaryKeyValue() interface{} {
	return m.ID
}
func (m *nullableIntegerModel) SetCreatedAt(value time.Time) { m.CreatedAt = &value }
func (m *nullableIntegerModel) SetUpdatedAt(value time.Time) { m.UpdatedAt = &value }

func TestNullableIntegerIsDistinctFromZeroAcrossRealEndpoints(t *testing.T) {
	require.NoError(t, godotenv.Load("../../../.env"))
	config.LoadConfig()
	db := database.Init(database.LoadDatabaseConfig())
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, err := db.Exec("DROP TABLE IF EXISTS `" + featureTableName + "`")
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE ` + featureTableName + ` (
			id CHAR(26) NOT NULL,
			label VARCHAR(32) NOT NULL,
			score INT NULL,
			created_at DATETIME DEFAULT NULL,
			updated_at DATETIME DEFAULT NULL,
			deleted_at DATETIME DEFAULT NULL,
			PRIMARY KEY (id),
			KEY idx_nullable_integer_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, cleanupErr := db.Exec("DROP TABLE IF EXISTS `" + featureTableName + "`")
		require.NoError(t, cleanupErr)
	})

	repo := repository.NewRepository(db, func() *nullableIntegerModel { return &nullableIntegerModel{} })
	ctrl := controller.NewBaseController(repo, "/nullable", func(m *nullableIntegerModel, id string) {
		m.ID = id
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/nullable/add", ctrl.Add)
	mux.Handle("/nullable/bulk", middleware.ResponseMiddleware(http.HandlerFunc(ctrl.Bulk)))
	mux.HandleFunc("/nullable/bulk_add", ctrl.BulkAdd)
	mux.HandleFunc("/nullable/dead_detail/", ctrl.DeadDetail)
	mux.HandleFunc("/nullable/dead_list", ctrl.DeadList)
	mux.HandleFunc("/nullable/delete/", ctrl.Delete)
	mux.HandleFunc("/nullable/detail/", ctrl.Detail)
	mux.HandleFunc("/nullable/edit/", ctrl.Edit)
	mux.Handle("/nullable/list", middleware.ResponseMiddleware(http.HandlerFunc(ctrl.List)))
	mux.HandleFunc("/nullable/list_one", ctrl.ListOne)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	nullID := addRecord(t, server.URL, map[string]interface{}{"label": "omitted"})
	zeroID := addRecord(t, server.URL, map[string]interface{}{"label": "zero", "score": 0})
	explicitNullID := addRecord(t, server.URL, map[string]interface{}{"label": "explicit-null", "score": nil})

	assertDatabaseScore(t, db, nullID, false, 0)
	assertDatabaseScore(t, db, zeroID, true, 0)
	assertDatabaseScore(t, db, explicitNullID, false, 0)

	assertEndpointScore(t, server.URL+"/nullable/detail/"+nullID, nil)
	assertEndpointScore(t, server.URL+"/nullable/detail/"+zeroID, float64(0))
	assertEndpointScore(t, server.URL+"/nullable/list_one?filter=score:nul:true", nil)

	list := getList(t, server.URL+"/nullable/list?limit=10&order_by=label&order=asc")
	assertScoreByID(t, list, nullID, nil)
	assertScoreByID(t, list, zeroID, float64(0))
	orderedByScore := getList(t, server.URL+"/nullable/list?limit=10&order_by=score&order=asc")
	assertScoreByID(t, orderedByScore, nullID, nil)
	assertScoreByID(t, orderedByScore, zeroID, float64(0))

	bulk := requestList(t, http.MethodPost, server.URL+"/nullable/bulk", map[string]interface{}{
		"ids": []string{nullID, zeroID},
	})
	assertScoreByID(t, bulk, nullID, nil)
	assertScoreByID(t, bulk, zeroID, float64(0))

	patch(t, server.URL+"/nullable/edit/"+zeroID, map[string]interface{}{"score": nil})
	assertDatabaseScore(t, db, zeroID, false, 0)
	assertEndpointScore(t, server.URL+"/nullable/detail/"+zeroID, nil)

	patch(t, server.URL+"/nullable/edit/"+nullID, map[string]interface{}{"score": 0})
	assertDatabaseScore(t, db, nullID, true, 0)
	assertEndpointScore(t, server.URL+"/nullable/detail/"+nullID, float64(0))

	bulkAddResponse := requestMap(t, http.MethodPost, server.URL+"/nullable/bulk_add", []map[string]interface{}{
		{"label": "bulk-null"},
		{"label": "bulk-zero", "score": 0},
	})
	bulkIDs, ok := bulkAddResponse["ids"].([]interface{})
	require.True(t, ok)
	require.Len(t, bulkIDs, 2)
	assertDatabaseScore(t, db, bulkIDs[0].(string), false, 0)
	assertDatabaseScore(t, db, bulkIDs[1].(string), true, 0)

	requestNoContent(t, http.MethodDelete, server.URL+"/nullable/delete/"+zeroID, nil)
	assertEndpointScore(t, server.URL+"/nullable/dead_detail/"+zeroID, nil)
	deadList := getList(t, server.URL+"/nullable/dead_list?limit=10")
	assertScoreByID(t, deadList, zeroID, nil)

	assertNullableScorePagination(t, db, server.URL, "asc")
	assertNullableScorePagination(t, db, server.URL, "desc")
}

func addRecord(t *testing.T, baseURL string, payload map[string]interface{}) string {
	t.Helper()
	response := requestMap(t, http.MethodPost, baseURL+"/nullable/add", payload)
	id, ok := response["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, id)
	return id
}

func patch(t *testing.T, url string, payload map[string]interface{}) {
	t.Helper()
	requestNoContent(t, http.MethodPatch, url, payload)
}

func requestNoContent(t *testing.T, method, url string, payload interface{}) {
	t.Helper()
	response := doRequest(t, method, url, payload)
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d: %s", response.StatusCode, readBody(t, response))
	}
}

func requestMap(t *testing.T, method, url string, payload interface{}) map[string]interface{} {
	t.Helper()
	response := doRequest(t, method, url, payload)
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		t.Fatalf("status %d: %s", response.StatusCode, readBody(t, response))
	}
	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	return body
}

func requestList(t *testing.T, method, url string, payload interface{}) []map[string]interface{} {
	t.Helper()
	response := doRequest(t, method, url, payload)
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		t.Fatalf("status %d: %s", response.StatusCode, readBody(t, response))
	}
	var body []map[string]interface{}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	return body
}

func doRequest(t *testing.T, method, url string, payload interface{}) *http.Response {
	t.Helper()
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		require.NoError(t, err)
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, url, body)
	require.NoError(t, err)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	return response
}

func getList(t *testing.T, url string) []map[string]interface{} {
	t.Helper()
	return requestList(t, http.MethodGet, url, nil)
}

func assertEndpointScore(t *testing.T, url string, expected interface{}) {
	t.Helper()
	response := requestMap(t, http.MethodGet, url, nil)
	require.Equal(t, expected, response["score"])
}

func assertScoreByID(t *testing.T, records []map[string]interface{}, id string, expected interface{}) {
	t.Helper()
	for _, record := range records {
		if record["id"] == id {
			require.Equal(t, expected, record["score"])
			return
		}
	}
	t.Fatalf("record %s was not returned", id)
}

func assertDatabaseScore(t *testing.T, db *sql.DB, id string, valid bool, expected int64) {
	t.Helper()
	var score sql.NullInt64
	require.NoError(t, db.QueryRow("SELECT score FROM `"+featureTableName+"` WHERE id = ?", id).Scan(&score))
	require.Equal(t, valid, score.Valid)
	if valid {
		require.Equal(t, expected, score.Int64)
	}
}

func assertNullableScorePagination(t *testing.T, db *sql.DB, baseURL, order string) {
	t.Helper()

	var expectedCount int
	require.NoError(t, db.QueryRow(
		"SELECT COUNT(*) FROM `"+featureTableName+"` WHERE deleted_at IS NULL",
	).Scan(&expectedCount))

	endpoint := baseURL + "/nullable/list?limit=1&order_by=score&order=" + order
	seen := make(map[string]struct{}, expectedCount)
	nullCount := 0
	zeroCount := 0

	for page := 0; page <= expectedCount; page++ {
		response := doRequest(t, http.MethodGet, endpoint, nil)
		require.Equal(t, http.StatusOK, response.StatusCode)

		var records []map[string]interface{}
		require.NoError(t, json.NewDecoder(response.Body).Decode(&records))
		require.NoError(t, response.Body.Close())
		require.LessOrEqual(t, len(records), 1)

		for _, record := range records {
			id, ok := record["id"].(string)
			require.True(t, ok)
			if _, duplicate := seen[id]; duplicate {
				t.Fatalf("record %s was repeated during %s nullable pagination", id, order)
			}
			seen[id] = struct{}{}
			if record["score"] == nil {
				nullCount++
			} else if record["score"] == float64(0) {
				zeroCount++
			}
		}

		cursor := response.Header.Get("X-Page-Cursor")
		if cursor == "" {
			break
		}
		endpoint = baseURL + "/nullable/list?limit=1&order_by=score&order=" + order +
			"&page_cursor=" + url.QueryEscape(cursor)
	}

	require.Len(t, seen, expectedCount)
	require.Positive(t, nullCount)
	require.Positive(t, zeroCount)
}

func readBody(t *testing.T, response *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	return string(body)
}

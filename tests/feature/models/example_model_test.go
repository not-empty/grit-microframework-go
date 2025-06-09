package feature

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"github.com/not-empty/grit/app"
	"github.com/not-empty/grit/app/repository/models"
)

var baseURL string

func TestMain(m *testing.M) {
	wd, _ := os.Getwd()
	println("CURRENT WORK DIR:", wd)
	_ = godotenv.Load("../../../.env")
	testAppPort := "8002"
	os.Setenv("APP_PORT", testAppPort)
	os.Setenv("APP_NO_AUTH", "true")
	baseURL = "http://localhost:" + testAppPort

	app.Bootstrap()

	go app.StartServer()

	time.Sleep(1 * time.Second)

	code := m.Run()
	os.Exit(code)
}

func TestExampleEndpoints(t *testing.T) {
	var insertedID string

	t.Run("Add - Success", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Example User",
			"age":  30,
		}
		payload, _ := json.Marshal(body)

		resp, err := http.Post(baseURL+"/example/add", "application/json", bytes.NewBuffer(payload))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var respBody map[string]string
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.NoError(t, err)
		insertedID = respBody["id"]
		assert.NotEmpty(t, insertedID)
	})

	t.Run("Add - Fail Validation", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Ex",
			"age":  -5,
		}
		payload, _ := json.Marshal(body)

		resp, err := http.Post(baseURL+"/example/add", "application/json", bytes.NewBuffer(payload))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("Add - Method Not Allowed", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/example/add")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("Add - Invalid Json", func(t *testing.T) {
		invalidJson := []byte(`{"test":}`)

		resp, err := http.Post(baseURL+"/example/add", "application/json", bytes.NewBuffer(invalidJson))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Detail - Success", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/example/detail/" + insertedID)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assert.NoError(t, err)
		assert.Equal(t, "Example User", data["name"])
	})

	t.Run("Edit - Success", func(t *testing.T) {
		body := map[string]interface{}{
			"age": 40,
		}
		payload, _ := json.Marshal(body)

		req, _ := http.NewRequest(http.MethodPatch, baseURL+"/example/edit/"+insertedID, bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("List - Success", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/example/list")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var data []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("Bulk - Success", func(t *testing.T) {
		body := map[string]interface{}{
			"ids": []string{insertedID},
		}
		payload, _ := json.Marshal(body)

		resp, err := http.Post(baseURL+"/example/bulk", "application/json", bytes.NewBuffer(payload))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var data []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assert.NoError(t, err)
		assert.Len(t, data, 1)
	})

	t.Run("ListOne - List one", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/example/list_one")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assert.NoError(t, err)
		assert.Equal(t, "Example User", data["name"])
	})

	t.Run("Delete - Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, baseURL+"/example/delete/"+insertedID, nil)
		client := &http.Client{}
		resp, err := client.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("DeadDetail - After Delete", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/example/dead_detail/" + insertedID)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assert.NoError(t, err)
		assert.Equal(t, "Example User", data["name"])
	})

	t.Run("DeadList - List deleted", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/example/dead_list")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var data []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("Bulk Add - Success", func(t *testing.T) {
		body := []models.Example{
			{
				Name: "Example User",
				Age:  30,
			},
			{
				Name: "Example User 2",
				Age:  31,
			},
		}

		payload, _ := json.Marshal(body)

		resp, err := http.Post(baseURL+"/example/bulk_add", "application/json", bytes.NewBuffer(payload))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var respBody map[string][]string
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		assert.NoError(t, err)
		insertedIDs := respBody["ids"]

		assert.NotEmpty(t, insertedIDs)
		assert.Equal(t, 2, len(insertedIDs))
	})

	t.Run("Bulk Add - Fail Validation", func(t *testing.T) {
		body := []models.Example{
			{
				Name: "Ex",
				Age:  -5,
			},
			{
				Name: "Example User 2",
				Age:  31,
			},
		}

		payload, _ := json.Marshal(body)

		resp, err := http.Post(baseURL+"/example/bulk_add", "application/json", bytes.NewBuffer(payload))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("Bulk Add - Invalid Elements Quantity", func(t *testing.T) {
		body := []models.Example{}

		payload, _ := json.Marshal(body)

		resp, err := http.Post(baseURL+"/example/bulk_add", "application/json", bytes.NewBuffer(payload))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var respBody map[string]string
		json.NewDecoder(resp.Body).Decode(&respBody)

		errString, ok := respBody["error"]

		assert.True(t, ok)
		assert.Equal(t, "Payload must contain between 1 and 25 items", errString)
	})

	t.Run("Bulk Add - Method Not Allowed", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/example/bulk_add")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("Bulk Add - Invalid Json", func(t *testing.T) {
		invalidJson := []byte(`{"test":}`)

		resp, err := http.Post(baseURL+"/example/bulk_add", "application/json", bytes.NewBuffer(invalidJson))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Raw - Success", func(t *testing.T) {
		var body struct {
			Query  string
			Params map[string]any
		}

		body.Query = "count"
		payload, _ := json.Marshal(body)

		resp, err := http.Post(baseURL+"/example/select_raw", "application/json", bytes.NewBuffer(payload))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Raw - Invalid query", func(t *testing.T) {
		var body struct {
			Query  string
			Params map[string]any
		}

		body.Query = "test"
		payload, _ := json.Marshal(body)

		resp, err := http.Post(baseURL+"/example/select_raw", "application/json", bytes.NewBuffer(payload))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

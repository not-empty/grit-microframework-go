package controller

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    "io"
    "errors"

    "github.com/not-empty/grit/app/controller"
    "github.com/not-empty/grit/app/helper"
    "github.com/stretchr/testify/require"
)

type fakeModel struct {
    ID    string `json:"id"`
    Field string `json:"field" validate:"required"`
    CreatedAt *time.Time `json:"created_at"`
    UpdatedAt *time.Time `json:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at"`
}

func (m *fakeModel) TableName() string {
    return "fake"
}

func (m *fakeModel) Columns() []string {
    return []string{"id", "field"}
}

func (m *fakeModel) Values() []interface{} {
    return []interface{}{m.ID, m.Field}
}

func (m *fakeModel) PrimaryKey() string {
    return "id"
}

func (m *fakeModel) PrimaryKeyValue() interface{} {
    return m.ID
}

func (m *fakeModel) SetCreatedAt(t time.Time) {
    m.CreatedAt = &t
}

func (m *fakeModel) SetUpdatedAt(t time.Time) {
    m.UpdatedAt = &t
}


func (m *fakeModel) Schema() map[string]string {
    return map[string]string{"id": "string", "field": "string"}
}

type fakeRepository struct {
    insertedModel *fakeModel
    insertedError error

    updateFieldsCalled bool
    updateFieldsCols   []string
    updateFieldsVals   []interface{}
    updateFieldsError  error

    deleteCalled bool
    deleteError  error

    getResult map[string]any
    getError  error

    getDeletedResult map[string]any
    getDeletedError  error

    listActiveResult []map[string]any
    listActiveError  error

    listDeletedResult []map[string]any
    listDeletedError  error

    bulkGetResult []map[string]any
    bulkGetError  error
}

func (fr *fakeRepository) New() *fakeModel {
    return &fakeModel{}
}

func (fr *fakeRepository) Insert(m *fakeModel) error {
    fr.insertedModel = m
    return fr.insertedError
}

func (fr *fakeRepository) UpdateFields(table, pk string, pkVal interface{}, cols []string, vals []interface{}) error {
    fr.updateFieldsCalled = true
    fr.updateFieldsCols = cols
    fr.updateFieldsVals = vals
    return fr.updateFieldsError
}

func (fr *fakeRepository) Delete(m *fakeModel) error {
    fr.deleteCalled = true
    return fr.deleteError
}

func (fr *fakeRepository) Get(id interface{}, fields []string) (map[string]any, error) {
    return fr.getResult, fr.getError
}

func (fr *fakeRepository) GetDeleted(id interface{}, fields []string) (map[string]any, error) {
    return fr.getDeletedResult, fr.getDeletedError
}

func (fr *fakeRepository) ListActive(limit int , pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
    return fr.listActiveResult, fr.listActiveError
}

func (fr *fakeRepository) ListDeleted(limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
    return fr.listDeletedResult, fr.listDeletedError
}

func (fr *fakeRepository) BulkGet(ids []string, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string) ([]map[string]any, error) {
    return fr.bulkGetResult, fr.bulkGetError
}

type fakeULIDGenerator struct{}
func (f *fakeULIDGenerator) IsValidFormat(ulidStr string) bool {
    return true
}
func (f *fakeULIDGenerator) GetTimeFromUlid(ulidStr string) (int64, error) {
    return 0, nil
}
func (f *fakeULIDGenerator) GetDateFromUlid(ulidStr string) (string, error) {
    return "2020-01-01 00:00:00", nil
}
func (f *fakeULIDGenerator) GetRandomnessFromString(ulidStr string) (string, error) {
    return "", nil
}
func (f *fakeULIDGenerator) IsDuplicatedTime(t int64) bool {
    return false
}
func (f *fakeULIDGenerator) HasIncrementLastRandChars(duplicateTime bool) bool {
    return false
}
func (f *fakeULIDGenerator) Generate(t int64) (string, error) {
    return "fake-ulid", nil
}
func (f *fakeULIDGenerator) DecodeTime(timePart string) (int64, error) {
    return 0, nil
}

type errorULIDGen struct{}

func (e *errorULIDGen) IsValidFormat(ulidStr string) bool {
    return true
}
func (e *errorULIDGen) GetTimeFromUlid(ulidStr string) (int64, error) {
    return 0, nil
}
func (e *errorULIDGen) GetDateFromUlid(ulidStr string) (string, error) {
    return "", nil
}
func (e *errorULIDGen) GetRandomnessFromString(ulidStr string) (string, error) {
    return "", nil
}
func (e *errorULIDGen) IsDuplicatedTime(t int64) bool {
    return false
}
func (e *errorULIDGen) HasIncrementLastRandChars(duplicateTime bool) bool {
    return false
}
func (e *errorULIDGen) Generate(t int64) (string, error) {
    return "", errors.New("ULID error")
}
func (e *errorULIDGen) DecodeTime(timePart string) (int64, error) {
    return 0, nil
}

func TestNewBaseController(t *testing.T) {
    fr := &fakeRepository{}
    
    setPK := func(m *fakeModel, id string) {
        m.ID = id
    }
    
    prefix := "/fake"
    
    bc := controller.NewBaseController[*fakeModel](fr, prefix, setPK)
    
    require.NotNil(t, bc, "BaseController should not be nil")
    require.Equal(t, fr, bc.Repo, "Repository should be set")
    require.Equal(t, prefix, bc.Prefix, "Prefix should match the provided value")
    
    model := &fakeModel{}
    bc.SetPK(model, "123")
    require.Equal(t, "123", model.ID, "SetPK should set the model's ID")
}

func TestBaseController_Bulk(t *testing.T) {
    fr := &fakeRepository{
        bulkGetResult: []map[string]any{
            {"id": "1", "field": "bulkValue1"},
            {"id": "2", "field": "bulkValue2"},
        },
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) { m.ID = id },
    }

    payload := map[string]interface{}{
        "ids": []string{"1", "2"},
    }
    body, err := json.Marshal(payload)
    require.NoError(t, err)

    req := httptest.NewRequest(http.MethodPost, "/fake/bulk", bytes.NewBuffer(body))
    rr := httptest.NewRecorder()

    bc.Bulk(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusOK, res.StatusCode)

    var responseBody []map[string]any
    err = json.NewDecoder(res.Body).Decode(&responseBody)
    require.NoError(t, err)
    require.Len(t, responseBody, 2)
    require.Equal(t, "bulkValue1", responseBody[0]["field"])
}

func TestBaseController_DeadDetail(t *testing.T) {
    fr := &fakeRepository{
        getDeletedResult: map[string]any{"id": "1", "field": "deadDetail"},
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) { m.ID = id },
    }
    req := httptest.NewRequest(http.MethodGet, "/fake/dead_detail/1", nil)
    rr := httptest.NewRecorder()

    bc.DeadDetail(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusOK, res.StatusCode)

    var responseBody map[string]any
    err := json.NewDecoder(res.Body).Decode(&responseBody)
    require.NoError(t, err)
    require.Equal(t, "deadDetail", responseBody["field"])
}

func TestBaseController_DeadList(t *testing.T) {
    fr := &fakeRepository{
        listDeletedResult: []map[string]any{
            {"id": "1", "field": "deadListValue"},
        },
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) { m.ID = id },
    }

    req := httptest.NewRequest(http.MethodGet, "/fake/dead_list", nil)
    rr := httptest.NewRecorder()

    bc.DeadList(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusOK, res.StatusCode)

    var responseBody []map[string]any
    err := json.NewDecoder(res.Body).Decode(&responseBody)
    require.NoError(t, err)
    require.Len(t, responseBody, 1)
    require.Equal(t, "deadListValue", responseBody[0]["field"])
}

func TestBaseController_Delete(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) { m.ID = id },
    }
    req := httptest.NewRequest(http.MethodDelete, "/fake/delete/1", nil)
    rr := httptest.NewRecorder()

    bc.Delete(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusNoContent, res.StatusCode)
    require.True(t, fr.deleteCalled, "Expected Delete to be called in repository")
}

func TestBaseController_Detail(t *testing.T) {
    fr := &fakeRepository{
        getResult: map[string]any{"id": "1", "field": "detailValue"},
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) { m.ID = id },
    }
    req := httptest.NewRequest(http.MethodGet, "/fake/detail/1", nil)
    rr := httptest.NewRecorder()

    bc.Detail(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusOK, res.StatusCode)

    var responseBody map[string]any
    err := json.NewDecoder(res.Body).Decode(&responseBody)
    require.NoError(t, err)
    require.Equal(t, "detailValue", responseBody["field"])
}

func TestBaseController_Edit(t *testing.T) {
    initialRecord := map[string]any{"id": "1", "field": "oldValue"}
    fr := &fakeRepository{
        getResult: initialRecord,
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) { m.ID = id },
    }
    patchData := map[string]interface{}{
        "field": "newValue",
    }
    patchBytes, err := json.Marshal(patchData)
    require.NoError(t, err)
    req := httptest.NewRequest(http.MethodPatch, "/fake/edit/1", bytes.NewBuffer(patchBytes))
    rr := httptest.NewRecorder()

    bc.Edit(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusNoContent, res.StatusCode)
    require.True(t, fr.updateFieldsCalled, "Expected UpdateFields to be called")
    found := false
    for _, col := range fr.updateFieldsCols {
        if col == "field" {
            found = true
            break
        }
    }
    require.True(t, found, "Expected 'field' to be updated")
}

func TestBaseController_List(t *testing.T) {
    fr := &fakeRepository{
        listActiveResult: []map[string]any{
            {"id": "1", "field": "listValue"},
        },
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) { m.ID = id },
    }

    req := httptest.NewRequest(http.MethodGet, "/fake/list", nil)
    rr := httptest.NewRecorder()

    bc.List(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusOK, res.StatusCode)

    var responseBody []map[string]any
    err := json.NewDecoder(res.Body).Decode(&responseBody)
    require.NoError(t, err)
    require.Len(t, responseBody, 1)
    require.Equal(t, "listValue", responseBody[0]["field"])
}

func TestBaseController_Add(t *testing.T) {
    fr := &fakeRepository{}

    setPK := func(m *fakeModel, id string) {
        m.ID = id
    }

    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK:  setPK,
        ULIDGen: &fakeULIDGenerator{},
    }

    payload := map[string]string{"field": "test value"}
    payloadBytes, err := json.Marshal(payload)
    require.NoError(t, err)

    req := httptest.NewRequest(http.MethodPost, "/fake/add", bytes.NewBuffer(payloadBytes))
    rr := httptest.NewRecorder()

    bc.Add(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusCreated, res.StatusCode, "Expected status 201 Created")

    var respBody map[string]string
    err = json.NewDecoder(res.Body).Decode(&respBody)
    require.NoError(t, err)
    id, ok := respBody["id"]
    require.True(t, ok, "Response should contain an 'id' field")
    require.NotEmpty(t, id, "The 'id' field should not be empty")

    require.NotNil(t, fr.insertedModel, "Expected the model to be inserted")
    require.Equal(t, id, fr.insertedModel.ID, "The inserted model ID should match the returned id")
    require.Equal(t, "test value", fr.insertedModel.Field, "The model field should match the input payload")
}

func TestBaseController_Add_MethodNotAllowed(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) {
            m.ID = id
        },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodGet, "/fake/add", nil)
    rr := httptest.NewRecorder()

    bc.Add(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Method not allowed")
}

func TestBaseController_Add_InvalidJSON(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) {
            m.ID = id
        },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodPost, "/fake/add", bytes.NewBufferString("{"))
    rr := httptest.NewRecorder()

    bc.Add(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusBadRequest, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Invalid JSON")
}

func TestBaseController_Add_InvalidPayload(t *testing.T) {
    fr := &fakeRepository{}
    
    setPK := func(m *fakeModel, id string) {
        m.ID = id
    }

    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK:  setPK,
        ULIDGen: &fakeULIDGenerator{},
    }

    payload := map[string]string{"field": ""}
    payloadBytes, err := json.Marshal(payload)
    require.NoError(t, err)

    req := httptest.NewRequest(http.MethodPost, "/fake/add", bytes.NewBuffer(payloadBytes))
    rr := httptest.NewRecorder()

    bc.Add(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)

    var respJSON map[string]any
    err = json.Unmarshal(body, &respJSON)
    require.NoError(t, err)

    errorsVal, ok := respJSON["errors"]
    require.True(t, ok, "Expected an 'errors' field in the response")
    errSlice, ok := errorsVal.([]any)
    require.True(t, ok, "Expected 'errors' field to be an array")
    require.NotEmpty(t, errSlice, "Expected at least one validation error message")
}

func TestBaseController_Add_InsertError(t *testing.T) {
    fr := &fakeRepository{
        insertedError: errors.New("Insert error"),
    }

    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) {
            m.ID = id
        },
        ULIDGen: &fakeULIDGenerator{},
    }

    payload := map[string]string{"field": "test value"}
    payloadBytes, err := json.Marshal(payload)
    require.NoError(t, err)

    req := httptest.NewRequest(http.MethodPost, "/fake/add", bytes.NewBuffer(payloadBytes))
    rr := httptest.NewRecorder()

    bc.Add(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusInternalServerError, res.StatusCode)

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Insert error")
}

func TestBaseController_Add_ULIDGenerationError(t *testing.T) {
    fr := &fakeRepository{}

    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) {
            m.ID = id
        },
        ULIDGen: &errorULIDGen{},
    }

    payload := map[string]string{"field": "test value"}
    payloadBytes, err := json.Marshal(payload)
    require.NoError(t, err)

    req := httptest.NewRequest(http.MethodPost, "/fake/add", bytes.NewBuffer(payloadBytes))
    rr := httptest.NewRecorder()

    bc.Add(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusInternalServerError, res.StatusCode)

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "ULID error")
}

func TestBaseController_Bulk_MethodNotAllowed(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) {
            m.ID = id
        },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodGet, "/fake/bulk", nil)
    rr := httptest.NewRecorder()

    bc.Bulk(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Method not allowed")
}

func TestBaseController_Bulk_InvalidJSON(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK:  func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodPost, "/fake/bulk", bytes.NewBufferString("{"))
    rr := httptest.NewRecorder()

    bc.Bulk(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusBadRequest, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Invalid or empty Ids list")
}


func TestBaseController_Bulk_EmptyIDs(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK:  func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodPost, "/fake/bulk", bytes.NewBufferString(`{"ids": []}`))
    rr := httptest.NewRecorder()

    bc.Bulk(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusBadRequest, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Invalid or empty Ids list")
}

func TestBaseController_Bulk_BulkGetError(t *testing.T) {
    fr := &fakeRepository{
        bulkGetError: errors.New("bulk error"),
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK:  func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodPost, "/fake/bulk", bytes.NewBufferString(`{"ids": ["1", "2"]}`))
    rr := httptest.NewRecorder()

    bc.Bulk(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusInternalServerError, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Bulk error")
}

func TestDeadDetail_MethodNotAllowed(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) {
            m.ID = id
        },
    }
    req := httptest.NewRequest(http.MethodPost, "/fake/dead_detail/1", nil)
    rr := httptest.NewRecorder()

    bc.DeadDetail(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Method not allowed")
}

func TestDeadDetail_MissingId(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) {
            m.ID = id
        },
    }
    req := httptest.NewRequest(http.MethodGet, "/fake/dead_detail/", nil)
    rr := httptest.NewRecorder()

    bc.DeadDetail(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusBadRequest, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Missing Id")
}

func TestDeadDetail_GetDeletedError(t *testing.T) {
    fr := &fakeRepository{
        getDeletedError: errors.New("get deleted error"),
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) {
            m.ID = id
        },
    }
    req := httptest.NewRequest(http.MethodGet, "/fake/dead_detail/1", nil)
    rr := httptest.NewRecorder()

    bc.DeadDetail(rr, req)

    res := rr.Result()

    require.Equal(t, http.StatusNotFound, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Fields error")
}

func TestBaseController_DeadList_MethodNotAllowed(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK:  func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodPost, "/fake/dead_list", nil)
    rr := httptest.NewRecorder()
    bc.DeadList(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
    body, _ := io.ReadAll(res.Body)
    require.Contains(t, string(body), "Method not allowed")
}

func TestDeadList_ListError(t *testing.T) {
    fr := &fakeRepository{
        listDeletedError: errors.New("list error"),
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodGet, "/fake/dead_list", nil)
    rr := httptest.NewRecorder()
    bc.DeadList(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusInternalServerError, res.StatusCode)
    body, _ := io.ReadAll(res.Body)
    require.Contains(t, string(body), "List error")
}

func TestBaseController_Delete_MethodNotAllowed(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) {
            m.ID = id
        },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodGet, "/fake/delete/1", nil)
    rr := httptest.NewRecorder()

    bc.Delete(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Method not allowed")
}

func TestBaseController_Delete_MissingId(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) {
            m.ID = id
        },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodDelete, "/fake/delete/", nil)
    rr := httptest.NewRecorder()

    bc.Delete(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusBadRequest, res.StatusCode)

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Missing Id")
}

func TestBaseController_Delete_DeleteError(t *testing.T) {
    fr := &fakeRepository{
        deleteError: errors.New("delete error"),
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:   fr,
        Prefix: "/fake",
        SetPK: func(m *fakeModel, id string) {
            m.ID = id
        },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodDelete, "/fake/delete/1", nil)
    rr := httptest.NewRecorder()

    bc.Delete(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusInternalServerError, res.StatusCode)

    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Delete error")
}

func TestBaseController_Detail_MethodNotAllowed(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodPost, "/fake/detail/1", nil)
    rr := httptest.NewRecorder()
    
    bc.Detail(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Method not allowed")
}

func TestBaseController_Detail_MissingId(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }
    req := httptest.NewRequest(http.MethodGet, "/fake/detail/", nil)
    rr := httptest.NewRecorder()
    
    bc.Detail(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusBadRequest, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Missing Id")
}

func TestBaseController_Detail_GetError(t *testing.T) {
    fr := &fakeRepository{
        getError: errors.New("get error"),
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }
    req := httptest.NewRequest(http.MethodGet, "/fake/detail/1", nil)
    rr := httptest.NewRecorder()
    
    bc.Detail(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusNotFound, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Detail error")
}

func TestBaseController_Edit_MethodNotAllowed(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodGet, "/fake/edit/1", nil)
    rr := httptest.NewRecorder()
    
    bc.Edit(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Method not allowed")
}

func TestBaseController_Edit_MissingId(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodPatch, "/fake/edit/", bytes.NewBufferString(`{"field": "newValue"}`))
    rr := httptest.NewRecorder()
    
    bc.Edit(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusBadRequest, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Missing Id")
}

func TestBaseController_Edit_InvalidData(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodPatch, "/fake/edit/1", bytes.NewBufferString("{"))
    rr := httptest.NewRecorder()
    
    bc.Edit(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusBadRequest, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Invalid data")
}

func TestBaseController_Edit_GetError(t *testing.T) {

    fr := &fakeRepository{
        getError: errors.New("get error"),
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodPatch, "/fake/edit/1", bytes.NewBufferString(`{"field": "newValue"}`))
    rr := httptest.NewRecorder()
    
    bc.Edit(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusNotFound, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Not found")
}

func TestBaseController_Edit_UpdateError(t *testing.T) {
    fr := &fakeRepository{
        updateFieldsError: errors.New("update error"),
        getResult: map[string]any{"id": "1", "field": "oldValue"},
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }
    patchData := map[string]interface{}{
        "field": "newValue",
    }
    patchBytes, err := json.Marshal(patchData)
    require.NoError(t, err)
    req := httptest.NewRequest(http.MethodPatch, "/fake/edit/1", bytes.NewBuffer(patchBytes))
    rr := httptest.NewRecorder()
    
    bc.Edit(rr, req)
    res := rr.Result()
    require.Equal(t, http.StatusInternalServerError, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Edit error")
}

func TestBaseController_List_MethodNotAllowed(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }
    req := httptest.NewRequest(http.MethodPost, "/fake/list", nil)
    rr := httptest.NewRecorder()

    bc.List(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Method not allowed")
}


func TestBaseController_List_ListError(t *testing.T) {
    fr := &fakeRepository{
        listActiveError: errors.New("list error"),
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }
    req := httptest.NewRequest(http.MethodGet, "/fake/list", nil)
    rr := httptest.NewRecorder()

    bc.List(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusInternalServerError, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "List error")
}

func TestBaseController_List_InvalidPageCursor(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodGet, "/fake/list?page_cursor=not-base64!", nil)
    rr := httptest.NewRecorder()

    bc.List(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusBadRequest, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Invalid Page Cursor")
}

func TestBaseController_DeadList_InvalidPageCursor(t *testing.T) {
    fr := &fakeRepository{}
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }

    req := httptest.NewRequest(http.MethodGet, "/fake/dead_list?page_cursor=!!!", nil)
    rr := httptest.NewRecorder()

    bc.DeadList(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusBadRequest, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Invalid Page Cursor")
}

func TestBaseController_Bulk_InvalidPageCursor(t *testing.T) {
    fr := &fakeRepository{
        bulkGetResult: []map[string]any{{"id": "1", "field": "foo"}},
    }
    bc := &controller.BaseController[*fakeModel]{
        Repo:    fr,
        Prefix:  "/fake",
        SetPK:   func(m *fakeModel, id string) { m.ID = id },
        ULIDGen: &fakeULIDGenerator{},
    }

    payload := map[string][]string{"ids": {"1"}}
    b, _ := json.Marshal(payload)
    req := httptest.NewRequest(http.MethodPost, "/fake/bulk?page_cursor=xxx!", bytes.NewBuffer(b))
    rr := httptest.NewRecorder()

    bc.Bulk(rr, req)

    res := rr.Result()
    require.Equal(t, http.StatusBadRequest, res.StatusCode)
    body, err := io.ReadAll(res.Body)
    require.NoError(t, err)
    require.Contains(t, string(body), "Invalid Page Cursor")
}
package testhelper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/austyuzhaninov/test-task-org-api/internal/handler"
	"github.com/austyuzhaninov/test-task-org-api/internal/handler/respond"
	"github.com/austyuzhaninov/test-task-org-api/internal/middleware"
	"github.com/austyuzhaninov/test-task-org-api/internal/service"
	"github.com/austyuzhaninov/test-task-org-api/pkg/logger"
)

// Setup собирает весь стек: моки → сервисы → хендлеры → роутер.
type Setup struct {
	DeptRepo *DeptRepoMock
	EmpRepo  *EmpRepoMock
	Router   http.Handler
}

func NewSetup() *Setup {
	log := logger.New()
	resp := respond.New(log)

	deptRepo := NewDeptRepoMock()
	empRepo := NewEmpRepoMock()

	deptSvc := service.NewDepartmentService(deptRepo, empRepo)
	empSvc := service.NewEmployeeService(empRepo, deptRepo)

	deptHandler := handler.NewDepartmentHandler(deptSvc, resp)
	empHandler := handler.NewEmployeeHandler(empSvc, resp)

	logMw := middleware.NewLogger(log)
	router := handler.NewRouter(deptHandler, empHandler, logMw)

	return &Setup{
		DeptRepo: deptRepo,
		EmpRepo:  empRepo,
		Router:   router,
	}
}

// Request выполняет HTTP запрос через httptest и возвращает recorder.
func (s *Setup) Request(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	s.Router.ServeHTTP(rec, req)
	return rec
}

// DecodeResponse десериализует JSON ответ в переданную структуру.
func DecodeResponse(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

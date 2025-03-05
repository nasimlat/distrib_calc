package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nasimlat/distrib_calc/internal/orchestrator"
)

func clearStorages() {
	orchestrator.Expressions = make(map[string]orchestrator.Expression)
	orchestrator.Tasks = make(map[string]orchestrator.Task)
}

func TestAddExpression(t *testing.T) {
	clearStorages()
	recorder := httptest.NewRecorder()
	reqBody := map[string]string{"expression": "2 + 2"}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	orchestrator.AddExpression(recorder, req)

	resp := recorder.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Ожидался статус код %d, получен %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestGetExpressions(t *testing.T) {
	clearStorages()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)

	orchestrator.GetExpressions(recorder, req)

	resp := recorder.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус код %d, получен %d", http.StatusOK, resp.StatusCode)
	}
}

func TestGetExpressionByID(t *testing.T) {
	clearStorages()

	exprID := "test-id"
	orchestrator.Expressions[exprID] = orchestrator.Expression{ID: exprID, Expr: "2 + 2", Status: "pending"}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/"+exprID, nil)

	orchestrator.GetExpressionByID(recorder, req)

	resp := recorder.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус код %d, получен %d", http.StatusOK, resp.StatusCode)
	}
}

func TestGetTask(t *testing.T) {
	clearStorages()
	taskID := "test-task"
	orchestrator.Tasks[taskID] = orchestrator.Task{ID: taskID, Arg1: 2, Arg2: 2, Operation: "+", OperationTime: 1000}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/task", nil)

	orchestrator.GetTask(recorder, req)

	resp := recorder.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус код %d, получен %d", http.StatusOK, resp.StatusCode)
	}
}

func TestSubmitResult(t *testing.T) {
	clearStorages()
	taskID := "test-task"
	orchestrator.Tasks[taskID] = orchestrator.Task{ID: taskID, Arg1: 2, Arg2: 2, Operation: "+", OperationTime: 1000}

	recorder := httptest.NewRecorder()
	reqBody := map[string]interface{}{"id": taskID, "result": 4.0}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/internal/task", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	orchestrator.SubmitResult(recorder, req)

	resp := recorder.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус код %d, получен %d", http.StatusOK, resp.StatusCode)
	}
}

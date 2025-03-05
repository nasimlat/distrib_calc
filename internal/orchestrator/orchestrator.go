package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID            string  `json:"id"`
	ExpressionID  string  `json:"expression_id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type Expression struct {
	ID     string  `json:"id"`
	Expr   string  `json:"expression"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

var (
	Expressions = make(map[string]Expression)
	Tasks       = make(map[string]Task)
	mu          sync.RWMutex
)

func ClearStorages() {
    mu.Lock()
    defer mu.Unlock()
    Expressions = make(map[string]Expression)
    Tasks = make(map[string]Task)
}

func AddTestExpression(id string, expr Expression) {
    mu.Lock()
    defer mu.Unlock()
    Expressions[id] = expr
}

func AddTestTask(id string, task Task) {
    mu.Lock()
    defer mu.Unlock()
    Tasks[id] = task
}
func Start() {
	http.HandleFunc("/api/v1/calculate", AddExpression)
	http.HandleFunc("/api/v1/expressions", GetExpressions)
	http.HandleFunc("/api/v1/expressions/", GetExpressionByID)
	http.HandleFunc("/internal/task", internalTaskHandler)

	fmt.Println("Оркестратор запущен на :8080")
	http.ListenAndServe(":8080", nil)
}

func AddExpression(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Невалидные данные", http.StatusUnprocessableEntity)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	id := fmt.Sprintf("expr-%d", time.Now().UnixNano())
	Expressions[id] = Expression{
		ID:     id,
		Expr:   req.Expression,
		Status: "pending",
	}

	taskID := fmt.Sprintf("task-%d", time.Now().UnixNano())
	Tasks[taskID] = Task{
		ID:           taskID,
		ExpressionID: id,
		Arg1:         2,
		Arg2:         2,
		Operation:    "*",
		OperationTime: 1000,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func GetExpressions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	var exprs []Expression
	for _, expr := range Expressions {
		exprs = append(exprs, expr)
	}

	json.NewEncoder(w).Encode(map[string][]Expression{"expressions": exprs})
}

func GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")

	mu.RLock()
	expr, exists := Expressions[id]
	mu.RUnlock()

	if !exists {
		http.Error(w, "Выражение не найдено", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]Expression{"expression": expr})
}

func internalTaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetTask(w, r)
	case http.MethodPost:
		SubmitResult(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if len(Tasks) == 0 {
		http.Error(w, "Нет задач", http.StatusNotFound)
		return
	}

	for _, task := range Tasks {
		json.NewEncoder(w).Encode(map[string]Task{"task": task})
		delete(Tasks, task.ID)
		return
	}
}

func SubmitResult(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Невалидные данные", http.StatusUnprocessableEntity)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	task, exists := Tasks[req.ID]
	if !exists {
		http.Error(w, "Задача не найдена", http.StatusNotFound)
		return
	}
	expr, exprExists := Expressions[task.ExpressionID]
	if exprExists {
		expr.Status = "done"
		expr.Result = req.Result
		Expressions[task.ExpressionID] = expr
	}

	delete(Tasks, req.ID)
	w.WriteHeader(http.StatusOK)
}

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
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type Expression struct {
	ID     string  `json:"id"`
	Expr   string  `json:"expression"`
	Status string  `json:"status"` // "pending", "processing", "done"
	Result float64 `json:"result"`
}

var (
	expressions = make(map[string]Expression)
	tasks       = make(map[string]Task)
	mu          sync.Mutex
)

func Start() {
	http.HandleFunc("/api/v1/calculate", addExpression)
	http.HandleFunc("/api/v1/expressions", getExpressions)
	http.HandleFunc("/api/v1/expressions/", getExpressionByID)
	http.HandleFunc("/internal/task", internalTaskHandler)

	fmt.Println("Оркестратор запущен на :8080")
	http.ListenAndServe(":8080", nil)
}


func addExpression(w http.ResponseWriter, r *http.Request) {
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

	// Генерация уникального ID
	id := fmt.Sprintf("expr-%d", time.Now().UnixNano())

	// Создание выражения
	expressions[id] = Expression{
		ID:     id,
		Expr:   req.Expression,
		Status: "pending",
		Result: 0,
	}

	tasks["task-1"] = Task{
		ID:            "task-1",
		Arg1:          2,
		Arg2:          2,
		Operation:     "*",
		OperationTime: 1000,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}


func getExpressions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	var exprs []Expression
	for _, expr := range expressions {
		exprs = append(exprs, expr)
	}

	json.NewEncoder(w).Encode(map[string][]Expression{"expressions": exprs})
}

func getExpressionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Извлекаем ID из URL
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")

	expr, exists := expressions[id]
	if !exists {
		http.Error(w, "Выражение не найдено", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]Expression{"expression": expr})
}


func internalTaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTask(w, r)
	case http.MethodPost:
		submitResult(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}


func getTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if len(tasks) == 0 {
		http.Error(w, "Нет задач", http.StatusNotFound)
		return
	}

	// Берем первую задачу
	var task Task
	for _, t := range tasks {
		task = t
		break
	}

	json.NewEncoder(w).Encode(map[string]Task{"task": task})
}


func submitResult(w http.ResponseWriter, r *http.Request) {
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

	// todo Обновляем результат задачи
	delete(tasks, req.ID)

	w.WriteHeader(http.StatusOK)
}
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type Agent struct {
	OrchestratorURL string
}

func (a *Agent) Start() {
	for {
		task, err := a.getTask()
		if err != nil {
			log.Println("Нет задач, ожидание...")
			time.Sleep(2 * time.Second)
			continue
		}

		result, err := executeTask(task)
		if err != nil {
			log.Printf("Ошибка выполнения задачи: %v\n", err)
			continue
		}

		if err := a.submitResult(task.ID, result); err != nil {
			log.Printf("Ошибка отправки результата: %v\n", err)
		}
	}
}

func (a *Agent) getTask() (Task, error) {
	resp, err := http.Get(a.OrchestratorURL + "/internal/task")
	if err != nil {
		return Task{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Task{}, fmt.Errorf("ошибка запроса задачи: %s", resp.Status)
	}

	var result struct {
		Task Task `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Task{}, err
	}

	return result.Task, nil
}


func executeTask(task Task) (float64, error) {
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2, nil
	case "-":
		return task.Arg1 - task.Arg2, nil
	case "*":
		return task.Arg1 * task.Arg2, nil
	case "/":
		if task.Arg2 == 0 {
			return 0, fmt.Errorf("деление на ноль")
		}
		return task.Arg1 / task.Arg2, nil
	default:
		return 0, fmt.Errorf("неизвестная операция: %s", task.Operation)
	}
}


func (a *Agent) submitResult(taskID string, result float64) error {

    reqBody, err := json.Marshal(map[string]interface{}{
        "id":     taskID,
        "result": result,
    })
    if err != nil {
        return fmt.Errorf("ошибка при сериализации JSON: %w", err)
    }

    resp, err := http.Post(a.OrchestratorURL+"/internal/task", "application/json", bytes.NewReader(reqBody))
    if err != nil {
        return fmt.Errorf("ошибка при отправке запроса: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("ошибка отправки результата: %s", resp.Status)
    }

    return nil
}
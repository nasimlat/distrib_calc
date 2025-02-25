package main

import (
    "github.com/nasimlat/distrib_calc/internal/agent"
    "log"
)


func main() {
	agent := agent.Agent{OrchestratorURL: "http://localhost:8080"}
	log.Println("Агент запущен")
	agent.Start()
}

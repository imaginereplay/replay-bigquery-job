package main

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

func main() {
	c := cron.New()

	// Adiciona uma tarefa para rodar todos os dias às 23h59
	c.AddFunc("59 23 * * *", func() {
		err := processJobs(time.Now())
		if err != nil {
			log.Println("Erro ao processar jobs:", err)
		}
	})

	// Inicia o cron
	c.Start()

	fmt.Println("Cron is running...")

	// Mantém o programa rodando
	select {}
}

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/robfig/cron/v3"
)

func main() {

	secretName := fmt.Sprintf("%s/imaginereplay", os.Getenv("environment"))

	fmt.Println(secretName)

	c := cron.New()

	// Adiciona uma tarefa para rodar todos os dias à 00:00
	c.AddFunc("* * * * *", func() {
		err := processJobs(time.Now(), secretName)
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

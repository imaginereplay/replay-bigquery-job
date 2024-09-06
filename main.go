package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dotenv-org/godotenvvault"
	"github.com/robfig/cron/v3"
)

func main() {
	err := godotenvvault.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	secretName := fmt.Sprintf("%s/imaginereplay", os.Getenv("environment"))

	c := cron.New()

	// Adiciona uma tarefa para rodar todos os dias às 23:59
	c.AddFunc("59 23 * * *", func() {
		err := processJobs(secretName)
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

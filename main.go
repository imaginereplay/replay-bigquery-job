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
		log.Println("Error loading .env file:", err)
	}

	secretName := fmt.Sprintf("%s/imaginereplay", os.Getenv("ENVIRONMENT"))

	c := cron.New()
	c.AddFunc("59 23 * * *", func() {
		err := processJobs(secretName)
		if err != nil {
			log.Println("Erro ao processar jobs:", err)
		}
	})

	c.Start()

	fmt.Println("Cron started...")

	select {}
}

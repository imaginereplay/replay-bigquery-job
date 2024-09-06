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

	// Create a new cron instance with a panic recovery wrapper
	c := cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger)))

	// Schedule the job to run at 00:10 every day
	_, err = c.AddFunc("35 21 * * *", func() {
		err := processJobs(secretName)
		if err != nil {
			log.Println("Erro ao processar jobs:", err)
		}
	})
	if err != nil {
		log.Println("Error scheduling the job:", err)
		return
	}

	// Start the cron scheduler
	c.Start()

	fmt.Println("Cron started...")

	// Block the main goroutine as long as the application is running
	select {}
}

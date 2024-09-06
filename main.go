package main

import (
	"fmt"
	"log"
	"net/http"
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

	// Route added to check if the cron is running and keep app alive
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Cron is running...")
	})

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	err = http.ListenAndServe(":"+port, nil)

	if err != nil {
		log.Println("Error starting server:", err)
	}
}

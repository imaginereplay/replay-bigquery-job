package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/dotenv-org/godotenvvault"
	"github.com/robfig/cron/v3"
)

func CreateAWSSessionTest() (*session.Session, error) {
	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if awsAccessKeyID == "" {
		log.Println("AWS_ACCESS_KEY_ID is not set in environment variables")
	} else {
	}

	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if awsSecretAccessKey == "" {
		log.Println("AWS_SECRET_ACCESS_KEY is not set in environment variables")
	} else {
		log.Printf("AWS_SECRET_ACCESS_KEY retrieved: [SECURE]")
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		log.Printf("Fatal Error: Unable to create AWS session: %v", err)
	}

	fmt.Println(awsAccessKeyID)
	fmt.Println(awsSecretAccessKey)

	fmt.Println("Session created")

	return sess, err
}

func main() {
	err := godotenvvault.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	CreateAWSSessionTest()

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

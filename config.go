package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"cloud.google.com/go/bigquery"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/bytedance/sonic"
	"google.golang.org/api/option"
)

// GetBigQueryClient connects with AWS Secrets Manager to retrieve the PEM files for JWT and creates a BigQuery client.
func GetBigQueryClient(secretName string) (*bigquery.Client, error) {
	projectID, err := GetSecret(secretName, "bigquery_project_id")
	if err != nil {
		log.Println("Fatal Error: Unable to retrieve 'bigquery_project_id' from AWS Secrets Manager")
		return nil, err
	}

	bigQueryPemStr, err := GetSecret(secretName, "bigquery_project_secret_pem")
	if err != nil {
		log.Println("Fatal Error: Unable to retrieve 'bigquery_project_secret_pem' from AWS Secrets Manager")
		return nil, err
	}

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID, option.WithCredentialsJSON([]byte(bigQueryPemStr)))
	if err != nil {
		log.Println("Fatal Error: Unable to initialize BigQuery client")
		return nil, err
	}

	return client, nil
}

// GetSecret retrieves a specific secret value from AWS Secrets Manager.
func GetSecret(secretName, secretKey string) (string, error) {
	svc, err := CreateSecretsManagerSession()
	if err != nil {
		log.Println("Fatal Error: Unable to create a session for AWS Secrets Manager")
		return "", err
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		log.Println("Fatal Error: Unable to retrieve secret value from AWS Secrets Manager")
		return "", err
	}

	var secretString string
	var secretData map[string]interface{}

	if result.SecretString != nil {
		secretString = *result.SecretString
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		length, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			log.Println("Fatal Error: Unable to decode binary secret from AWS Secrets Manager")
			return "", err
		}
		secretString = string(decodedBinarySecretBytes[:length])
	}

	if err := sonic.Unmarshal([]byte(secretString), &secretData); err != nil {
		log.Println("Fatal Error: Unable to unmarshal secret data from AWS Secrets Manager")
		return "", err
	}

	secretValue, ok := secretData[secretKey].(string)
	if !ok {
		log.Printf("Fatal Error: Secret key '%s' not found in AWS Secrets Manager response", secretKey)
		return "", fmt.Errorf("secret key '%s' not found", secretKey)
	}

	return secretValue, nil
}

// CreateSecretsManagerSession creates a session for AWS Secrets Manager.
func CreateSecretsManagerSession() (*secretsmanager.SecretsManager, error) {
	sess, err := CreateAWSSession()
	if err != nil {
		log.Println("Fatal Error: Unable to create AWS session")
		return nil, err
	}
	return secretsmanager.New(sess), nil
}

// CreateAWSSession creates and returns an AWS session configured for the 'us-east-1' region.
func CreateAWSSession() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		log.Println("Fatal Error: Unable to create AWS session")
	}
	return sess, err
}

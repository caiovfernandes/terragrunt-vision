package utils

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
)

func GetAwsCredentials() (string, string, string, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-2" // Default region if not set
	}

	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "default" // Default profile if not set
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithSharedConfigProfile(profile))
	if err != nil {
		return "", "", "", err
	}
	creds, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		return "", "", "", err
	}

	return creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, nil
}

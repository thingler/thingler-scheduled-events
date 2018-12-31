package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iotdataplane"
)

// Event passed to the Lambda function
type Event struct {
	Action   string `json:"action"`
	Region   string `json:"region"`
	Endpoint string `json:"iot_endpoint"`
	Topic    string `json:"iot_topic"`
}

// HandleEvent will update the IOT endpoint
func HandleEvent(ctx context.Context, event Event) error {

	sess := session.Must(session.NewSession())

	// Try first with Environment variables and secondly with IAM role
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(sess),
			},
		})

	config := &aws.Config{
		Region:      &event.Region,
		Credentials: creds,
		Endpoint:    &event.Endpoint,
	}

	clientIOT := iotdataplane.New(sess, config)

	turnOn := &TurnOn{
		IOTClient: clientIOT,
		Topic:     &event.Topic,
	}

	turnOff := &TurnOff{
		IOTClient: clientIOT,
		Topic:     &event.Topic,
	}

	action, err := NewActionFactory().
		AddAction(turnOn).
		AddAction(turnOff).
		GetAction(&event.Action)
	if err == nil {
		err = action.Do()
	}

	return err
}

func main() {
	lambda.Start(HandleEvent)
}

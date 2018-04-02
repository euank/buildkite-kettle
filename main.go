package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/euank/buildkite-kettle/buildkite"
	"github.com/euank/buildkite-kettle/config"
)

var c config.Config

func handleCron() error {
	return nil
}

func handleBuildEvent(e buildkite.BuildEvent) error {
	switch e.Event {
	case buildkite.EventTypeBuildScheduled:
		pc, err := c.GetPipelineConfig(*e.Pipeline.Name)
		if err != nil {
			return err
		}
		client := ec2.New(session.New())
		instances, err := client.DescribeInstances(&ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:buildkite"),
					Values: aws.StringSlice([]string{"name-" + pc.Name}),
				},
			},
		})
		if err != nil {
			log.Printf("error getting instance list: %v\n", err)
			return err
		}
		numInstances := 0
		for _, res := range instances.Reservations {
			for _, inst := range res.Instances {
				if *inst.State.Name == ec2.InstanceStateNameRunning || *inst.State.Name == ec2.InstanceStateNamePending {
					numInstances++
				}
			}
		}

		if numInstances >= pc.Config.MaxInstances {
			log.Printf("already %v instances for pipeline, no more needed", numInstances)
			return nil
		}

		_, err = client.RunInstances(&ec2.RunInstancesInput{
			ImageId: &pc.Config.AMI,
			KeyName: &pc.Config.KeypairName,
			BlockDeviceMappings: []*ec2.BlockDeviceMapping{
				{
					DeviceName: aws.String("/dev/xvda"),
					Ebs: &ec2.EbsBlockDevice{
						VolumeSize: aws.Int64(int64(pc.Config.StorageSize)),
					},
				},
			},
			InstanceInitiatedShutdownBehavior: aws.String("terminate"),
			InstanceType:                      &pc.Config.InstanceType,
			UserData:                          aws.String(base64.StdEncoding.EncodeToString([]byte(pc.Config.UserData))),
			MinCount:                          aws.Int64(1),
			MaxCount:                          aws.Int64(1),
			TagSpecifications: []*ec2.TagSpecification{
				{
					ResourceType: aws.String("instance"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("tag:buildkite"),
							Value: aws.String("name-" + pc.Name),
						},
					},
				},
			},
		})
		if err != nil {
			log.Printf("error launching instances: %v", err)
		}

		return err
	}

	return nil
}

func handleRequest(ctx context.Context, event events.APIGatewayProxyRequest) error {
	if os.Getenv("CRON") != "" {
		log.Printf("running cron\n")
		return handleCron()
	}

	if event.Headers["X-Buildkite-Token"] != c.BuildkiteToken {
		log.Printf("bad token\n")
		return fmt.Errorf("invalid buildkite token")
	}

	bke, err := buildkite.Unmarshal([]byte(event.Body))
	if err != nil {
		log.Printf("bad event: %v\n", err)
		return fmt.Errorf("unable to unmarshal event: %v", err)
	}

	switch e := bke.(type) {
	case buildkite.BuildEvent:
		return handleBuildEvent(e)
	default:
		log.Printf("did not handle event: %v\n", e)
	}

	return nil
}

func main() {
	c2, err := config.New()
	if err != nil {
		log.Fatalf("improper build: %s", err)
	}
	if os.Getenv("TEST_CONFIG") != "" {
		return
	}
	c = c2
	lambda.Start(handleRequest)
}

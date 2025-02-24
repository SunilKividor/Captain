package workers

import (
	"fmt"
	"os"

	"github.com/SunilKividor/Captain/pkg/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type ECSJobConfig struct {
	Container      *string
	Cluster        *string
	TaskDefinition *string
	Session        *session.Session
	BucketName     *string
	Key            *string
}

func NewECSJobConfig(session *session.Session, bucketName *string, Key *string) *ECSJobConfig {
	ecsCluster := os.Getenv("CLUSTER")
	container := os.Getenv("CONTAINER")
	taskDefinition := os.Getenv("TASK_DEFINITION")
	return &ECSJobConfig{
		Container:      &container,
		Cluster:        &ecsCluster,
		TaskDefinition: &taskDefinition,
		Session:        session,
		BucketName:     bucketName,
		Key:            Key,
	}
}

func (ecsJob *ECSJobConfig) RunECSJob() error {
	ecsClient := ecs.New(ecsJob.Session)

	updatedENV := []*ecs.KeyValuePair{
		{
			Name:  aws.String("bucket"),
			Value: ecsJob.BucketName,
		},
		{
			Name:  aws.String("key"),
			Value: ecsJob.Key,
		},
	}

	containerOverrides := []*ecs.ContainerOverride{
		{
			Name:        ecsJob.Container,
			Environment: updatedENV,
		},
	}

	taskOverride := &ecs.TaskOverride{
		ContainerOverrides: containerOverrides,
	}

	networkConfiguration := &ecs.NetworkConfiguration{
		AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
			AssignPublicIp: aws.String("ENABLED"),
			SecurityGroups: []*string{aws.String("sg-04bb500c0f2d9da67")},
			Subnets:        []*string{aws.String("subnet-0b7caf04e9f3a7bd6"), aws.String("subnet-0ee390b425d585f1b")},
		},
	}

	ecsRunTaskInput := &ecs.RunTaskInput{
		Cluster:              ecsJob.Cluster,
		TaskDefinition:       ecsJob.TaskDefinition,
		Count:                aws.Int64(1),
		LaunchType:           aws.String("FARGATE"),
		Overrides:            taskOverride,
		NetworkConfiguration: networkConfiguration,
	}

	runTaskOutput, err := ecsClient.RunTask(ecsRunTaskInput)
	utils.FailOnError(err, "Error Running Task")

	fmt.Printf("Task started with ARN: %s\n", *runTaskOutput.Tasks[0].TaskArn)
	return err
}

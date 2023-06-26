package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

func main() {
	// Get all S3 buckets in the AWS account
	allBucketsInAccount := describeAllBuckets()

	// Iterate all CloudFormation stacks and collect "AWS::S3::Bucket" resources
	bucketsFromActiveStacks := describeBucketNamesFromActiveStacks()

	// Intersect both arrays and delete buckets that are not in any active CloudFormation stack
	for _, bucket := range allBucketsInAccount {
		if !contains(bucketsFromActiveStacks, bucket) {
			fmt.Printf("Bucket %s is not in any active stack\n", bucket)
			// @TODO - delete bucket
		}
	}
}

func contains(haystack []string, needle string) bool {
	for _, a := range haystack {
		if a == needle {
			return true
		}
	}
	return false
}

func describeBucketNamesFromActiveStacks() []string {
	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	newContext := context.Background()
	cfg, err := config.LoadDefaultConfig(newContext)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	var bucketsAcrossStacks []string
	var stackNames []string

	cloudFormationClient := cloudformation.NewFromConfig(cfg)
	listStacksPaginator := cloudformation.NewListStacksPaginator(cloudFormationClient, &cloudformation.ListStacksInput{
		StackStatusFilter: []types.StackStatus{types.StackStatusCreateComplete, types.StackStatusUpdateComplete, types.StackStatusRollbackComplete, types.StackStatusUpdateCompleteCleanupInProgress, types.StackStatusUpdateRollbackComplete, types.StackStatusImportComplete, types.StackStatusImportRollbackComplete},
	})

	stacksCount := 0

	for listStacksPaginator.HasMorePages() {
		stacksCount++

		output, err := listStacksPaginator.NextPage(newContext)
		if err != nil {
			log.Fatalf("failed to list stacks, %v", err)
		}

		stackName := output.StackSummaries[0].StackName
		fmt.Printf("Stack Name: %s\n", *stackName)

		stackNames = append(stackNames, *stackName)

		singleStackResourcesPaginator := cloudformation.NewListStackResourcesPaginator(cloudFormationClient, &cloudformation.ListStackResourcesInput{
			StackName: output.StackSummaries[0].StackName,
		})
		for singleStackResourcesPaginator.HasMorePages() {
			output, err := singleStackResourcesPaginator.NextPage(newContext)

			if err != nil {
				log.Fatalf("failed to get resources in stack %v - %v", stackName, err)
			}

			for _, r := range output.StackResourceSummaries {
				fmt.Printf("%s - %s - %s\n", aws.ToString(r.ResourceType), aws.ToString(r.PhysicalResourceId), aws.ToString(r.LogicalResourceId))
				if aws.ToString(r.ResourceType) == "AWS::S3::Bucket" {
					bucketsAcrossStacks = append(bucketsAcrossStacks, aws.ToString(r.PhysicalResourceId))
				}
			}
		}
	}

	fmt.Printf("Total stacks: %d\n", stacksCount)

	return bucketsAcrossStacks
}

func describeAllBuckets() []string {
	newContext := context.Background()
	cfg, err := config.LoadDefaultConfig(newContext)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	var bucketNames []string

	s3Client := s3.NewFromConfig(cfg)
	listBucketsOutput, err := s3Client.ListBuckets(newContext, &s3.ListBucketsInput{})
	if err != nil {
		log.Fatalf("unable to list S3 buckets, %v", err)
	}
	for _, b := range listBucketsOutput.Buckets {
		bucketNames = append(bucketNames, aws.ToString(b.Name))
	}
	return bucketNames
}

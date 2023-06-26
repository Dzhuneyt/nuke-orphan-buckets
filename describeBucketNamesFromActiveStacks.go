package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"log"
)

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

		// Go through the stacks in this page of results
		for i := range output.StackSummaries {
			stackName := output.StackSummaries[i].StackName
			stackNames = append(stackNames, *stackName)

			singleStackResourcesPaginator := cloudformation.NewListStackResourcesPaginator(cloudFormationClient, &cloudformation.ListStackResourcesInput{
				StackName: stackName,
			})

			// Go through the resources of this Stack
			for singleStackResourcesPaginator.HasMorePages() {
				output, err := singleStackResourcesPaginator.NextPage(newContext)

				if err != nil {
					log.Fatalf("failed to get resources in stack %v - %v", stackName, err)
				}

				for _, r := range output.StackResourceSummaries {
					if aws.ToString(r.ResourceType) == "AWS::S3::Bucket" {
						bucketsAcrossStacks = append(bucketsAcrossStacks, aws.ToString(r.PhysicalResourceId))
					}
				}
			}
		}
	}

	return bucketsAcrossStacks
}

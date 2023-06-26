package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	// Get all S3 buckets in the AWS account
	allBucketsInAccount := describeAllBuckets()

	// Iterate all CloudFormation stacks and collect "AWS::S3::Bucket" resources
	bucketsFromActiveStacks := describeBucketNamesFromActiveStacks()

	var bucketsToDelete []string

	// Intersect both arrays and delete buckets that are not in any active CloudFormation stack
	for _, bucket := range allBucketsInAccount {
		if !contains(bucketsFromActiveStacks, bucket) {
			bucketsToDelete = append(bucketsToDelete, bucket)
		}
	}

	if len(bucketsToDelete) == 0 {
		println("No orphan buckets found.")
		return
	}

	println("Discovered orphan buckets (not managed by any CloudFormation stack:")

	for _, bucket := range bucketsToDelete {
		println(bucket)
	}

	c := askForConfirmation("Do you really want to delete the buckets listed above?")
	if c {
		// Delete bucket contents first, otherwise buckets can not be deleted
		for _, bucket := range bucketsToDelete {
			cfg, _ := config.LoadDefaultConfig(context.Background())
			BucketBasics{S3Client: s3.NewFromConfig(cfg)}.purgeBucket(bucket)
		}
		return
	}
}

package main

import (
	"context"
	"github.com/Dzhuneyt/nuke-orphan-buckets/s3helpers"
	"github.com/Dzhuneyt/nuke-orphan-buckets/util"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	cfg, _ := config.LoadDefaultConfig(context.Background())
	bucketBasics := s3helpers.BucketBasics{S3Client: s3.NewFromConfig(cfg)}

	// Get all S3 buckets in the AWS account
	allBucketsInAccount := bucketBasics.DescribeAllBuckets()

	// Iterate all CloudFormation stacks and collect "AWS::S3::Bucket" resources
	bucketsFromActiveStacks := bucketBasics.DescribeBucketNamesFromActiveStacks()

	var bucketsToDelete []string

	// Intersect both arrays and delete buckets that are not in any active CloudFormation stack
	for _, bucket := range allBucketsInAccount {
		if !util.Contains(bucketsFromActiveStacks, bucket) {
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

	c := util.AskForConfirmation("Do you really want to delete the buckets listed above?")
	if c {
		// Delete bucket contents first, otherwise buckets can not be deleted
		for _, bucket := range bucketsToDelete {
			bucketBasics.PurgeBucket(bucket)
		}
		return
	}
}

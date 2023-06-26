package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
)

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

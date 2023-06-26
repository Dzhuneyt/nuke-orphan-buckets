package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"log"
)

// BucketBasics encapsulates the Amazon Simple Storage Service (Amazon S3) actions
// used in the examples.
// It contains S3Client, an Amazon S3 service client that is used to perform bucket
// and object actions.
type BucketBasics struct {
	S3Client *s3.Client
}

func (basics BucketBasics) purgeBucket(bucketName string) {
	basics._purgeCurrentVersions(bucketName)
	basics._purgeDeleteMarkers(bucketName)
	basics._purgeBucket(bucketName)
}

func (basics BucketBasics) _purgeCurrentVersions(bucketName string) {
	var objects []types.ObjectIdentifier

	list, _ := basics.S3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	for _, version := range list.Contents {
		objects = append(objects, types.ObjectIdentifier{Key: version.Key})
	}

	chunks := chunkBy(objects, 100)

	for _, chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}
		_, err := basics.S3Client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &types.Delete{
				Objects: chunk,
			},
		})
		if err != nil {
			log.Printf("Couldn't batch delete objects in bucket %v. Here's why: %v\n", bucketName, err)
		}
		log.Printf("Deleted %v objects in bucket %v\n", len(chunk), bucketName)
	}
}

func (basics BucketBasics) _purgeDeleteMarkers(bucketName string) {
	var objects []types.ObjectIdentifier

	list, _ := basics.S3Client.ListObjectVersions(context.TODO(), &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	})

	for _, version := range list.Versions {
		objects = append(objects, types.ObjectIdentifier{Key: version.Key, VersionId: version.VersionId})
	}
	for _, version := range list.DeleteMarkers {
		objects = append(objects, types.ObjectIdentifier{Key: version.Key, VersionId: version.VersionId})
	}

	chunks := chunkBy(objects, 100)

	for _, chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}
		_, err := basics.S3Client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &types.Delete{
				Objects: chunk,
			},
		})
		if err != nil {
			log.Fatalf("Couldn't batch delete objects in bucket %v. Here's why: %v\n", bucketName, err)
		}
		log.Printf("Deleted %v objects in bucket %v\n", len(chunk), bucketName)
	}
}

func (basics BucketBasics) _purgeBucket(bucketName string) {
	cfg, _ := config.LoadDefaultConfig(context.Background())
	_, err := s3.NewFromConfig(cfg).DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		log.Printf("Failed to delete bucket %v. Here's why: %v\n", bucketName, err)
	} else {
		log.Printf("Deleted bucket %v\n", bucketName)
	}
}

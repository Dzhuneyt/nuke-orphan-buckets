package s3helpers

import (
	"context"
	"github.com/Dzhuneyt/nuke-orphan-buckets/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"log"
)

// BucketBasics encapsulates the Amazon Simple Storage Service (Amazon S3) actions
// used in the examples.
// It contains S3Client, an Amazon S3 service client that is used to perform bucket
// and object actions.
type BucketBasics struct {
	S3Client *s3.Client
}

func (basics BucketBasics) DescribeAllBuckets() []string {
	var bucketNames []string

	listBucketsOutput, err := basics.S3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatalf("unable to list S3 buckets, %v", err)
	}
	for _, b := range listBucketsOutput.Buckets {
		bucketNames = append(bucketNames, aws.ToString(b.Name))
	}
	return bucketNames
}

func (basics BucketBasics) PurgeBucket(bucketName string) {
	basics._purgeCurrentVersions(bucketName)
	basics._purgeDeleteMarkers(bucketName)
	basics._purgeBucket(bucketName)
}

func (basics BucketBasics) DescribeBucketNamesFromActiveStacks() []string {
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
		StackStatusFilter: []cfTypes.StackStatus{cfTypes.StackStatusCreateComplete, cfTypes.StackStatusUpdateComplete, cfTypes.StackStatusRollbackComplete, cfTypes.StackStatusUpdateCompleteCleanupInProgress, cfTypes.StackStatusUpdateRollbackComplete, cfTypes.StackStatusImportComplete, cfTypes.StackStatusImportRollbackComplete},
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

func (basics BucketBasics) _purgeCurrentVersions(bucketName string) {
	var objects []s3Types.ObjectIdentifier

	list, _ := basics.S3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	for _, version := range list.Contents {
		objects = append(objects, s3Types.ObjectIdentifier{Key: version.Key})
	}

	chunks := util.ChunkBy(objects, 100)

	for _, chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}
		_, err := basics.S3Client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &s3Types.Delete{
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
	var objects []s3Types.ObjectIdentifier

	list, _ := basics.S3Client.ListObjectVersions(context.TODO(), &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	})

	for _, version := range list.Versions {
		objects = append(objects, s3Types.ObjectIdentifier{Key: version.Key, VersionId: version.VersionId})
	}
	for _, version := range list.DeleteMarkers {
		objects = append(objects, s3Types.ObjectIdentifier{Key: version.Key, VersionId: version.VersionId})
	}

	chunks := util.ChunkBy(objects, 100)

	for _, chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}
		_, err := basics.S3Client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &s3Types.Delete{
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

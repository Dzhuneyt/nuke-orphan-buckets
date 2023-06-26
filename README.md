<img width="300" alt="repo icon" src="./.github/assets/icon.png"/>

# Nuke Orphaned S3 buckets

Find S3 buckets that are not managed by active CloudFormation stacks (orphaned buckets) and purge them.

Use cases:
* You are using Infrastructure as Code (e.g. AWS CDK, CloudFormation) and opted for the default `RetentionPolicy=Retain`, which steers on the side of caution and leaves behind orphaned S3 buckets, after their corresponding CloudFormation stack is deleted. This is fine for production environments, where data retention is important, but bad for development or QA environments, that are recreated often, leaving behind lots of junk S3 buckets.

## Getting started

1. Download one of the recent binaries for your OS from the <a href="https://github.com/Dzhuneyt/nuke-orphan-buckets/releases">Releases</a> page
2. Execute it `./nuke-orphan-buckets`
3. Optionally provide the AWS region or other optional parameters, respected by the AWS CLI: `AWS_REGION=eu-west-2 ./nuke-orphan-buckets`

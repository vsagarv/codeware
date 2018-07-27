# 1. aws-running-instances
Lists all running instances in an AWS account, region wise, with some instance attributes.

* Pre-requisites

  * [aws-cli] (http://docs.aws.amazon.com/cli/latest/userguide/installing.html) needs to be installed & configured with your AWS a/c credentials.

* Example run

  * <pre>$ aws-running-instances > "$(date '+%Y-%m-%d--%H:%M:%S')-AWS-Running-Instances.txt"</pre>

* Note: The o/p is best viewed in a terminal with sufficient column width or in a web-browser.

* Currently known regions:
  * us-east-1 US East (N. Virginia)
  * us-east-2 US East (Ohio)
  * us-west-2 US West (Oregon)
  * us-west-1 US West (N. California)
  * eu-west-1 EU (Ireland)
  * eu-central-1 EU (Frankfurt)
  * ap-southeast-1 Asia Pacific (Singapore)
  * ap-northeast-1 Asia Pacific (Tokyo)
  * ap-southeast-2 Asia Pacific (Sydney)
  * ap-northeast-2 Asia Pacific (Seoul)
  * ap-south-1 Asia Pacific (Mumbai)
  * sa-east-1 South America (São Paulo)
  
  ----

# 2. set-s3-ia-lc

Sets S3IA lifecycle transitions for AWS S3 buckets.
Also aborts incomplete multipart uploads.

* Usage

  <pre>set-s3-ia-lc.sh -p \<AWS CLI profile> -r \<AWS region> [-d \<days>]</pre>

\# reads a list of S3 bucket names - one per line - from stdin

\# -p: AWS CLI profile (--profile arg of aws command)

\# -r: AWS region of the S3 buckets

\# -d: Number of days (>= 30) after which an asset has to move from S3 to S3-IA (default = 30)

* Pre-requisites

  * [aws-cli] (http://docs.aws.amazon.com/cli/latest/userguide/installing.html) needs to be installed & configured with your AWS a/c credentials.
 
* Example run

  Run the script for all buckets in ap-south-1 region with the AWS CLI profile "test-aws-account"
  <pre>$ aws --profile=test-aws-account s3 --region=ap-south-1 ls | awk '{print $3}' | set-s3-ia-lc.sh -p test-aws-account -r ap-south-1 -d 45</pre>
  
* Dependencies
 
  1. A reference JSON file "s3_to_ia_lifecycle.json" in the same directory where this script is being run from. The JSON can be infered from a skeleton generated with this command:
  
  <pre>$ aws s3api put-bucket-lifecycle --generate-cli-skeleton</pre>

 * Tested on macOS 10.12.5 with aws-cli/1.15.60 Python/3.7.0 Darwin/16.6.0 botocore/1.10.59

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

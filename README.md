# aws-sfn-map-demo

This repository contains a small demo for image-resizing images in an S3 bucket using AWS Step Function Distribute Map.  
The repository is part of a blog post: https://alessandromarinoac.com/posts/aws/aws-step-function-distributed-map/

Infrastructure is built with AWS CDK typescript and the Lambda is instead built using Go.
Lower-level constructs were used for Step Function to allow the use of some unsupported features (at the time of writing 2023/01/29).  

The repository is just a demo project, it can and should be improved before using something like this in production.

## Useful cdk commands

* `npm run build`   compile typescript to js
* `npm run watch`   watch for changes and compile
* `npm run test`    perform the jest unit tests
* `cdk deploy`      deploy this stack to your default AWS account/region
* `cdk diff`        compare deployed stack with current state
* `cdk synth`       emits the synthesized CloudFormation template

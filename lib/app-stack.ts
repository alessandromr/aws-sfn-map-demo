import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { aws_s3, aws_iam } from 'aws-cdk-lib';
import { SFN } from './sfn/sfn';
import * as fs from 'fs';

export class AppStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);
    const file = fs.readFileSync("sfn-map-demo-state-machine.asl.json");

    const imageProcessor = new GoFunction(this, "image-processor", {
      entry: 'src/cmd/image-processor',
      timeout: cdk.Duration.seconds(30),
      memorySize: 2048,
    });

    const sourceBucket = new aws_s3.Bucket(this, 'images-source', {
      enforceSSL: true,
      blockPublicAccess: aws_s3.BlockPublicAccess.BLOCK_ALL,
      encryption: aws_s3.BucketEncryption.S3_MANAGED,
      versioned: true,
    });
    const destinationBucket = new aws_s3.Bucket(this, 'images-destination', {
      enforceSSL: true,
      blockPublicAccess: aws_s3.BlockPublicAccess.BLOCK_ALL,
      encryption: aws_s3.BucketEncryption.S3_MANAGED,
      versioned: true,
    });

    sourceBucket.grantRead(imageProcessor, "*");
    destinationBucket.grantPut(imageProcessor, "*");

    const stateMachine = new SFN(this, "state-machine", {
      namePrefix: "state-machine",
      definitionString: file.toString(),
    })

    sourceBucket.grantRead(stateMachine.role, "*");
    imageProcessor.grantInvoke(stateMachine.role);
  }
}

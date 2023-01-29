import { aws_iam, aws_stepfunctions } from 'aws-cdk-lib';
import { Construct } from 'constructs';

interface SFNProps {
    definitionString: string
    namePrefix: string
}

export class SFN extends Construct {
    public role: aws_iam.Role
    public sfn: aws_stepfunctions.CfnStateMachine

    constructor(scope: Construct, id: string, props: SFNProps) {
        super(scope, id);

        this.role = new aws_iam.Role(scope, props.namePrefix + '-role', {
            assumedBy: new aws_iam.ServicePrincipal("states.amazonaws.com"),
        });

        this.sfn = new aws_stepfunctions.CfnStateMachine(
            scope,
            "cfnStepFunction",
            {
                roleArn: this.role.roleArn,
                definitionString: props.definitionString,
                stateMachineName: props.namePrefix + '-machine',
            }
        );


        const sfnPolicy = new aws_iam.PolicyStatement({
            effect: aws_iam.Effect.ALLOW,
            resources: [
                this.sfn.attrArn,
            ],
            actions: [
                'states:StartExecution'
            ]
        });
          
        this.role.addToPolicy(sfnPolicy);
    }
}
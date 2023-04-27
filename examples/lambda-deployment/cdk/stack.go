package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	awslambdago "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CounterStackProps struct {
	awscdk.StackProps
}

func NewCounterStack(scope constructs.Construct, id string, props *CounterStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create a global count database.
	db := awsdynamodb.NewTable(stack, jsii.String("count"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("_pk"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode: awsdynamodb.BillingMode_PAY_PER_REQUEST,
		// Change this for production systems.
		RemovalPolicy:       awscdk.RemovalPolicy_DESTROY,
		TimeToLiveAttribute: jsii.String("_ttl"),
	})

	// Strip the binary, and remove the deprecated Lambda SDK RPC code for performance.
	// These options are not required, but make cold start faster.
	bundlingOptions := &awslambdago.BundlingOptions{
		GoBuildFlags: &[]*string{jsii.String(`-ldflags "-s -w" -tags lambda.norpc`)},
	}
	f := awslambdago.NewGoFunction(stack, jsii.String("handler"), &awslambdago.GoFunctionProps{
		Runtime:      awslambda.Runtime_PROVIDED_AL2(),
		Architecture: awslambda.Architecture_ARM_64(),
		Entry:        jsii.String("../lambda"),
		Bundling:     bundlingOptions,
		Environment: &map[string]*string{
			"TABLE_NAME": db.TableName(),
		},
	})
	// Grant DB access.
	db.GrantReadWriteData(f)
	// Add a Function URL.
	url := f.AddFunctionUrl(&awslambda.FunctionUrlOptions{
		AuthType: awslambda.FunctionUrlAuthType_NONE,
	})
	awscdk.NewCfnOutput(stack, jsii.String("lambdaFunctionUrl"), &awscdk.CfnOutputProps{
		ExportName: jsii.String("lambdaFunctionUrl"),
		Value:      url.Url(),
	})

	return stack
}

func main() {
	defer jsii.Close()
	app := awscdk.NewApp(nil)
	NewCounterStack(app, "CounterStack", &CounterStackProps{})
	app.Synth(nil)
}

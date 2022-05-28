module github.com/a-h/templ/storybook/example

go 1.16

replace github.com/a-h/templ => ../../

require (
	github.com/a-h/templ v0.0.0-00010101000000-000000000000
	github.com/aws/aws-cdk-go/awscdk/v2 v2.25.0
	github.com/aws/aws-cdk-go/awscdkapigatewayv2alpha/v2 v2.25.0-alpha.0
	github.com/aws/aws-cdk-go/awscdkapigatewayv2integrationsalpha/v2 v2.25.0-alpha.0
	github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2 v2.25.0-alpha.0
	github.com/aws/aws-lambda-go v1.27.0
	github.com/aws/constructs-go/constructs/v10 v10.1.20
	github.com/aws/jsii-runtime-go v1.59.0
)

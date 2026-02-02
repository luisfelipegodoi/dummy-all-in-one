package aws_only

import "tests/system/spec"

var Infra = spec.InfraSpec{
	Localstack: true,
	DynamoSeed: true,
}

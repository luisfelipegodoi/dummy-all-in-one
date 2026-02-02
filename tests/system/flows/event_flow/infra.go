package event_flow

import "tests/system/spec"

var Infra = spec.InfraSpec{
	Localstack: true,
	DynamoSeed: true,
	NATS:       true,
	Redis:      true,
}

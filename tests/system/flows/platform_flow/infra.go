package platform_flow

import "tests/system/spec"

var Infra = spec.InfraSpec{
	NATS:   true,
	Redis:  true,
	ArgoCD: true,
}

package spec

type InfraSpec struct {
	Localstack bool
	DynamoSeed bool
	NATS       bool
	Redis      bool
	ArgoCD     bool
}

// por cluster (target) -> InfraSpec
type Plan map[string]InfraSpec

package spec

type InfraSpec struct {
	Localstack bool
	DynamoSeed bool
	NATS       bool
	Redis      bool
	ArgoCD     bool
}

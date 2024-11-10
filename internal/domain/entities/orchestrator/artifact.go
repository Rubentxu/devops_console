// internal/domain/entities/artifact.go
package entities

type Artifact struct {
	Name string
	Data []byte
	Type string // Por ejemplo, "text/plain", "application/json"
}

package entities

type Worker interface {
	GetID() string
	GetType() string
	GetDetails() map[string]interface{}
}

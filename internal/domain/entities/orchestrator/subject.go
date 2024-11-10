package entities

type Subject interface {
	GetID() string
	GetName() string
}

type User struct {
	ID   string
	Name string
	Emai string
}

func (u User) GetID() string {
	return u.ID
}

func (u User) GetName() string {
	return u.Name
}

type Group struct {
	ID   string
	Name string
}

func (g Group) GetID() string {
	return g.ID
}

func (g Group) GetName() string {
	return g.Name
}

type ServiceAccount struct {
	ID   string
	Name string
}

func (sa ServiceAccount) GetID() string {
	return sa.ID
}

func (sa ServiceAccount) GetName() string {
	return sa.Name
}

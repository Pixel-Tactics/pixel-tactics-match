package matches_interfaces

type IAction interface {
	Apply(session ISession) error
	GetSourcePlayerId() string
	GetSourceHero() IHero
	GetData() map[string]interface{}
	GetName() string
}

package matches_interfaces

type IPlayer interface {
	GetSession() ISession
	GetId() string
	GetHeroList() []IHero
	GetData() map[string]interface{}
}

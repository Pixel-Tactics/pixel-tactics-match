package matches_interfaces

type HeroTemplateStats struct {
	MaxHealth   int `json:"maxHealth"`
	Damage      int `json:"damage"`
	AttackRange int `json:"attackRange"`
	MoveRange   int `json:"moveRange"`
}

type HeroTemplate interface {
	GetBaseStats() HeroTemplateStats
	GetName() string
	GetData() map[string]interface{}
}

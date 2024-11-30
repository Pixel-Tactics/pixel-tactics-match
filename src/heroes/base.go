package heroes

type BaseHeroEnum = string

const (
	BaseHeroKnight = "knight"
	BaseHeroMage   = "mage"
)

type BaseHero interface {
	GetInfo() *BaseHeroInfo
}

type BaseHeroInfo struct {
	Name      string
	BaseStats *BaseHeroStats
}

type BaseHeroStats struct {
	MaxHealth   int `json:"maxHealth"`
	Damage      int `json:"damage"`
	AttackRange int `json:"attackRange"`
	MoveRange   int `json:"moveRange"`
}

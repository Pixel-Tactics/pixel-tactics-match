package repositories

import (
	"errors"
	"sync"

	matches_heroes_templates "pixeltactics.com/match/src/matches/heroes/templates"
	matches_interfaces "pixeltactics.com/match/src/matches/interfaces"
)

type TemplateRepository struct {
	Knight matches_heroes_templates.Knight
	Mage   matches_heroes_templates.Mage
}

func (repo TemplateRepository) GetTemplateFromName(name string) (matches_interfaces.HeroTemplate, error) {
	if name == "knight" {
		return &repo.Knight, nil
	} else if name == "mage" {
		return &repo.Mage, nil
	}
	return nil, errors.New("invalid hero name")
}

var templateRepository *TemplateRepository = nil
var onceTemplate sync.Once

func GetTemplateRepository() *TemplateRepository {
	onceTemplate.Do(func() {
		templateRepository = &TemplateRepository{
			Knight: matches_heroes_templates.Knight{},
			Mage:   matches_heroes_templates.Mage{},
		}
	})
	return templateRepository
}

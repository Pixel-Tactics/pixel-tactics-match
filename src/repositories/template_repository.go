package repositories

import (
	"errors"

	"pixeltactics.com/match/src/matches"
)

type TemplateRepository struct {
	Knight matches.Knight
	Mage   matches.Mage
}

func (repo TemplateRepository) GetTemplateFromName(name string) (matches.HeroTemplate, error) {
	if name == "knight" {
		return &repo.Knight, nil
	} else if name == "mage" {
		return &repo.Mage, nil
	}
	return nil, errors.New("invalid hero name")
}

var templateRepository *TemplateRepository = nil

func GetTemplateRepository() *TemplateRepository {
	if templateRepository == nil {
		templateRepository = &TemplateRepository{
			Knight: matches.Knight{},
			Mage:   matches.Mage{},
		}
	}
	return templateRepository
}

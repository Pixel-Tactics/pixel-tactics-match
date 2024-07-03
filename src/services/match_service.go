package services

import (
	"time"

	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/matches"
	matches_interfaces "pixeltactics.com/match/src/matches/interfaces"
	"pixeltactics.com/match/src/repositories"
)

type MatchService struct {
	sessionRepository  *repositories.SessionRepository
	templateRepository *repositories.TemplateRepository
}

func (service *MatchService) CreateSession(data CreateSessionRequestDTO) (*matches.Session, error) {
	// TODO: check by fetch from account service (to check opponent if they exists)
	session, err := service.sessionRepository.CreateSession(data.PlayerId, data.OpponentId)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (service *MatchService) GetPlayerSession(playerId string) (map[string]interface{}, error) {
	session := service.sessionRepository.GetSessionByPlayerId(playerId)
	if session == nil {
		return nil, exceptions.SessionNotFound()
	}

	return session.GetDataSync(), nil
}

func (service *MatchService) PreparePlayer(data PreparePlayerRequestDTO) (bool, error) {
	session := service.sessionRepository.GetSessionByPlayerId(data.PlayerId)
	if session == nil {
		return false, exceptions.SessionNotFound()
	}

	chosenHeroList, err := service.nameToTemplate(data.ChosenHeroList)
	if err != nil {
		return false, err
	}

	err = session.PreparePlayerSync(data.PlayerId, chosenHeroList)
	if err != nil {
		return false, err
	}

	err = session.StartBattleSync()
	if err != nil && err.Error() == exceptions.HeroPickupError().Error() {
		return false, nil // other player not yet pickup
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (service *MatchService) ExecuteAction(data ExecuteActionRequestDTO) (map[string]interface{}, error) {
	session := service.sessionRepository.GetSessionByPlayerId(data.PlayerId)
	if session == nil {
		return nil, exceptions.SessionNotFound()
	}

	data.ActionSpecific["playerId"] = data.PlayerId
	return session.ExecuteActionSync(data.ActionName, data.ActionSpecific)
}

func (service *MatchService) EndTurn(playerId string) error {
	session := service.sessionRepository.GetSessionByPlayerId(playerId)
	if session == nil {
		return exceptions.SessionNotFound()
	}

	return session.EndTurnSync(playerId)
}

func (service *MatchService) GetOpponentId(playerId string) (string, error) {
	session := service.sessionRepository.GetSessionByPlayerId(playerId)
	if session == nil {
		return "", exceptions.SessionNotFound()
	}

	return session.GetOpponentPlayerIdSync(playerId)
}

func (service *MatchService) GetServerTime() time.Time {
	return time.Now()
}

func (service *MatchService) nameToTemplate(heroList []string) ([]matches_interfaces.HeroTemplate, error) {
	arr := []matches_interfaces.HeroTemplate{}
	for _, heroName := range heroList {
		template, err := service.templateRepository.GetTemplateFromName(heroName)
		if err != nil {
			return nil, err
		}
		arr = append(arr, template)
	}
	return arr, nil
}

func NewMatchService() *MatchService {
	return &MatchService{
		sessionRepository:  repositories.GetSessionRepository(),
		templateRepository: repositories.GetTemplateRepository(),
	}
}

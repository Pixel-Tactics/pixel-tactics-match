package services

import (
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/repositories"
)

type InvitationService interface{}

type InvitationServiceImpl struct {
	InvitationRepository repositories.InvitationRepository
	TransactionManager   databases.TransactionManager
	SessionService       SessionService
}

func (service *InvitationServiceImpl) Invite(playerId string, opponentId string) error {
	tx := service.TransactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	inviters, err := service.InvitationRepository.AllInInvitation(tx, playerId)
	if err != nil {
		return err
	}

	for _, inviter := range inviters {
		if inviter == opponentId {
			_, err = service.SessionService.CreateSession(opponentId, playerId)
			return err
		}
	}

	err = service.InvitationRepository.SaveInvitation(tx, playerId, opponentId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

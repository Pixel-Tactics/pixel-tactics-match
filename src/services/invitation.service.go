package services

import (
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/repositories"
)

type InvitationService interface {
	// Invites opponent to play. Returns true when there is mutual invitation.
	Invite(playerId string, opponentId string) (bool, error)
}

type InvitationServiceImpl struct {
	InvitationRepository repositories.InvitationRepository
	TransactionManager   databases.TransactionManager
	SessionService       SessionService
}

// Invites opponent to play. Returns true when there is mutual invitation.
func (service *InvitationServiceImpl) Invite(playerId string, opponentId string) (bool, error) {
	tx := service.TransactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	inviters, err := service.InvitationRepository.AllInInvitation(tx, playerId)
	if err != nil {
		return false, err
	}
	for _, inviter := range inviters {
		if inviter != opponentId {
			continue
		}

		err := service.InvitationRepository.DeleteInvitation(tx, opponentId, playerId)
		if err != nil {
			return false, err
		}
		_, err = service.SessionService.CreateSession(tx, opponentId, playerId)
		if err != nil {
			return false, err
		}
		err = tx.Commit()
		if err != nil {
			return false, err
		}
		return true, nil
	}

	err = service.InvitationRepository.SaveInvitation(tx, playerId, opponentId)
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, err
	}
	return false, nil
}

func NewInvitationService(
	invitationRepository repositories.InvitationRepository,
	transactionManager databases.TransactionManager,
	sessionService SessionService,
) InvitationService {
	return &InvitationServiceImpl{
		InvitationRepository: invitationRepository,
		TransactionManager:   transactionManager,
		SessionService:       sessionService,
	}
}

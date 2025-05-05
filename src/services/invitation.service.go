package services

import (
	"log"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/repositories"
)

const INVITE_EVENT = "INVITE_EVENT"

type InvitationService interface {
	// Invites opponent to play. Returns true when there is mutual invitation.
	Invite(inviteId string, playerId string, opponentId string) (bool, error)
}

type InviteEvent struct {
	SrcPlayerId string
	DstPlayerId string
}

type InvitationServiceImpl struct {
	InvitationRepository repositories.InvitationRepository
	TransactionManager   databases.TransactionManager
	SessionService       SessionService
	EventManager         events.EventManager
}

// Invites opponent to play. Returns true when there is mutual invitation.
func (service *InvitationServiceImpl) Invite(inviteId string, playerId string, opponentId string) (bool, error) {
	tx := service.TransactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	err := service.InvitationRepository.IsDuplicate(tx, inviteId)
	if err != nil {
		return false, err
	}

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

	err = service.EventManager.Emit(INVITE_EVENT, &InviteEvent{
		SrcPlayerId: playerId,
		DstPlayerId: opponentId,
	})
	if err != nil {
		log.Println(err)
	}

	return false, nil
}

func NewInvitationService(
	invitationRepository repositories.InvitationRepository,
	transactionManager databases.TransactionManager,
	sessionService SessionService,
	eventManager events.EventManager,
) InvitationService {
	return &InvitationServiceImpl{
		InvitationRepository: invitationRepository,
		TransactionManager:   transactionManager,
		SessionService:       sessionService,
		EventManager:         eventManager,
	}
}

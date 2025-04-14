package repositories

import (
	"errors"

	"pixeltactics.com/match/src/databases"
)

const MAX_INVITATION = 5

var ErrLimitReached = errors.New("limit reached")
var ErrAlreadyExists = errors.New("already exists")

type InvitationRepository interface {
	// Get all out invitations. Returns empty array when record doesn't exists.
	AllOutInvitation(tx databases.BadgerTx, playerId string) ([]string, error)

	// Get all in invitations. Returns empty array when record doesn't exists.
	AllInInvitation(tx databases.BadgerTx, playerId string) ([]string, error)

	// Checks whether invitation exists.
	Exists(tx databases.BadgerTx, playerId string, opponentId string) (bool, error)

	// Counts the number of invitation from players. When record doesn't exist, returns 0 invitation.
	Count(tx databases.BadgerTx, playerId string) (int, error)

	// Saves invitation. When there is atleast MAX_INVITATION invitations, ErrInvitationLimitReached will be thrown.
	SaveInvitation(tx databases.BadgerTx, playerId string, opponentId string) error

	// Deletes invitation. If invitation doesn't exists, it will return nil.
	DeleteInvitation(tx databases.BadgerTx, playerId string, opponentId string) error
}

type InvitationRepositoryImpl struct {
	badger databases.Badger
}

func (repo *InvitationRepositoryImpl) outKey(playerId string) string {
	return "invitation_out:" + playerId
}

func (repo *InvitationRepositoryImpl) inKey(playerId string) string {
	return "invitation_in:" + playerId
}

// Get all out invitations. Returns empty array when record doesn't exists.
func (repo *InvitationRepositoryImpl) AllOutInvitation(tx databases.BadgerTx, playerId string) ([]string, error) {
	var playerOut []string
	err := tx.Get(repo.outKey(playerId), &playerOut)
	if err != nil && err == databases.NotFoundException() {
		playerOut = make([]string, 0)
	}
	if err != nil {
		return nil, err
	}
	return playerOut, nil
}

// Get all in invitations. Returns empty array when record doesn't exists.
func (repo *InvitationRepositoryImpl) AllInInvitation(tx databases.BadgerTx, playerId string) ([]string, error) {
	var playerIn []string
	err := tx.Get(repo.inKey(playerId), &playerIn)
	if err != nil && err == databases.NotFoundException() {
		playerIn = make([]string, 0)
	}
	if err != nil {
		return nil, err
	}
	return playerIn, nil
}

// Checks whether invitation exists.
func (repo *InvitationRepositoryImpl) Exists(tx databases.BadgerTx, playerId string, opponentId string) (bool, error) {
	query := databases.GetQuery(tx, repo.badger)

	var inviters []string
	err := query.Get(repo.inKey(opponentId), &inviters)
	if err != nil && err == databases.NotFoundException() {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	for _, inviter := range inviters {
		if inviter == playerId {
			return true, nil
		}
	}
	return false, nil
}

// Counts the number of invitation from players. When record doesn't exist, returns 0 invitation.
func (repo *InvitationRepositoryImpl) Count(tx databases.BadgerTx, playerId string) (int, error) {
	query := databases.GetQuery(tx, repo.badger)

	var invited []string
	err := query.Get(repo.outKey(playerId), &invited)
	if err != nil && err == databases.NotFoundException() {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return len(invited), nil
}

// Saves invitation. When there is atleast MAX_INVITATION invitations, ErrInvitationLimitReached will be thrown.
func (repo *InvitationRepositoryImpl) SaveInvitation(tx databases.BadgerTx, playerId string, opponentId string) error {
	if tx == nil {
		return databases.ErrNilTransaction
	}

	var playerOut []string
	err := tx.Get(repo.outKey(playerId), &playerOut)
	if err != nil && err == databases.NotFoundException() {
		playerOut = make([]string, 0)
	}
	if err != nil {
		return err
	}

	if len(playerOut) >= MAX_INVITATION {
		return ErrLimitReached
	}

	for _, invited := range playerOut {
		if invited == opponentId {
			return ErrAlreadyExists
		}
	}

	var opponentIn []string
	err = tx.Get(repo.inKey(opponentId), &opponentIn)
	if err != nil && err == databases.NotFoundException() {
		opponentIn = make([]string, 0)
	}
	if err != nil {
		return err
	}

	playerOut = append(playerOut, opponentId)
	opponentIn = append(opponentIn, playerId)

	err = tx.Set(repo.outKey(playerId), &playerOut)
	if err != nil {
		return err
	}

	err = tx.Set(repo.inKey(opponentId), &opponentIn)
	if err != nil {
		return err
	}
	return nil
}

// Deletes invitation. If invitation doesn't exists, it will return nil.
func (repo *InvitationRepositoryImpl) DeleteInvitation(tx databases.BadgerTx, playerId string, opponentId string) error {
	if tx == nil {
		return databases.ErrNilTransaction
	}

	var playerOut []string
	err := tx.Get(repo.outKey(playerId), &playerOut)
	if err != nil && err == databases.NotFoundException() {
		playerOut = make([]string, 0)
	}
	if err != nil {
		return err
	}

	index := -1
	for i, invited := range playerOut {
		if invited == opponentId {
			index = i
		}
	}

	// Not found, so just return
	if index == -1 {
		return nil
	}

	playerOut = append(playerOut[:index], playerOut[index+1:]...)

	var opponentIn []string
	err = tx.Get(repo.inKey(opponentId), &opponentIn)
	if err != nil && err == databases.NotFoundException() {
		opponentIn = make([]string, 0)
	}
	if err != nil {
		return err
	}

	index = -1
	for i, inviter := range opponentIn {
		if inviter == playerId {
			index = i
		}
	}

	// Not supposed to be here, but yeah
	if index == -1 {
		return nil
	}

	opponentIn = append(opponentIn[:index], opponentIn[index+1:]...)

	err = tx.Set(repo.outKey(playerId), &playerOut)
	if err != nil {
		return err
	}

	err = tx.Set(repo.inKey(opponentId), &opponentIn)
	if err != nil {
		return err
	}
	return nil
}

func NewInvitationRepository(
	badger databases.Badger,
) InvitationRepository {
	return &InvitationRepositoryImpl{
		badger: badger,
	}
}

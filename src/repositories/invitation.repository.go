package repositories

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

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

	// Saves invitation. When there is atleast MAX_INVITATION invitations, ErrInvitationLimitReached will be thrown.
	SaveInvitation(tx databases.BadgerTx, playerId string, opponentId string) error

	// Deletes invitation. If invitation doesn't exists, it will return nil.
	DeleteInvitation(tx databases.BadgerTx, playerId string, opponentId string) error
}

type InvitationRepositoryImpl struct {
}

// Get all out invitations. Returns empty array when record doesn't exists.
func (repo *InvitationRepositoryImpl) AllOutInvitation(tx databases.BadgerTx, playerId string) ([]string, error) {
	// TODO: must prevent injections (all repos)
	if tx == nil {
		return nil, databases.ErrNilTransaction
	}
	_, rawValues, err := tx.PrefixScan("invitation:out:" + playerId + ":")
	if err != nil {
		return nil, err
	}
	inviteds, err := repo.unmarshalToStrings(rawValues)
	if err != nil {
		return nil, err
	}
	return inviteds, nil
}

// Get all in invitations. Returns empty array when record doesn't exists.
func (repo *InvitationRepositoryImpl) AllInInvitation(tx databases.BadgerTx, playerId string) ([]string, error) {
	if tx == nil {
		return nil, databases.ErrNilTransaction
	}
	_, rawValues, err := tx.PrefixScan("invitation:in:" + playerId + ":")
	if err != nil {
		return nil, err
	}
	inviters, err := repo.unmarshalToStrings(rawValues)
	if err != nil {
		return nil, err
	}
	return inviters, nil
}

// Checks whether invitation exists.
func (repo *InvitationRepositoryImpl) Exists(tx databases.BadgerTx, playerId string, opponentId string) (bool, error) {
	if tx == nil {
		return false, databases.ErrNilTransaction
	}
	opponentIds, err := repo.AllOutInvitation(tx, playerId)
	if err != nil {
		return false, err
	}
	for _, curId := range opponentIds {
		if curId == opponentId {
			return true, nil
		}
	}
	return false, nil
}

// Saves invitation. When there is already MAX_INVITATION invitations, earliest invitation will be deleted.
func (repo *InvitationRepositoryImpl) SaveInvitation(tx databases.BadgerTx, playerId string, opponentId string) error {
	if tx == nil {
		return databases.ErrNilTransaction
	}

	playerOutKeys, rawValues, err := tx.PrefixScan("invitation:out:" + playerId + ":")
	if err != nil {
		return err
	}
	playerOutIds, err := repo.unmarshalToStrings(rawValues)
	if err != nil {
		return err
	}

	for _, curId := range playerOutIds {
		if curId == opponentId {
			return ErrAlreadyExists
		}
	}

	if len(playerOutKeys) >= MAX_INVITATION {
		for i := 0; i < len(playerOutKeys)-MAX_INVITATION+1; i++ {
			err = tx.Delete(playerOutKeys[i])
			if err != nil {
				return err
			}
		}
	}

	curTime := strconv.Itoa(int(time.Now().Unix()))
	err = tx.Set("invitation:out:"+playerId+":"+curTime+":"+opponentId, &opponentId)
	if err != nil {
		return err
	}
	err = tx.Set("invitation:in:"+opponentId+":"+curTime+":"+playerId, &playerId)
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

	playerOutKeys, rawValues, err := tx.PrefixScan("invitation:out:" + playerId + ":")
	if err != nil {
		return err
	}
	playerOutIds, err := repo.unmarshalToStrings(rawValues)
	if err != nil {
		return err
	}
	for i, curId := range playerOutIds {
		if curId == opponentId {
			err = tx.Delete(playerOutKeys[i])
			if err != nil {
				return err
			}
		}
	}

	opponentInKeys, rawValues, err := tx.PrefixScan("invitation:in:" + opponentId + ":")
	if err != nil {
		return err
	}
	opponentInIds, err := repo.unmarshalToStrings(rawValues)
	if err != nil {
		return err
	}
	for i, curId := range opponentInIds {
		if curId == playerId {
			err = tx.Delete(opponentInKeys[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (repo *InvitationRepositoryImpl) unmarshalToStrings(raw [][]byte) ([]string, error) {
	values := make([]string, 0)
	for _, rawValue := range raw {
		var value string
		err := json.Unmarshal(rawValue, &value)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func NewInvitationRepository() InvitationRepository {
	return &InvitationRepositoryImpl{}
}

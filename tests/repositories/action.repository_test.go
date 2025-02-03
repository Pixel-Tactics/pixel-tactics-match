package repositories_test

import (
	"log"
	"testing"

	"github.com/go-playground/assert/v2"
	"pixeltactics.com/match/src/core/actions"
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/repositories"
	"pixeltactics.com/match/src/utils/physics"
	test_utils "pixeltactics.com/match/tests/utils"
)

func TestGetSessionActionLogsValid(t *testing.T) {
	db, repo, seededLogs := createActionLogRepository()
	defer test_utils.CloseTestBadger(db)

	logs, err := repo.GetSessionActionLogs(nil, "rendem")
	assert.Equal(t, err, nil)
	assert.Equal(t, len(logs), len(seededLogs))
}

func TestGetSessionActionLogsInvalid(t *testing.T) {
	db, repo, _ := createActionLogRepository()
	defer test_utils.CloseTestBadger(db)

	_, err := repo.GetSessionActionLogs(nil, "rendem3")
	assert.NotEqual(t, err, nil)
}

func TestCreateActionLog(t *testing.T) {
	db, repo, _ := createActionLogRepository()
	defer test_utils.CloseTestBadger(db)

	tx := db.NewReadWriteTransaction()
	defer tx.Discard()

	expectedLog := &models.ActionLog{
		SessionId: "testos",
		PlayerId:  "plr2",
		Order:     0,
		Action:    actions.NewAttackAction("plr2", heroes.BaseHeroKnight, heroes.BaseHeroKnight, 100),
	}
	_, err := repo.CreateActionLog(tx, expectedLog)
	assert.Equal(t, err, nil)

	tx.Commit()

	logs, err := repo.GetSessionActionLogs(nil, "testos")
	assert.Equal(t, err, nil)
	assert.Equal(t, len(logs), 1)

	lastLog := logs[len(logs)-1]
	assert.Equal(t, lastLog.Action.GetType(), actions.ActionTypeAttack)
	assert.Equal(t, *lastLog, *expectedLog)
}

func createActionLogRepository() (databases.Badger, repositories.ActionLogRepository, []*models.ActionLog) {
	db := test_utils.NewTestBadger()
	repo := repositories.NewActionLogRepositoryImpl(db)
	seededLogs := seedActionLogs(db, repo)
	return db, repo, seededLogs
}

func seedActionLogs(db databases.Badger, repo repositories.ActionLogRepository) []*models.ActionLog {
	actions := []actions.Action{
		actions.NewAttackAction("plr1", heroes.BaseHeroMage, heroes.BaseHeroMage, 10),
		actions.NewMoveAction("plr1", heroes.BaseHeroMage, []physics.Direction{physics.DirectionDown, physics.DirectionRight}),
		actions.NewAttackAction("plr1", heroes.BaseHeroMage, heroes.BaseHeroMage, 10),
		actions.NewMoveAction("plr1", heroes.BaseHeroMage, []physics.Direction{physics.DirectionUp}),
	}
	tx := db.NewReadWriteTransaction()
	defer tx.Discard()
	actionLogs := make([]*models.ActionLog, 0)
	for i, action := range actions {
		curLog := &models.ActionLog{
			SessionId: "rendem",
			PlayerId:  "plr",
			Order:     i,
			Action:    action,
		}
		_, err := repo.CreateActionLog(tx, curLog)
		if err != nil {
			log.Println(err)
			panic("invalid seeding")
		}
		actionLogs = append(actionLogs, curLog)
	}
	tx.Commit()
	return actionLogs
}

package repositories_test

import (
	"testing"

	"github.com/go-playground/assert/v2"
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/repositories"
	test_utils "pixeltactics.com/match/tests/utils"
)

func TestGetMapEmpty(t *testing.T) {
	db, repo, _ := createMapRepository()
	defer test_utils.CloseTestBadger(db)

	_, err := repo.GetMapBySessionId(nil, "1")
	assert.Equal(t, err, nil)
	// assert.Equal(t, len(logs), len(seededLogs))
}

func createMapRepository() (databases.Badger, repositories.MapRepository, []*models.ActionLog) {
	db := test_utils.NewTestBadger()
	repo := repositories.NewMapRepository(db)
	// seededLogs := seedActionLogs(db, repo)
	return db, repo, nil
}

// func seedActionLogs(db databases.Badger, repo repositories.ActionLogRepository) []*models.ActionLog {
// 	actions := []actions.Action{
// 		actions.NewAttackAction("plr1", heroes.BaseHeroMage, heroes.BaseHeroMage, 10),
// 		actions.NewMoveAction("plr1", heroes.BaseHeroMage, []physics.Direction{physics.DirectionDown, physics.DirectionRight}),
// 		actions.NewAttackAction("plr1", heroes.BaseHeroMage, heroes.BaseHeroMage, 10),
// 		actions.NewMoveAction("plr1", heroes.BaseHeroMage, []physics.Direction{physics.DirectionUp}),
// 	}
// 	tx := db.NewReadWriteTransaction()
// 	defer tx.Discard()
// 	actionLogs := make([]*models.ActionLog, 0)
// 	for i, action := range actions {
// 		curLog := &models.ActionLog{
// 			SessionId: "rendem",
// 			PlayerId:  "plr",
// 			Order:     i,
// 			Action:    action,
// 		}
// 		_, err := repo.CreateActionLog(tx, curLog)
// 		if err != nil {
// 			log.Println(err)
// 			panic("invalid seeding")
// 		}
// 		actionLogs = append(actionLogs, curLog)
// 	}
// 	_ = tx.Commit()
// 	return actionLogs
// }

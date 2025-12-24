package syncengine

import (
	"github.com/jasmineyas/splitwise-lunchmoney/config"
	"github.com/jasmineyas/splitwise-lunchmoney/lunchmoney"
	"github.com/jasmineyas/splitwise-lunchmoney/models"
	"github.com/jasmineyas/splitwise-lunchmoney/splitwise"
)

type Engine struct {
	swClient *splitwise.Client
	lmClient *lunchmoney.Client
	config   *config.Config
}

func New(swClient *splitwise.Client, lmClient *lunchmoney.Client, cfg *config.Config) *Engine {
	return &Engine{
		swClient: swClient,
		lmClient: lmClient,
		config:   cfg,
	}
}

// NTS
// also let's understand the test http handler the crazy wrap what are they
// goal tmr is to have a working integration test somewhere - either just running this
// and see a new transaciton in splitwise is logged in luchmoney
// the compare and update can come in the future :D
// and find a place to deploy it
// then we do feature update :D

func (e *Engine) Sync(toCreate, toUpdate, toDelete []models.SplitwiseExpense) error {
	err := e.syncUpdate(toUpdate)
	if err != nil {
		return err
	}

	err = e.syncCreate(toCreate)
	if err != nil {
		return err
	}

	err = e.syncDelete(toDelete)
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) syncCreate(toCreate []models.SplitwiseExpense) error {
	// transform a list of Splitwise expenses to LunchMoney transactions

	// post the transactions to lunch money

	// post a comment to splitwise expense with lunch money transaction id
	return nil
}

// to come
func (e *Engine) syncUpdate(toUpdate []models.SplitwiseExpense) error {
	//  compare existing splitwise expense with the expense in the comment
	return nil
}

// to come
func (e *Engine) syncDelete(toDelete []models.SplitwiseExpense) error {
	//  placeholder implementation
	return nil
}

// dont worry about generate snapshot hash etc. for now.

func TransformSWToLMTransaction() error {
	// will come back to this later

	return nil
}

func PostTransactionToLunchMoney() error {
	return nil
}

func PostCommentToSplitwise() error {
	return nil
}

func GenerateSnapshotHash() error {
	return nil
}

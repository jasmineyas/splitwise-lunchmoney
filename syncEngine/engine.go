package syncengine

import (
	"github.com/jasmineyas/splitwise-lunchmoney/config"
	"github.com/jasmineyas/splitwise-lunchmoney/lunchmoney"
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

// NTS - read about class approach vs not class approah (refresh memory)
// update the main file (orchestrator)
// create a new util class for detector
// also let's understand the test http handler the crazy wrap what are they
// goal tmr is to have a working integration test somewhere - either just running this
// and see a new transaciton in splitwise is logged in luchmoney
// the compare and update can come in the future :D
// and find a place to deploy it
// then we do feature update :D

// it should receive a list of Splitwise expenses that need to be updated
func InitializeSyncEngine() error {
	// transform a list of Splitwise expenses to LunchMoney transactions

	// post the transactions to lunch money

	// post a comment to splitwise expense with lunch money transaction id

	return nil
}

// dont worry about generate snapshot hash etc. for now.

func TransformSWToLMTransaction() error {
	// will come back to this later

	//
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

// is that part of the sync engine? or separate util?
func DetectChanges() error {
	return nil
}

func ParseSyncComment() error {
	return nil
}

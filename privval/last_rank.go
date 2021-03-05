package privval

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	tm_json "github.com/tendermint/tendermint/libs/json"
)

const (
	// LastRankFile is the full file name of the file that persists the validator's
	// last rank.
	LastRankFile = "last_rank.json"

	// PermLastRankFile determines the default file permissions for the last_rank.json
	// file.
	PermLastRankFile = os.FileMode(0644)
)

// LastRank defines the contents of the last_rank.json file.
type LastRank struct {
	Rank int `json:"rank"`
}

// LastRankFilePath returns the absolute path to the last_rank.json file.
func LastRankFilePath(cfgDir string) string {
	return cfgDir + "/" + LastRankFile
}

// CheckAndLoadLastRank checks whether the last_rank.json file exists or not. If it
// does, it loads the rank from that file.
func (pv *SCFilePV) CheckAndLoadLastRank(cfgDir string, logger *log.Logger) error {
	if _, err := os.Stat(LastRankFilePath(cfgDir)); !os.IsNotExist(err) {
		logger.Printf("[DEBUG] signctrl: Found %v at %v", LastRankFile, cfgDir)
		bytes, err := ioutil.ReadFile(LastRankFilePath(cfgDir))
		if err != nil {
			return err
		}

		var lr LastRank
		if err := tm_json.Unmarshal(bytes, &lr); err != nil {
			return err
		}
		if lr.Rank < 1 {
			return fmt.Errorf("rank in last_rank.json must be 1 or higher")
		}
		if pv.BaseSignCtrled.GetRank() != lr.Rank {
			logger.Printf("[WARN] signctrl: Updating validator's rank to %v from last_rank.json", lr.Rank)
			pv.BaseSignCtrled.SetRank(lr.Rank)
		}
		return nil
	}

	logger.Printf("[DEBUG] signctrl: No %v found at %v, using rank %v from config", LastRankFile, cfgDir, pv.Config.Base.StartRank)

	return nil
}

// Save saves the last rank to the last_rank.json file.
func (pv *SCFilePV) Save(cfgDir string, logger *log.Logger) error {
	pv.Logger.Printf("[INFO] signctrl: Saving current rank %v to %v...", pv.BaseSignCtrled.GetRank(), LastRankFile)
	lrFile, err := tm_json.MarshalIndent(&LastRank{Rank: pv.BaseSignCtrled.GetRank()}, "", "\t")
	if err != nil {
		return err
	}
	os.Remove(LastRankFilePath(cfgDir)) // TODO: Remove before creating a new last_rank.json?

	return ioutil.WriteFile(LastRankFilePath(cfgDir), lrFile, PermLastRankFile)
}

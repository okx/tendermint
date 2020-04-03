package state

import (
	"fmt"
	"github.com/tendermint/tendermint/libs/log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func InitBackup(logger log.Logger, rootdir string, backupPerHeight int64) error {
	historyDataRootPath := filepath.Join(rootdir, "state_history")

	err := createStateHistoryRootFolder(historyDataRootPath)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create %s. error:%s", historyDataRootPath, err.Error()))
		return err
	}
	logger.Info("succeesfully create folder " + historyDataRootPath)

	return nil
}

func InitRecover(logger log.Logger, rootdir string, recoverHeight int64) error {
	sourceDataPath := filepath.Join(rootdir, "data")
	historyPath := filepath.Join(rootdir, "state_history", strconv.FormatInt(recoverHeight, 10))

	//[example] ./state_history/20000 -> ./data
	err := copyStateHistoryFolder(historyPath, sourceDataPath)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to copy %s to %s . error:%s", historyPath, sourceDataPath, err.Error()))
		return err
	}
	logger.Info("succeesfully copy " + historyPath + " to " + sourceDataPath)

	return nil
}

func copyStateHistoryFolder(historyPath string, sourceDataPath string) error {
	//To judge if the state_history path exsits
	exist, err := pathExists(historyPath)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("Path: [%s] is not exist", historyPath)
	}

	//delete old state data
	err = DeleteStateData(sourceDataPath)
	if err != nil {
		return err
	}

	//Copy the [./state_history/xxx/.]  to [./data]
	_, err = CopyHistoryToSource(historyPath, sourceDataPath)
	if err != nil {
		return err
	}

	return nil
}

func DeleteStateData(DataPath string) error {
	err := os.RemoveAll(filepath.Join(DataPath, "application.db"))
	if err != nil {
		return err
	}

	err = os.RemoveAll(filepath.Join(DataPath, "evidence.db"))
	if err != nil {
		return err
	}

	err = os.RemoveAll(filepath.Join(DataPath, "state.db"))
	if err != nil {
		return err
	}

	err = os.RemoveAll(filepath.Join(DataPath, "priv_validator_state.json"))
	if err != nil {
		return err
	}

	err = os.RemoveAll(filepath.Join(DataPath, "tx_index.db"))
	if err != nil {
		return err
	}

	err = os.RemoveAll(filepath.Join(DataPath, "cs.wal"))
	if err != nil {
		return err
	}

	return nil
}

func CopyHistoryToSource(historyPath string, sourceDataPath string) ([]byte, error) {
	f, err := exec.Command("cp", "-rf", historyPath+"/.", sourceDataPath).Output()
	if err != nil {
		return nil, err
	}
	return f, nil
}

func CopySourceToHistory(sourceDataPath string, historyPath string) ([]byte, error) {
	err := os.MkdirAll(historyPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	f, err := exec.Command("cp", "-r", filepath.Join(sourceDataPath, "application.db"), historyPath).Output()
	if err != nil {
		return nil, err
	}
	f, err = exec.Command("cp", "-r", filepath.Join(sourceDataPath, "evidence.db"), historyPath).Output()
	if err != nil {
		return nil, err
	}
	f, err = exec.Command("cp", "-r", filepath.Join(sourceDataPath, "state.db"), historyPath).Output()
	if err != nil {
		return nil, err
	}
	f, err = exec.Command("cp", "-r", filepath.Join(sourceDataPath, "priv_validator_state.json"), historyPath).Output()
	if err != nil {
		return nil, err
	}
	f, err = exec.Command("cp", "-r", filepath.Join(sourceDataPath, "tx_index.db"), historyPath).Output()
	if err != nil {
		return nil, err
	}
	f, err = exec.Command("cp", "-r", filepath.Join(sourceDataPath, "cs.wal"), historyPath).Output()
	if err != nil {
		return nil, err
	}
	return f, nil
}

func createStateHistoryRootFolder(historyRootPath string) error {
	err := os.MkdirAll(historyRootPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}


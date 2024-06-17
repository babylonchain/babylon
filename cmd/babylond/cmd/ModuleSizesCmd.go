package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/babylonchain/babylon/app"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"
)

func OpenDB(dir string) (dbm.DB, error) {
	switch {
	case strings.HasSuffix(dir, ".db"):
		dir = dir[:len(dir)-3]
	case strings.HasSuffix(dir, ".db/"):
		dir = dir[:len(dir)-4]
	default:
		return nil, fmt.Errorf("database directory must end with .db")
	}

	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	// TODO: doesn't work on windows!
	cut := strings.LastIndex(dir, "/")
	if cut == -1 {
		return nil, fmt.Errorf("cannot cut paths on %s", dir)
	}
	name := dir[cut+1:]
	db, err := dbm.NewGoLevelDB(name, dir[:cut], nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

type ModuleStats struct {
	NodeCount      int
	TotalSizeBytes int
}

func getModuleName(key string, modules []string) (string, error) {
	for _, name := range modules {
		stringToFind := fmt.Sprintf("k:%s", name)

		if strings.Contains(key, stringToFind) {
			return name, nil
		}
	}
	return "", fmt.Errorf("no module found for key %s", key)
}

func PrintDBStats(db dbm.DB, moduleNames []string) {

	prefix := map[string]*ModuleStats{}
	for _, name := range moduleNames {
		prefix[name] = &ModuleStats{
			NodeCount:      0,
			TotalSizeBytes: 0,
		}
	}

	count := 0
	totalSizeBytes := 0
	itr, err := db.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}

	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		fullKey := itr.Key()
		fullValue := itr.Value()
		fullKeyString := string(fullKey)

		moduleName, err := getModuleName(fullKeyString, moduleNames)

		if err != nil {
			continue
		}

		if err != nil {
			fmt.Printf("Error making node: %s\n", err.Error())
		}

		prefix[moduleName].NodeCount++

		keyValueSize := len(fullKey) + len(fullValue)

		prefix[moduleName].TotalSizeBytes += keyValueSize
		totalSizeBytes += keyValueSize
		count++
	}

	if err := itr.Error(); err != nil {
		panic(err)
	}
	fmt.Printf("DB contains %d entries\n", count)
	fmt.Printf("Total size of database: %d bytes\n", totalSizeBytes)
	for k, v := range prefix {
		fmt.Printf("Module %s:\n", k)
		fmt.Printf("Number of tree state nodes: %d\n", v.NodeCount)
		fmt.Printf("Total size of of module storage: %d bytes\n", v.TotalSizeBytes)
		fmt.Printf("Fraction of total size: %.3f\n", float64(v.TotalSizeBytes)/float64(totalSizeBytes))
	}
}

func ModuleSizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module-sizes",
		Short: "print sizes of each module in the database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pathToDB := args[0]

			babylonApp := app.NewTmpBabylonApp()

			names := babylonApp.ModuleManager.ModuleNames()

			db, err := OpenDB(pathToDB)

			if err != nil {
				return err
			}

			PrintDBStats(db, names)

			return nil
		},
	}

	return cmd
}

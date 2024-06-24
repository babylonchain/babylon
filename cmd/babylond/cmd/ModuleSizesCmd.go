package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/babylonchain/babylon/app"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"
)

const (
	FlagPrintInterval = "print-interval"
)

func OpenDB(dir string) (dbm.DB, error) {
	fmt.Printf("Opening database at: %s\n", dir)
	defer fmt.Printf("Opened database at: %s\n", dir)

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

func printModuleStats(stats map[string]*ModuleStats, totalCount uint64, numErrors uint64, totalSizeBytes int) {
	fmt.Printf("****************** Printing module stats ******************\n")
	fmt.Printf("Total number of nodes in db: %d\n", totalCount)
	fmt.Printf("Total number of unknown nodes: %d\n", numErrors)
	fmt.Printf("Total size of database: %d bytes\n", totalSizeBytes)
	for k, v := range stats {
		fmt.Printf("Module %s:\n", k)
		fmt.Printf("Number of tree state nodes: %d\n", v.NodeCount)
		fmt.Printf("Total size of of module storage: %d bytes\n", v.TotalSizeBytes)
		fmt.Printf("Fraction of total size: %.3f\n", float64(v.TotalSizeBytes)/float64(totalSizeBytes))
	}
	fmt.Printf("****************** Printed stats for all Babylon modules ******************\n")
}

func PrintDBStats(db dbm.DB, moduleNames []string, printInterval int) {
	fmt.Printf("****************** Starting to iterate over whole database ******************\n")
	prefix := map[string]*ModuleStats{}
	for _, name := range moduleNames {
		prefix[name] = &ModuleStats{
			NodeCount:      0,
			TotalSizeBytes: 0,
		}
	}

	count := uint64(0)
	numErrors := uint64(0)
	totalSizeBytes := 0
	itr, err := db.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("****************** Retrived database iterator ******************\n")

	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		count++
		if count%uint64(printInterval) == 0 {
			printModuleStats(prefix, count, numErrors, totalSizeBytes)
		}

		fullKey := itr.Key()
		fullValue := itr.Value()
		fullKeyString := string(fullKey)

		moduleName, err := getModuleName(fullKeyString, moduleNames)

		if err != nil {
			fmt.Printf("Error: %s\n", err)
			numErrors++
			continue
		}

		prefix[moduleName].NodeCount++

		keyValueSize := len(fullKey) + len(fullValue)

		prefix[moduleName].TotalSizeBytes += keyValueSize
		totalSizeBytes += keyValueSize
	}

	if err := itr.Error(); err != nil {
		panic(err)
	}
	fmt.Printf("****************** Finished iterating over whole database ******************\n")
	printModuleStats(prefix, count, numErrors, totalSizeBytes)
}

func ModuleSizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module-sizes",
		Short: "print sizes of each module in the database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := cmd.Flags().GetInt(FlagPrintInterval)

			if err != nil {
				return err
			}

			if d <= 0 {
				return fmt.Errorf("print interval must be greater than 0")
			}

			pathToDB := args[0]

			fmt.Printf("Try to get Babylon module names\n")
			babylonApp := app.NewTmpBabylonApp()

			names := babylonApp.ModuleManager.ModuleNames()
			fmt.Printf("Got Babylon module names\n")
			for _, name := range names {
				fmt.Printf("Module name: %s\n", name)
			}
			fmt.Printf("Got Babylon module names\n")

			db, err := OpenDB(pathToDB)

			if err != nil {
				return err
			}

			PrintDBStats(db, names, d)

			return nil
		},
	}

	cmd.Flags().Int(FlagPrintInterval, 100000, "interval between printing databse stats")

	return cmd
}

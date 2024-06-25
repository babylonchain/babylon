package cmd

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"
)

const (
	FlagPrintInterval = "print-interval"
)

// r is a regular expression that matched the store key prefix
// we cannot use modules names direclty as sometimes module key != store key
// for example account module has store key "acc" and module key "auth"
var r, _ = regexp.Compile("s/k:[A-Za-z]+/")

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
	NodeCount      uint64
	TotalSizeBytes uint64
}

type GlobalStats struct {
	TotalNodeCount       uint64
	TotalSizeBytes       uint64
	UnknownStoreKeyCount uint64
	UnknownStoreKeySize  uint64
}

func extractStoreKey(fullKey string) string {
	return r.FindString(fullKey)
}

func printModuleStats(stats map[string]*ModuleStats, gs *GlobalStats) {
	fmt.Printf("****************** Printing module stats ******************\n")
	fmt.Printf("Total number of nodes in db: %d\n", gs.TotalNodeCount)
	fmt.Printf("Total size of database: %d bytes\n", gs.TotalSizeBytes)
	fmt.Printf("Total number of unknown storekeys: %d\n", gs.UnknownStoreKeyCount)
	fmt.Printf("Total size of unknown storekeys: %d bytes\n", gs.UnknownStoreKeySize)
	fmt.Printf("Fraction of unknown storekeys: %.3f\n", float64(gs.UnknownStoreKeySize)/float64(gs.TotalSizeBytes))
	for k, v := range stats {
		fmt.Printf("Store key %s:\n", k)
		fmt.Printf("Number of tree state nodes: %d\n", v.NodeCount)
		fmt.Printf("Total size of of module storage: %d bytes\n", v.TotalSizeBytes)
		fmt.Printf("Fraction of total size: %.3f\n", float64(v.TotalSizeBytes)/float64(gs.TotalSizeBytes))
	}
	fmt.Printf("****************** Printed stats for all Babylon modules ******************\n")
}

func PrintDBStats(db dbm.DB, printInterval int) {
	fmt.Printf("****************** Starting to iterate over whole database ******************\n")
	storeKeyStats := make(map[string]*ModuleStats)

	gs := GlobalStats{}

	itr, err := db.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("****************** Retrived database iterator ******************\n")

	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		gs.TotalNodeCount++
		if gs.TotalNodeCount%uint64(printInterval) == 0 {
			printModuleStats(storeKeyStats, &gs)
		}

		fullKey := itr.Key()
		fullValue := itr.Value()
		fullKeyString := string(fullKey)
		keyValueSize := uint64(len(fullKey) + len(fullValue))
		extractedStoreKey := extractStoreKey(fullKeyString)

		if extractedStoreKey == "" {
			gs.UnknownStoreKeyCount++
			gs.TotalSizeBytes += keyValueSize
			gs.UnknownStoreKeySize += keyValueSize
			continue
		}

		if _, ok := storeKeyStats[extractedStoreKey]; !ok {
			storeKeyStats[extractedStoreKey] = &ModuleStats{}
		}

		storeKeyStats[extractedStoreKey].NodeCount++
		storeKeyStats[extractedStoreKey].TotalSizeBytes += keyValueSize
		gs.TotalSizeBytes += keyValueSize
	}

	if err := itr.Error(); err != nil {
		panic(err)
	}
	fmt.Printf("****************** Finished iterating over whole database ******************\n")
	printModuleStats(storeKeyStats, &gs)
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

			db, err := OpenDB(pathToDB)

			if err != nil {
				return err
			}

			PrintDBStats(db, d)

			return nil
		},
	}

	cmd.Flags().Int(FlagPrintInterval, 100000, "interval between printing databse stats")

	return cmd
}

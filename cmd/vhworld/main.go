package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Inozuma/vhpackage"
)

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatalf("usage: %s world_file.fwl [world_file.db]", os.Args[0])
	}

	metaPath := flag.Arg(0)
	dbPath := flag.Arg(1)

	world, err := vhpackage.NewWorldFromFile(metaPath, dbPath)
	if err != nil {
		log.Fatalf("Failed to load world: %s", err)
	}

	jsondata, err := json.MarshalIndent(world, "", "  ")
	if err != nil {
		log.Fatalf("cannot encode world: %s", err)
	}
	fmt.Fprintln(os.Stdout, string(jsondata))
}

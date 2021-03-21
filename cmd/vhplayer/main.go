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
		log.Fatalf("usage: %s save.fcl", os.Args[0])
	}

	savePath := flag.Arg(0)

	playerProfile, err := vhpackage.NewPlayerProfileFromFile(savePath)
	if err != nil {
		log.Fatalf("Failed to load player save: %s", err)
	}

	jsondata, err := json.MarshalIndent(playerProfile, "", "  ")
	if err != nil {
		log.Fatalf("cannot encode player profile: %s", err)
	}
	fmt.Fprintln(os.Stdout, string(jsondata))
}

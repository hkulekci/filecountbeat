package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/hkulekci/filecountbeat/beater"
)

func main() {
	err := beat.Run("filecountbeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}

package beater

import (
	"fmt"
	"time"
	"io/ioutil"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/hkulekci/filecountbeat/config"
)

type Filecountbeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
}

type CountResult struct {
	File   int
	Folder int
}

func (a *CountResult) Add(v CountResult) {
	a.File += v.File
	a.Folder += v.Folder
}

func (a CountResult) String() string {
	return fmt.Sprintf("File : %d, Folder: %d", a.File, a.Folder)
}

func (bt *Filecountbeat) countFile(dirFile string) CountResult {
	result := CountResult{File: 0, Folder: 0}
	files, _ := ioutil.ReadDir(dirFile)
	for _, f := range files {
		if f.IsDir() {
			result.Folder += 1
			result.Add(bt.countFile(dirFile + "/" + f.Name()))
		} else {
			result.File += 1
		}
	}
	return result
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Filecountbeat{
		done:   make(chan struct{}),
		config: config,
	}
	return bt, nil
}

func (bt *Filecountbeat) Run(b *beat.Beat) error {
	logp.Info("filecountbeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		res := bt.countFile(bt.config.Path)

		event := common.MapStr{
			"@timestamp": common.Time(time.Now()),
			"type":       b.Name,
			"message":    res.String(),
			"fodler":     res.File,
			"file":       res.Folder,
		}
		bt.client.PublishEvent(event)
		logp.Info("Event sent")
	}
}

func (bt *Filecountbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

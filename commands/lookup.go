package commands

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/maliceio/malice/malice/database/elasticsearch"
	"github.com/maliceio/malice/malice/docker/client"
	"github.com/maliceio/malice/malice/docker/client/container"
	er "github.com/maliceio/malice/malice/errors"
	"github.com/maliceio/malice/plugins"
	"github.com/maliceio/malice/utils"
)

func cmdLookUp(hash string, logs bool) error {

	docker := client.NewDockerClient()

	// Check that ElasticSearch is running
	if _, running, _ := container.Running(docker, "elk"); !running {
		log.Error("ELK is NOT running, starting now...")
		elk, err := container.StartELK(docker, false)
		er.CheckError(err)
		eInfo, err := container.Inspect(docker, elk.ID)
		er.CheckError(err)
		er.CheckError(elasticsearch.TestConnection(eInfo.NetworkSettings.IPAddress))
	}

	// Setup ElasticSearch
	elasticsearch.InitElasticSearch()

	if plugins.InstalledPluginsCheck(docker) {
		log.Debug("All enabled plugins are installed.")
	} else {
		// Prompt user to install all plugins?
		fmt.Println("All enabled plugins not installed would you like to install them now? (yes/no)")
		fmt.Println("[Warning] This can take a while if it is the first time you have ran Malice.")
		if utils.AskForConfirmation() {
			plugins.UpdateAllPlugins(docker)
		}
	}

	/////////////////////////////
	// Write hash to the Database
	resp := elasticsearch.WriteHashToDatabase(hash)

	plugins.RunIntelPlugins(docker, hash, resp.Id, true)

	return nil
}

// APILookUp is an API wrapper for cmdLookUp
func APILookUp(hash string) error {
	return cmdLookUp(hash, false)
}

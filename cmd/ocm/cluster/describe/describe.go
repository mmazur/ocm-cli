/*
Copyright (c) 2019 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package describe

import (
	"bytes"
	"fmt"
	"os"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/spf13/cobra"

	clusterpkg "github.com/openshift-online/ocm-cli/pkg/cluster"
	"github.com/openshift-online/ocm-cli/pkg/config"
	"github.com/openshift-online/ocm-cli/pkg/dump"
)

var args struct {
	json   bool
	output bool
}

var Cmd = &cobra.Command{
	Use:   "describe CLUSTERID [--output] [--short]",
	Short: "Describe a cluster",
	Long:  "Get info about a cluster identified by its cluster ID",
	RunE:  run,
}

func init() {
	// Add flags to rootCmd:
	flags := Cmd.Flags()
	flags.BoolVar(
		&args.output,
		"output",
		false,
		"Output result into JSON file.",
	)
	flags.BoolVar(
		&args.json,
		"json",
		false,
		"Output the entire JSON structure",
	)
}

func run(cmd *cobra.Command, argv []string) error {

	if len(argv) != 1 {
		return fmt.Errorf("Expected exactly one cluster")
	}

	// Load the configuration file:
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("Can't load config file: %v", err)
	}
	if cfg == nil {
		return fmt.Errorf("Not logged in, run the 'login' command")
	}

	// Check that the configuration has credentials or tokens that haven't have expired:
	armed, err := cfg.Armed()
	if err != nil {
		return fmt.Errorf("Can't check if tokens have expired: %v", err)
	}
	if !armed {
		return fmt.Errorf("Tokens have expired, run the 'login' command")
	}

	// Create the connection, and remember to close it:
	connection, err := cfg.Connection()
	if err != nil {
		return fmt.Errorf("Can't create connection: %v", err)
	}
	defer connection.Close()

	// Get the client for the resource that manages the collection of clusters:
	resource := connection.ClustersMgmt().V1().Clusters()

	// Get the resource that manages the cluster that we want to display:
	clusterResource := resource.Cluster(argv[0])

	// Retrieve the collection of clusters:
	response, err := clusterResource.Get().Send()

	// Get cluster body:
	cluster := response.Body()

	if err != nil {
		return fmt.Errorf("Can't retrieve clusters: %s", err)
	}

	if args.output {
		// Create a filename based on cluster name:
		filename := fmt.Sprintf("cluster-%s.json", cluster.ID())

		// Attempt to create file:
		myFile, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("Failed to create file: %v", err)
		}

		// Dump encoder content into file:
		err = cmv1.MarshalCluster(cluster, myFile)
		if err != nil {
			return fmt.Errorf("Failed to Marshal cluster into file: %v", err)
		}
	}

	// Get full API response (JSON):
	if args.json {
		// Buffer for pretty output:
		buf := new(bytes.Buffer)
		fmt.Println()

		// Convert cluster to JSON and dump to encoder:
		err = cmv1.MarshalCluster(cluster, buf)
		if err != nil {
			return fmt.Errorf("Failed to Marshal cluster into JSON encoder: %v", err)
		}

		if response.Status() < 400 {
			err = dump.Pretty(os.Stdout, buf.Bytes())
		} else {
			err = dump.Pretty(os.Stderr, buf.Bytes())
		}
		if err != nil {
			return fmt.Errorf("Can't print body: %v", err)
		}

	} else {
		err = clusterpkg.PrintClusterDesctipion(connection, cluster)
		if err != nil {
			return err
		}
	}

	return nil
}

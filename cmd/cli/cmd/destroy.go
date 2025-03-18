/*
Copyright 2025 LuisEspinoza.

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

package cmd

import (
	"fmt"

	"github.com/espinozasenior/go-assesstment.git/pkg/auth"
	"github.com/espinozasenior/go-assesstment.git/pkg/client"
	"github.com/spf13/cobra"
)

var destroyName string

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy a deployment",
	Long:  `Delete a deployment from the system.`,
	Run: func(cmd *cobra.Command, args []string) {
		token, err := auth.GetToken()
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		// Create a new client with the API server URL and token
		c := client.NewClient("http://localhost:8080", token)

		// Destroy the deployment
		if err := c.DestroyDeployment(destroyName); err != nil {
			fmt.Printf("❌ Failed to destroy deployment: %v\n", err)
			return
		}

		fmt.Printf("❌ %s destroyed\n", destroyName)
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	destroyCmd.Flags().StringVar(&destroyName, "name", "", "Name of the deployment to destroy")
	if err := destroyCmd.MarkFlagRequired("name"); err != nil {
		fmt.Printf("Error marking name flag as required: %v\n", err)
	}
}

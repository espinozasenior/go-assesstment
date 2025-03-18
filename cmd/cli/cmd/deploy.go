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

var (
	image       string
	name        string
	memoryLimit string
	minReplicas int32
	maxReplicas int32
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy an application",
	Long:  `Deploy an application using the provided parameters.`,
	Run: func(cmd *cobra.Command, args []string) {
		token, err := auth.GetToken()
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		// Create a new client with the API server URL and token
		c := client.NewClient("http://localhost:8080", token)

		// Create the deploy request
		req := client.DeployRequest{
			Image:       image,
			Name:        name,
			MemoryLimit: memoryLimit,
			MinReplicas: minReplicas,
			MaxReplicas: maxReplicas,
		}

		fmt.Printf("📦 Deploying %s...\n", name)

		// Send the deploy request
		if err := c.Deploy(req); err != nil {
			fmt.Printf("❌ Deployment failed: %v\n", err)
			return
		}

		fmt.Println("✅ Deployment CRD created.")
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVar(&image, "image", "", "Container image to deploy")
	deployCmd.Flags().StringVar(&name, "name", "", "Name of the deployment")
	deployCmd.Flags().StringVar(&memoryLimit, "memoryLimit", "", "Memory limit for the deployment (e.g., 512Mi)")
	deployCmd.Flags().Int32Var(&minReplicas, "minReplicas", 1, "Minimum number of replicas")
	deployCmd.Flags().Int32Var(&maxReplicas, "maxReplicas", 3, "Maximum number of replicas")

	if err := deployCmd.MarkFlagRequired("image"); err != nil {
		fmt.Printf("Error marking image flag as required: %v\n", err)
	}
	if err := deployCmd.MarkFlagRequired("name"); err != nil {
		fmt.Printf("Error marking name flag as required: %v\n", err)
	}
	if err := deployCmd.MarkFlagRequired("memoryLimit"); err != nil {
		fmt.Printf("Error marking memoryLimit flag as required: %v\n", err)
	}
}

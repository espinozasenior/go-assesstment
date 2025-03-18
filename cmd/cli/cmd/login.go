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
	"github.com/spf13/cobra"
)

var (
	username string
	password string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the API",
	Long:  `Login to the API using your username and password.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := auth.Login(username, password); err != nil {
			fmt.Printf("❌ Login failed: %v\n", err)
			return
		}

		fmt.Println("✅ Login successful. Token stored.")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVar(&username, "username", "", "Username for authentication")
	loginCmd.Flags().StringVar(&password, "password", "", "Password for authentication")

	if err := loginCmd.MarkFlagRequired("username"); err != nil {
		fmt.Printf("Error marking username flag as required: %v\n", err)
	}
	if err := loginCmd.MarkFlagRequired("password"); err != nil {
		fmt.Printf("Error marking password flag as required: %v\n", err)
	}
}

package cmd

import (
	"fmt"
	"os"

	"github.com/luponetn/insighta-cli/utils"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user profile",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			fmt.Println("Error: --name is required")
			return
		}

		// 1. Load Session
		cfg, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("\nYou are not logged in. Please run 'login' first.")
			return
		}

		// 2. Create Profile
		stopSpinner := utils.StartSpinner(fmt.Sprintf("Creating profile for '%s'...", name))
		
		backendBaseUrl := os.Getenv("DEV_BACKEND_BASE_URL")
		if backendBaseUrl == "" {
			backendBaseUrl = "http://localhost:7000" // Fallback
		}

		body := map[string]string{
			"name": name,
		}

		respData, err := utils.MakeRequest(utils.RequestOptions{
			Method: "POST",
			URL:    backendBaseUrl + "/api/profiles",
			Body:   body,
			Token:  cfg.AccessToken,
		})
		stopSpinner()

		if err != nil {
			fmt.Printf("\nFailed to create profile: %v\n", err)
			return
		}

		// 3. Process Data
		data, ok := respData["data"].(map[string]any)
		if !ok {
			fmt.Println("\nError: Unexpected response format.")
			return
		}

		// 4. Display Results
		fmt.Printf("\nProfile created successfully!\n")
		fmt.Printf("ID:          %v\n", data["id"])
		fmt.Printf("Name:        %v\n", data["name"])
		fmt.Printf("Gender:      %v\n", data["gender"])
		fmt.Printf("Age:         %v\n", data["age"])
		fmt.Printf("Country:     %v\n", data["country_name"])
		fmt.Println("-----------------------")
	},
}

func init() {
	profilesCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("name", "n", "", "Name of the profile to create")
	createCmd.MarkFlagRequired("name")
}

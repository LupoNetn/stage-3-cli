package cmd

import (
	"fmt"
	"os"

	"github.com/luponetn/insighta-cli/utils"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all user profiles",
	Long: `The list command fetches a paginated list of all profiles 
stored in the system. It displays key information like ID, Name, and 
Location for each profile.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Load Session
		cfg, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("\nYou are not logged in. Please run 'login' first.")
			return
		}

		// 2. Fetch Profiles
		stopSpinner := utils.StartSpinner("Fetching profiles...")
		
		backendBaseUrl := os.Getenv("DEV_BACKEND_BASE_URL")
		if backendBaseUrl == "" {
			backendBaseUrl = "http://localhost:7000" // Fallback
		}

		respData, err := utils.MakeRequest(utils.RequestOptions{
			Method: "GET",
			URL:    backendBaseUrl + "/api/profiles",
			Token:  cfg.AccessToken,
		})
		stopSpinner()

		if err != nil {
			fmt.Printf("\nFailed to fetch profiles: %v\n", err)
			return
		}

		// 3. Process Data
		data, ok := respData["data"].([]any)
		if !ok {
			fmt.Println("\nError: Unexpected response format (missing data array).")
			return
		}

		if len(data) == 0 {
			fmt.Println("\nNo profiles found.")
			return
		}

		// 4. Display Results
		fmt.Println("\n--- Profiles List ---")
		fmt.Printf("%-38s | %-20s | %-15s\n", "ID", "Name", "Country")
		fmt.Println("---------------------------------------------------------------------------")

		for _, item := range data {
			p := item.(map[string]any)
			fmt.Printf("%-38v | %-20v | %-15v\n", 
				p["id"], 
				p["name"], 
				p["country_name"],
			)
		}

		fmt.Println("---------------------------------------------------------------------------")
		fmt.Printf("Page: %v/%v | Total Profiles: %v\n", 
			respData["page"], 
			respData["total_pages"], 
			respData["total"],
		)
	},
}

func init() {
	profilesCmd.AddCommand(listCmd)
}

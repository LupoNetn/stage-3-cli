package cmd

import (
	"fmt"
	"net/url"
	"os"

	"github.com/luponetn/insighta-cli/utils"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search profiles using natural language",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		// 1. Load Session
		cfg, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("\nYou are not logged in. Please run 'login' first.")
			return
		}

		// 2. Search Profiles
		stopSpinner := utils.StartSpinner(fmt.Sprintf("Searching for '%s'...", query))
		
		backendBaseUrl := os.Getenv("DEV_BACKEND_BASE_URL")
		if backendBaseUrl == "" {
			backendBaseUrl = "http://localhost:7000" // Fallback
		}

		queryParams := url.Values{}
		queryParams.Add("q", query)

		respData, err := utils.MakeRequest(utils.RequestOptions{
			Method: "GET",
			URL:    backendBaseUrl + "/api/profiles/search?" + queryParams.Encode(),
			Token:  cfg.AccessToken,
		})
		stopSpinner()

		if err != nil {
			fmt.Printf("\nSearch failed: %v\n", err)
			return
		}

		// 3. Process Data
		data, ok := respData["data"].([]any)
		if !ok {
			fmt.Println("\nError: Unexpected response format.")
			return
		}

		if len(data) == 0 {
			fmt.Println("\nNo profiles matched your search.")
			return
		}

		// 4. Display Results
		fmt.Printf("\n--- Search Results for '%s' ---\n", query)
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
		fmt.Printf("Results: %v | Total Matches: %v\n", 
			len(data),
			respData["total"],
		)
	},
}

func init() {
	profilesCmd.AddCommand(searchCmd)
}

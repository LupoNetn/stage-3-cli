package cmd

import (
	"fmt"
	"os"

	"github.com/luponetn/insighta-cli/utils"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get a specific user profile by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		// 1. Load Session
		cfg, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("\nYou are not logged in. Please run 'login' first.")
			return
		}

		// 2. Fetch Profile
		stopSpinner := utils.StartSpinner(fmt.Sprintf("Fetching profile %s...", id))
		
		backendBaseUrl := os.Getenv("DEV_BACKEND_BASE_URL")
		if backendBaseUrl == "" {
			backendBaseUrl = "https://stage-3-backend-azure.vercel.app"
		}

		respData, err := utils.MakeRequest(utils.RequestOptions{
			Method: "GET",
			URL:    fmt.Sprintf("%s/api/profiles/%s", backendBaseUrl, id),
			Token:  cfg.AccessToken,
		})
		stopSpinner()

		if err != nil {
			fmt.Printf("\nFailed to fetch profile: %v\n", err)
			return
		}

		// 3. Process Data
		data, ok := respData["data"].(map[string]any)
		if !ok {
			fmt.Println("\nError: Unexpected response format.")
			return
		}

		// 4. Display Results
		fmt.Printf("\n--- Profile Details ---\n")
		fmt.Printf("ID:          %v\n", data["id"])
		fmt.Printf("Name:        %v\n", data["name"])
		fmt.Printf("Gender:      %v (%v%%)\n", data["gender"], formatProb(data["gender_probability"]))
		fmt.Printf("Age:         %v (%v)\n", data["age"], data["age_group"])
		fmt.Printf("Country:     %v (%v) (%v%%)\n", data["country_name"], data["country_id"], formatProb(data["country_probability"]))
		fmt.Printf("Created At:  %v\n", data["created_at"])
		fmt.Println("-----------------------")
	},
}

func formatProb(v any) string {
	if f, ok := v.(float64); ok {
		return fmt.Sprintf("%.0f", f*100)
	}
	return "0"
}

func init() {
	profilesCmd.AddCommand(getCmd)
}

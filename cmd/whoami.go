package cmd

import (
	"fmt"
	"os"

	"github.com/luponetn/insighta-cli/utils"
	"github.com/spf13/cobra"
)

// whoamiCmd represents the whoami command
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display information about the currently logged-in user",
	Long: `The whoami command fetches your profile details from the backend 
to verify your authentication status and display your account information.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Load Session
		cfg, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("\nYou are not logged in. Please run 'login' first.")
			return
		}

		// 2. Fetch Profile from Backend
		stopSpinner := utils.StartSpinner("Fetching your profile...")
		
		backendBaseUrl := os.Getenv("DEV_BACKEND_BASE_URL")
		if backendBaseUrl == "" {
			backendBaseUrl = "http://localhost:7000" // Fallback
		}

		respData, err := utils.MakeRequest(utils.RequestOptions{
			Method: "GET",
			URL:    backendBaseUrl + "/auth/me",
			Token:  cfg.AccessToken,
		})
		stopSpinner()

		if err != nil {
			fmt.Printf("\nFailed to fetch profile: %v\n", err)
			return
		}

		// 3. Display Profile
		data, ok := respData["data"].(map[string]any)
		if !ok {
			fmt.Println("\nError: Unexpected response format from server.")
			return
		}

		fmt.Println("\n--- Current User Profile ---")
		fmt.Printf("ID:         %v\n", data["id"])
		fmt.Printf("Username:   %v\n", data["username"])
		fmt.Printf("Email:      %v\n", data["email"])
		fmt.Printf("Role:       %v\n", data["role"])
		if avatar, ok := data["avatar_url"].(string); ok && avatar != "" {
			fmt.Printf("Avatar:     %v\n", avatar)
		}
		fmt.Println("----------------------------")
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

package cmd

import (
	"fmt"
	"os"

	"github.com/luponetn/insighta-cli/utils"
	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of your account and clear local session data",
	Long: `The logout command removes your local authentication tokens and 
session data from your machine. Once logged out, you will need to run 
'login' again to access protected features.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Check if logged in
		cfg, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("\nYou are not currently logged in.")
			return
		}

		// 2. Perform Logout
		stopSpinner := utils.StartSpinner("Logging you out and clearing session...")
		
		// Invalidate on backend
		backendUrl := os.Getenv("DEV_BACKEND_BASE_URL")
		if backendUrl == "" {
			backendUrl = "https://stage-3-backend-azure.vercel.app"
		}

		// Call logout endpoint (it expects refresh_token in JSON body)
		utils.MakeRequest(utils.RequestOptions{
			Method: "POST",
			URL:    backendUrl + "/auth/logout",
			Body:   map[string]string{"refresh_token": cfg.RefreshToken},
			Token:  cfg.AccessToken,
		})

		// Clear local data regardless of backend success
		err = utils.ClearConfig()
		stopSpinner()

		if err != nil {
			fmt.Printf("\nError during logout: %v\n", err)
			return
		}

		fmt.Println("\nSuccessfully logged out. Session data invalidated.")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

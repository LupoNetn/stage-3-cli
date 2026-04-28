package cmd

import (
	"fmt"
	"time"

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
		_, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("\nYou are not currently logged in.")
			return
		}

		// 2. Perform Logout with Spinner
		stopSpinner := utils.StartSpinner("Logging you out and clearing session...")
		time.Sleep(500 * time.Millisecond)

		err = utils.ClearConfig()
		stopSpinner()

		if err != nil {
			fmt.Printf("\nError during logout: %v\n", err)
			return
		}

		fmt.Println("\nSuccessfully logged out. Session data cleared.")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

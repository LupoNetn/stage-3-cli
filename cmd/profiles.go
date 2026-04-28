package cmd

import (
	"github.com/spf13/cobra"
)

// profilesCmd represents the profiles command
var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage and search user profiles",
	Long: `The profiles command provides a suite of tools to interact 
with the profile management system, including listing, searching, 
and exporting profile data.`,
}

func init() {
	rootCmd.AddCommand(profilesCmd)
}

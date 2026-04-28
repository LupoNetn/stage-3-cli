package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/luponetn/insighta-cli/utils"
	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export profiles to a file",
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		if format == "" {
			fmt.Println("Error: --format is required (e.g. csv)")
			return
		}

		// 1. Load Session
		cfg, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("\nYou are not logged in. Please run 'login' first.")
			return
		}

		// 2. Prepare Query Params
		queryParams := url.Values{}
		queryParams.Add("format", format)
		
		if gender, _ := cmd.Flags().GetString("gender"); gender != "" {
			queryParams.Add("gender", gender)
		}
		if country, _ := cmd.Flags().GetString("country"); country != "" {
			queryParams.Add("country", country)
		}
		if minAge, _ := cmd.Flags().GetInt("min-age"); minAge != 0 {
			queryParams.Add("min_age", strconv.Itoa(minAge))
		}
		if maxAge, _ := cmd.Flags().GetInt("max-age"); maxAge != 0 {
			queryParams.Add("max_age", strconv.Itoa(maxAge))
		}
		if ageGroup, _ := cmd.Flags().GetString("age-group"); ageGroup != "" {
			queryParams.Add("age_group", ageGroup)
		}
		if sortBy, _ := cmd.Flags().GetString("sort-by"); sortBy != "" {
			queryParams.Add("sort_by", sortBy)
		}
		if order, _ := cmd.Flags().GetString("order"); order != "" {
			queryParams.Add("order", order)
		}
		if page, _ := cmd.Flags().GetInt("page"); page != 0 {
			queryParams.Add("page", strconv.Itoa(page))
		}
		if limit, _ := cmd.Flags().GetInt("limit"); limit != 0 {
			queryParams.Add("limit", strconv.Itoa(limit))
		}

		// 3. Export Profiles
		stopSpinner := utils.StartSpinner("Exporting profiles...")
		
		backendBaseUrl := os.Getenv("DEV_BACKEND_BASE_URL")
		if backendBaseUrl == "" {
			backendBaseUrl = "http://localhost:7000" // Fallback
		}

		data, filename, err := utils.DownloadFile(utils.RequestOptions{
			Method: "GET",
			URL:    backendBaseUrl + "/api/profiles/export?" + queryParams.Encode(),
			Token:  cfg.AccessToken,
		})
		stopSpinner()

		if err != nil {
			fmt.Printf("\nExport failed: %v\n", err)
			return
		}

		// 4. Save to Disk
		err = os.WriteFile(filename, data, 0644)
		if err != nil {
			fmt.Printf("\nFailed to save file: %v\n", err)
			return
		}

		fmt.Printf("\nProfiles exported successfully to %s\n", filename)
	},
}

func init() {
	profilesCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringP("format", "f", "csv", "Export format (currently only 'csv' is supported)")
	exportCmd.Flags().StringP("gender", "g", "", "Filter by Gender")
	exportCmd.Flags().StringP("country", "c", "", "Filter by Country")
	exportCmd.Flags().IntP("min-age", "m", 0, "Minimum Age")
	exportCmd.Flags().IntP("max-age", "M", 0, "Maximum Age")
	exportCmd.Flags().StringP("age-group", "A", "", "Filter by Age Group")
	exportCmd.Flags().StringP("sort-by", "s", "created_at", "Sort field")
	exportCmd.Flags().StringP("order", "o", "desc", "Sort order")
	exportCmd.Flags().IntP("page", "p", 0, "Page number to export (0 for all, if supported by backend)")
	exportCmd.Flags().IntP("limit", "l", 0, "Number of profiles to export")
}

package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

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
			backendBaseUrl = "https://stage-3-backend-azure.vercel.app"
		}

		queryParams := url.Values{}
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
		if page, _ := cmd.Flags().GetInt("page"); page != 0 {
			queryParams.Add("page", strconv.Itoa(page))
		}
		if limit, _ := cmd.Flags().GetInt("limit"); limit != 0 {
			queryParams.Add("limit", strconv.Itoa(limit))
		}
		if sortBy, _ := cmd.Flags().GetString("sort-by"); sortBy != "" {
			queryParams.Add("sort_by", sortBy)
		}
		if order, _ := cmd.Flags().GetString("order"); order != "" {
			queryParams.Add("order", order)
		}

		respData, err := utils.MakeRequest(utils.RequestOptions{
			Method: "GET",
			URL:    backendBaseUrl + "/api/profiles?" + queryParams.Encode(),
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

	listCmd.Flags().StringP("gender", "g", "", "Filter by Gender (male, female)")
	listCmd.Flags().StringP("country", "c", "", "Filter by Country name or ID (e.g. Nigeria or NG)")
	listCmd.Flags().IntP("min-age", "m", 0, "Minimum Age")
	listCmd.Flags().IntP("max-age", "M", 0, "Maximum Age")
	listCmd.Flags().StringP("age-group", "A", "", "Filter by Age Group (child, teenager, adult, senior)")
	listCmd.Flags().IntP("page", "p", 1, "Page number")
	listCmd.Flags().IntP("limit", "l", 10, "Number of profiles per page")
	listCmd.Flags().StringP("sort-by", "s", "created_at", "Sort profiles by a field (age, created_at, gender_probability)")
	listCmd.Flags().StringP("order", "o", "desc", "Sort order (asc, desc)")
}

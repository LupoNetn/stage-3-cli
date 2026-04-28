package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/luponetn/insighta-cli/utils"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with your GitHub account",
	Long: `The login command initiates an OAuth2 flow with GitHub. It will 
open your default browser for authorization and securely store your 
access tokens locally for future use.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Generate OAuth PKCE params
		state, codeVerifier, codeChallenge, err := utils.GeneratePKCEParams(32, 32)
		if err != nil {
			fmt.Println("\nError generating security params:", err)
			return
		}

		// 2. Prepare Auth URL
		githubClientId := os.Getenv("GITHUB_CLIENT_ID")
		githubRedirectUrl := os.Getenv("GITHUB_REDIRECT_URL")
		githubAuthUrl := os.Getenv("GITHUB_OAUTH_AUTHORIZE_URL")

		authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=user:email&state=%s&code_challenge=%s&code_challenge_method=S256", 
			githubAuthUrl, githubClientId, githubRedirectUrl, state, codeChallenge)
	    
		// 3. Open Browser
		if err := utils.OpenBrowser(authURL); err != nil {
			fmt.Println("\nError opening browser:", err)
			return
		}

		fmt.Println("\nPlease authorize the application in your browser and return here.")

		// 4. Start Local Callback Server
		stopSpinner := utils.StartSpinner("Waiting for browser authentication...")
		code, err := StartCallbackServer(state)
		stopSpinner()
		
		if err != nil {
			fmt.Println("\nError during authentication flow:", err)
			return
		}

		// 5. Exchange code for token with Backend
		stopSpinner = utils.StartSpinner("Exchanging token with backend...")
		
		backendUrl := os.Getenv("DEV_BACKEND_GITHUB_AUTH_URL")
		respData, err := utils.MakeRequest(utils.RequestOptions{
			Method: "POST",
			URL:    backendUrl,
			Body: map[string]string{
				"code":          code,
				"code_verifier": codeVerifier,
				"state":         state,
			},
		})
		stopSpinner()

		if err != nil {
			fmt.Printf("\nAuthentication failed: %v\n", err)
			return
		}
		
		// 6. Store Session using Config Utility
		cfg := utils.Config{
			AccessToken:  fmt.Sprintf("%v", respData["access_token"]),
			RefreshToken: fmt.Sprintf("%v", respData["refresh_token"]),
			Username:     fmt.Sprintf("%v", respData["username"]),
		}
		
		if err := utils.SaveConfig(cfg); err != nil {
			fmt.Println("\nError saving session data:", err)
			return
		}

		fmt.Printf("\nLogged in successfully as @%v\n", cfg.Username)
	},
}

func StartCallbackServer(expectedState string) (string, error) {
	codeChan := make(chan string)
	errChan := make(chan error)

	server := &http.Server{Addr: ":8484"}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != expectedState {
			w.WriteHeader(http.StatusBadRequest)
			errChan <- fmt.Errorf("invalid state")
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			errChan <- fmt.Errorf("empty code")
			return 
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<h2>Authentication Successful!</h2><p>You can close this tab and return to the CLI.</p>"))
		codeChan <- code
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case code := <- codeChan:
		server.Close()
		return code, nil
	case err := <- errChan:
		server.Close()
		return "", err
	}
} 

func init() {
	rootCmd.AddCommand(loginCmd)
}

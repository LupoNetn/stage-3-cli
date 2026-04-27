/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		//generate oauth pkce params
		state, _, code_challenge, err := GeneratePKCEParams(32, 32)
		if err != nil {
			fmt.Println("Error generating PKCE params:", err)
			return
		}

		githubClientId := os.Getenv("GITHUB_CLIENT_ID")
		githubRedirectUrl := os.Getenv("GITHUB_REDIRECT_URL")
		githubOAuthUrl := os.Getenv("GITHUB_OAUTH_URL")

		authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=user:email&state=%s&code_challenge=%s&code_challenge_method=S256", githubOAuthUrl, githubClientId, githubRedirectUrl, state, code_challenge)
	    
		err = OpenBrowser(authURL)
		if err != nil {
			fmt.Println("Error opening browser:", err)
			return
		}

		fmt.Println("Please authorize the application in your browser and return here.")

		code, err := StartCallbackServer(state)
		if err != nil {
			fmt.Println("Error starting callback server for logging you in")
			return
		}
		fmt.Println("github auth code : ", code)
	},
}

func StartCallbackServer(expectedState string) (string, error) {
	codeChan := make(chan string)
	errChan := make(chan error)

	mux := http.NewServeMux()

	server := &http.Server{
		Addr: ":8484",
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state != expectedState {
			fmt.Println("Invalid state! login failed you can close this window now")
			errChan <- fmt.Errorf("invalid state")
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			fmt.Println("Login failed, retry. you can close this window now")
			errChan <- fmt.Errorf("empty code")
			return 
		}

		codeChan <- code
	})

	go func() {
		fmt.Println("callback server is running at localhost:8484")
		if err := server.ListenAndServe(); err != nil {
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

func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func GeneratePKCEParams(stateInt, verifierInt int) (string, string, string, error) {
	state, err := GenerateRandomString(stateInt)
	if err != nil {
		return "", "", "", err
	}

	code_verifier, err := GenerateRandomString(verifierInt)
	if err != nil {
		return "", "", "", err
	}

	code_challenge := sha256.Sum256([]byte(code_verifier))
	code_challengeStr := base64.RawURLEncoding.EncodeToString(code_challenge[:])

	return state, code_verifier, code_challengeStr, nil
}

func OpenBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return err
}




func init() {
	rootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

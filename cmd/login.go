/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

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
		state, codeVerifier, code_challenge, err := GeneratePKCEParams(32, 32)
		if err != nil {
			fmt.Println("Error generating PKCE params:", err)
			return
		}

		githubClientId := os.Getenv("GITHUB_CLIENT_ID")
		githubRedirectUrl := os.Getenv("GITHUB_REDIRECT_URL")
		githubOAuthAuthorizationUrl := os.Getenv("GITHUB_OAUTH_AUTHORIZE_URL")

		authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=user:email&state=%s&code_challenge=%s&code_challenge_method=S256", githubOAuthAuthorizationUrl, githubClientId, githubRedirectUrl, state, code_challenge)
	    
		err = OpenBrowser(authURL)
		if err != nil {
			fmt.Println("Error opening browser:", err)
			return
		}

		fmt.Println("Please authorize the application in your browser and return here.")

		stopSpinner := startSpinner("Waiting for browser authentication...")
		code, err := StartCallbackServer(state)
		stopSpinner()
		
		if err != nil {
			fmt.Println("\nError starting callback server for logging you in")
			return
		}

		stopSpinner = startSpinner("Exchanging token with backend...")
		respData, err := SendRequest(code, codeVerifier, state)
		stopSpinner()
		if err != nil {
			fmt.Println("Error exchanging code for token:", err)
			return
		}
		
		// Store Tokens
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configDir := filepath.Join(homeDir, ".insighta")
			os.MkdirAll(configDir, 0755)
			
			configPath := filepath.Join(configDir, "config.json")
			configData := map[string]any{
				"access_token":  respData["access_token"],
				"refresh_token": respData["refresh_token"],
				"username":      respData["username"],
			}
			
			file, err := os.Create(configPath)
			if err == nil {
				json.NewEncoder(file).Encode(configData)
				file.Close()
			}
		}

		// Success message
		fmt.Printf("\nLogged in as @%v\n", respData["username"])
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
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid state. You can close this window."))
			errChan <- fmt.Errorf("invalid state")
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Empty code. You can close this window."))
			errChan <- fmt.Errorf("empty code")
			return 
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
	<title>Authentication Successful</title>
	<script>
		setTimeout(function() {
			window.close();
		}, 2000);
	</script>
</head>
<body style="font-family: Arial, sans-serif; text-align: center; margin-top: 50px;">
	<h2>Authentication Successful!</h2>
	<p>You can safely close this tab and return to the CLI.</p>
</body>
</html>
		`))

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

func SendRequest(code, codeVerifier, state string) (map[string]any, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
    
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
    
	backendUrl := os.Getenv("DEV_BACKEND_GITHUB_AUTH_URL")
	if backendUrl == "" {
		return nil, fmt.Errorf("DEV_BACKEND_GITHUB_AUTH_URL not set")
	}
	
	body := struct {
		Code         string `json:"code"`
		CodeVerifier string `json:"code_verifier"`
		State        string `json:"state"`
	}{
		Code:         code,
		CodeVerifier: codeVerifier,
		State:        state,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, backendUrl,bytes.NewBuffer(bodyBytes)) 
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("login failed: %s", resp.Status)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
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

func startSpinner(msg string) func() {
	stop := make(chan struct{})
	go func() {
		chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-stop:
				fmt.Printf("\r\033[K") // clear line
				return
			default:
				fmt.Printf("\r%s %s", chars[i], msg)
				i = (i + 1) % len(chars)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	return func() {
		close(stop)
	}
}

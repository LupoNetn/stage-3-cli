package utils

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// StartSpinner starts a terminal spinner with a message
func StartSpinner(msg string) func() {
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

// OpenBrowser opens the specified URL in the default system browser
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

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"lockbox/internal/server"
)

var (
	uiPort        int
	uiNoOpen      bool
	uiLockTimeout time.Duration
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Start the local web UI",
	Long: "Start a local web UI for lockbox.\n" +
		"Binds to 127.0.0.1 only and opens your browser.\n" +
		"The vault session auto-locks after inactivity (--lock-timeout, default 15m).",
	Example: `  lockbox ui
  lockbox ui --port 9000
  lockbox ui --no-open --lock-timeout 5m`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, err := server.New(server.Options{
			Port:        uiPort,
			LockTimeout: uiLockTimeout,
		})
		if err != nil {
			return fmt.Errorf("🚨 %w", err)
		}
		ln, err := srv.Start()
		if err != nil {
			return fmt.Errorf("🚨 Failed to start UI server: %w", err)
		}
		defer srv.Shutdown()

		url := "http://" + srv.Addr()
		fmt.Printf("🔒 lockbox UI running at %s\n", url)
		fmt.Printf("   Auto-locks after %s of inactivity.\n", uiLockTimeout)
		if !uiNoOpen {
			openBrowser(url)
		}
		fmt.Println("   Press Ctrl+C to stop.")

		if err := srv.Serve(ln); err != nil {
			return fmt.Errorf("🚨 UI server stopped: %w", err)
		}
		return nil
	},
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		if _, err := exec.LookPath("xdg-open"); err == nil {
			cmd = exec.Command("xdg-open", url)
		} else {
			return
		}
	}
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not open browser automatically — visit %s\n", url)
	}
}

func init() {
	uiCmd.Flags().IntVar(&uiPort, "port", 8787, "port to listen on (127.0.0.1)")
	uiCmd.Flags().BoolVar(&uiNoOpen, "no-open", false, "do not open the browser")
	uiCmd.Flags().DurationVar(&uiLockTimeout, "lock-timeout", 15*time.Minute, "auto-lock after inactivity")
	rootCmd.AddCommand(uiCmd)
}

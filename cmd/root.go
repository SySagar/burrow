package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "burrow [PORT]",
		Short: "Expose your local app to the internet via tunneling",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			port := args[0]
			if !isPortActive(port) {
				fmt.Println("❌ Port", port, "is not active on your system")
				os.Exit(1)
			}

			fmt.Println("✅ Port", port, "is active. Starting tunnel...")
			runClient(port)
		},
	}
	rootCmd.Execute()
}

func isPortActive(port string) bool {
	out, err := exec.Command("netstat", "-an").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), ":"+port)
}

func runClient(port string) {
	cmd := exec.Command("go", "run", "./client/client.go", port)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

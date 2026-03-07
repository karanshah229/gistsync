package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/karanshah229/gistsync/internal"
	"github.com/karanshah229/gistsync/providers"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gistsync configurations and state",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if internal.IsConfigPresent() {
			fmt.Print("⚠️  Configuration already exists. Overwrite? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				fmt.Println("Aborted.")
				return
			}
		}

		fmt.Println("🚀 Initializing gistsync...")

		config := internal.DefaultConfig()
		reader := bufio.NewReader(os.Stdin)

		// 1. Check Providers
		fmt.Println("\n🔍 Checking Providers...")
		gh := providers.NewGitHubProvider()
		gl := providers.NewGitLabProvider()

		ghOk, ghMsg, _ := gh.Verify()
		glOk, glMsg, _ := gl.Verify()

		availableProviders := []string{}

		if ghOk {
			fmt.Printf("✅ GitHub: Connected (%s)\n", strings.TrimSpace(ghMsg))
			config.Providers["github"] = "connected"
			availableProviders = append(availableProviders, "github")
		} else {
			fmt.Printf("❌ GitHub: Not Connected (%s)\n", strings.TrimSpace(ghMsg))
			config.Providers["github"] = "not connected"
		}

		if glOk {
			fmt.Printf("✅ GitLab: Connected (%s)\n", strings.TrimSpace(glMsg))
			config.Providers["gitlab"] = "connected"
			availableProviders = append(availableProviders, "gitlab")
		} else {
			fmt.Printf("❌ GitLab: Not Connected (%s)\n", strings.TrimSpace(glMsg))
			config.Providers["gitlab"] = "not connected"
		}

		if !ghOk && !glOk {
			fmt.Println("\n💡 No providers are connected.")
			showProviderInfo()
		}

		// 2. Select Default Provider
		fmt.Println("\n🎯 Default Provider Selection")
		fmt.Println("   The default provider is used for:")
		fmt.Println("   - Fast sync: used when no provider is specified in commands.")
		fmt.Println("   - Backup: your configuration and state will be backed up to this provider.")

		providerOptions := []string{"github", "gitlab"}
		// If we found connected ones, maybe we want to highlight them? 
		// For now, allow both but user can pick.

		var selectedProvider string
		prompt := &survey.Select{
			Message: "Select your default provider:",
			Options: providerOptions,
			Default: "github",
		}

		if err := survey.AskOne(prompt, &selectedProvider); err != nil {
			fmt.Printf("❌ Selection failed: %v. Defaulting to github.\n", err)
			selectedProvider = "github"
		}
		config.DefaultProvider = selectedProvider
		fmt.Printf("✅ Selected: %s\n", selectedProvider)

		// 3. Interactive Config
		fmt.Println("\n⚙️  Configuration Setup")
		options := internal.GetConfigOptions()
		for _, opt := range options {
			for {
				fmt.Printf("%s [%v]: ", opt.Prompt, opt.Default)
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)

				if input == "" {
					// Use default
					setField(config, opt.Key, opt.Default)
					break
				}

				if input == "?" || strings.ToLower(input) == "help" {
					fmt.Printf("   💡 %s\n", opt.Description)
					continue
				}

				// Try to parse input
				if err := updateField(config, opt.Key, input); err != nil {
					fmt.Printf("   ❌ Invalid input: %v. Please try again or press ENTER for default.\n", err)
					continue
				}
				break
			}
		}

		// 3. Save Config
		if err := internal.SaveConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to save config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n✅ Configuration saved to config.json")

		// 4. Initialize state.json
		statePath, err := internal.GetStateFilePath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to get state path: %v\n", err)
			os.Exit(1)
		}

		// Get version from root command or file
		version := Version
		if version == "" {
			version = "unknown"
		}

		initialState := struct {
			Version  string        `json:"version"`
			Mappings []interface{} `json:"mappings"`
		}{
			Version:  version,
			Mappings: []interface{}{},
		}

		data, _ := json.MarshalIndent(initialState, "", "  ")
		if err := internal.WriteAtomic(statePath, data); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to initialize state: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ State initialized in state.json")
		fmt.Println("\n🎉 gistsync is ready! Use 'gistsync sync <path>' to start syncing.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

// Helper to set field via reflection or simple switch
func setField(config *internal.Config, key string, val interface{}) {
	switch key {
	case "WatchInterval":
		config.WatchInterval = val.(int)
	case "WatchDebounce":
		config.WatchDebounce = val.(int)
	case "LogLevel":
		config.LogLevel = val.(string)
	}
}

// Helper to update field from string input
func updateField(config *internal.Config, key string, input string) error {
	switch key {
	case "WatchInterval":
		v, err := strconv.Atoi(input)
		if err != nil {
			return err
		}
		config.WatchInterval = v
	case "WatchDebounce":
		v, err := strconv.Atoi(input)
		if err != nil {
			return err
		}
		config.WatchDebounce = v
	case "LogLevel":
		allowed := []string{"debug", "info", "warn", "error"}
		if !slices.Contains(allowed, strings.ToLower(input)) {
			return fmt.Errorf("must be one of: %s", strings.Join(allowed, ", "))
		}
		config.LogLevel = strings.ToLower(input)
	}
	return nil
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// sysinfoCmd represents the sysinfo command
var sysinfoCmd = &cobra.Command{
	Use:   "sysinfo",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := printToolVersions(); err != nil {
			return fmt.Errorf("Failed to print tool versions - missing tool - error: %s", err)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(sysinfoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sysinfoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sysinfoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

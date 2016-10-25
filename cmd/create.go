package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-tools/replica/macosinstaller"
	"github.com/bitrise-tools/replica/vagrantbox"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create INSTALL_MACOS_APP_PATH",
	Short: `Create a vagrant box from an "Install macOS / OS X .." app`,
	Long:  `Create a vagrant box from an "Install macOS / OS X .." app`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("No 'Install macOS / OS X .. app' path provided")
		}
		installMacOSAppPath := args[0]
		return createVagrantBoxFromInstallMacOSApp(installMacOSAppPath)
	},
}

func init() {
	RootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func runToolVersionCommand(toolCmd string, toolCmdArgs ...string) (string, error) {
	return cmdex.NewCommand(toolCmd, toolCmdArgs...).RunAndReturnTrimmedCombinedOutput()
}

func printToolVersions() error {
	fmt.Println()
	fmt.Println("---------- TOOL VERSIONS: ----------")
	fmt.Println()

	{
		out, err := runToolVersionCommand("vboxmanage", "--version")
		if err != nil {
			return fmt.Errorf("Failed to get VirtualBox version, error: %s", err)
		}
		fmt.Println(colorstring.Green("* VirtualBox version:"))
		fmt.Println(out)
	}
	{
		out, err := runToolVersionCommand("vagrant", "version")
		if err != nil {
			return fmt.Errorf("Failed to get vagrant version, error: %s", err)
		}
		fmt.Println()
		fmt.Println(colorstring.Green("* vagrant version:"))
		fmt.Println(out)
	}
	{
		out, err := runToolVersionCommand("packer", "version")
		if err != nil {
			return fmt.Errorf("Failed to get packer version, error: %s", err)
		}
		fmt.Println()
		fmt.Println(colorstring.Green("* packer version:"))
		fmt.Println(out)
	}
	{
		out, err := runToolVersionCommand("sw_vers")
		if err != nil {
			return fmt.Errorf("Failed to get MacOS version, error: %s", err)
		}
		fmt.Println()
		fmt.Println(colorstring.Green("* Host MacOS version:"))
		fmt.Println(out)
	}

	fmt.Println()
	fmt.Println("NOTES: add your notes here")
	fmt.Println("------------------------------------")
	fmt.Println()

	return nil
}

func createVagrantBoxFromInstallMacOSApp(installMacOSAppPath string) error {
	if err := printToolVersions(); err != nil {
		return fmt.Errorf("Failed to print tool versions - missing tool - error: %s", err)
	}

	log.Printf("installMacOSAppPath: %s", installMacOSAppPath)

	macOSInstallDMGPath, err := macosinstaller.CreateInstallDMGFromInstallMacOSApp(installMacOSAppPath)
	if err != nil {
		return fmt.Errorf("Failed to create Install DMG, error: %s", err)
	}
	fmt.Println()
	log.Println(colorstring.Green("Done. Built image is located at " + macOSInstallDMGPath + "."))
	fmt.Println()

	fmt.Println()
	log.Println(colorstring.Green(" => Creating vagrant box ..."))
	fmt.Println()

	if err := vagrantbox.CreateVirtualboxVagrantBoxFromPreparedMacOSInstallDMG(macOSInstallDMGPath); err != nil {
		return fmt.Errorf("Failed to create vagrant box, error: %s", err)
	}

	return nil
}

package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/replica/vagrantbox"
	"github.com/spf13/cobra"
)

// boxCmd represents the box command
var boxCmd = &cobra.Command{
	Use:   "box INSTALL_DMG_PATH",
	Short: "Create a vagrant box, using an auto-installer dmg",
	Long: `Create a vagrant box, using an auto-installer dmg.

NOTE: You can create an auto installer DMG with: replica create dmg`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("No macOS installer DMG path provided")
		}
		installMacOSAppPath := args[0]
		return createVagrantBox(installMacOSAppPath)
	},
}

func init() {
	createCmd.AddCommand(boxCmd)
}

func createVagrantBox(macOSAutoInstallerDMGPath string) error {
	absInstallerDMGPth, err := pathutil.AbsPath(macOSAutoInstallerDMGPath)
	if err != nil {
		return fmt.Errorf("Failed to get absolute path for installer DMG (path was: %s), error: %s", macOSAutoInstallerDMGPath, err)
	}

	fmt.Println()
	log.Println(colorstring.Green(" => Creating vagrant box, using auto-installer DMG:"), absInstallerDMGPth)

	printFreeDiskSpace()

	if err := vagrantbox.CreateVirtualboxVagrantBoxFromPreparedMacOSInstallDMG(absInstallerDMGPth); err != nil {
		return fmt.Errorf("Failed to create vagrant box, error: %s", err)
	}

	printFreeDiskSpace()
	log.Println(colorstring.Green(" => vagrant box ready! [OK]"))
	return nil
}

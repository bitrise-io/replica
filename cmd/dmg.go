package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-tools/replica/macosinstaller"
	"github.com/spf13/cobra"
)

// dmgCmd represents the dmg command
var dmgCmd = &cobra.Command{
	Use:   "dmg INSTALL_MACOS_APP_PATH",
	Short: "Create the installer DMG",
	Long:  `Create the installer DMG`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("No 'Install macOS / OS X .. app' path provided")
		}
		installMacOSAppPath := args[0]
		_, err := createInstallDMG(installMacOSAppPath)
		return err
	},
}

func init() {
	createCmd.AddCommand(dmgCmd)
}

func createInstallDMG(installMacOSAppPath string) (string, error) {
	log.Printf("installMacOSAppPath: %s", installMacOSAppPath)

	printFreeDiskSpace()

	macOSInstallDMGPath, err := macosinstaller.CreateInstallDMGFromInstallMacOSApp(installMacOSAppPath)
	if err != nil {
		return "", fmt.Errorf("Failed to create Install DMG, error: %s", err)
	}

	printFreeDiskSpace()

	fmt.Println()
	log.Println(colorstring.Green("Done. Built image is located at " + macOSInstallDMGPath + "."))
	fmt.Println()

	return macOSInstallDMGPath, nil
}

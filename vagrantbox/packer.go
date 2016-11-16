package vagrantbox

import (
	"fmt"
	"log"

	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-tools/replica/resources"
)

// CreateVirtualboxVagrantBoxFromPreparedMacOSInstallDMG ...
func CreateVirtualboxVagrantBoxFromPreparedMacOSInstallDMG(macOSInstallDMGPath string) error {
	{
		if err := resources.UncompressDirectory("packer", "./_out/packer"); err != nil {
			return fmt.Errorf("Failed to uncompress packer directory, error: %s", err)
		}

		cmd := cmdex.NewCommandWithStandardOuts("packer",
			"build",
			"--only", "virtualbox-iso",
			"--var", "iso_url="+macOSInstallDMGPath,
			"--var", "autologin=true",
			"./template.json",
		).SetDir("./_out/packer")

		fmt.Println()
		log.Printf("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to run packer command, error: %s", err)
		}
	}
	return nil
}

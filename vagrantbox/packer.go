package vagrantbox

import (
	"fmt"
	"log"

	"github.com/bitrise-io/go-utils/cmdex"
)

// CreateVirtualboxVagrantBoxFromPreparedMacOSInstallDMG ...
func CreateVirtualboxVagrantBoxFromPreparedMacOSInstallDMG(macOSInstallDMGPath string) error {
	{
		cmd := cmdex.NewCommandWithStandardOuts("packer",
			"build",
			"--var", "iso_url="+macOSInstallDMGPath,
			"--var", "autologin=true",
			"--only", "virtualbox-iso",
			"./template.json",
		).SetDir("./data/packer")

		fmt.Println()
		log.Printf("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to run packer command, error: %s", err)
		}
	}
	return nil
}

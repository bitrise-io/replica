package vagrantbox

import (
	"fmt"
	"log"

	"path/filepath"

	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/replica/resources"
)

// CreateVirtualboxVagrantBoxFromPreparedMacOSInstallDMG ...
func CreateVirtualboxVagrantBoxFromPreparedMacOSInstallDMG(macOSInstallDMGPath string) (string, error) {
	outputDir, err := pathutil.AbsPath("./_out/packer")
	if err != nil {
		return "", fmt.Errorf("Failed to determin absolute output dir path, error: %s", err)
	}

	{
		if err := resources.UncompressDirectory("packer", outputDir); err != nil {
			return "", fmt.Errorf("Failed to uncompress packer directory, error: %s", err)
		}

		cmd := cmdex.NewCommandWithStandardOuts("packer",
			"build",
			"--only", "virtualbox-iso",
			"--var", "iso_url="+macOSInstallDMGPath,
			"--var", "autologin=true",
			"./template.json",
		).SetDir(outputDir)

		fmt.Println()
		log.Printf("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("Failed to run packer command, error: %s", err)
		}
	}
	return filepath.Join(outputDir, "packer_virtualbox-iso_virtualbox.box"), nil
}

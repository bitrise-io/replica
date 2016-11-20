package cmd

import (
	"errors"
	"fmt"
	"log"

	"path/filepath"

	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/spf13/cobra"
)

const (
	vagrantBoxName           = "bitrise-replica-macos"
	vagrantInitialSnapshotID = "bitrise-replica-initial"
)

// vagrantCmd represents the vagrant command
var vagrantCmd = &cobra.Command{
	Use:   "vagrant DESTINATION_DIR_PATH VAGRANT_BOX_PATH",
	Short: `Create a vagrant VM, using the vagrant box`,
	Long: `Create a vagrant VM, using the vagrant box.

NOTE: You can create the vagrant box with: replica create box`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("No vagrant box or destination directory path provided")
		}
		destinationDirPath := args[0]
		vagrantBoxPath := args[1]

		return createVagrantVM(vagrantBoxPath, destinationDirPath)
	},
}

func init() {
	createCmd.AddCommand(vagrantCmd)
}

func createVagrantVM(vagrantBoxPath, destinationDirPath string) error {
	fmt.Println()
	log.Println(colorstring.Green(" => Registering the vagrant box:"), vagrantBoxPath)
	printFreeDiskSpace()

	{
		cmd := cmdex.NewCommandWithStandardOuts("vagrant",
			"box", "add",
			"--force",
			"--name", vagrantBoxName,
			vagrantBoxPath,
		)

		fmt.Println()
		log.Printf("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to run command, error: %s", err)
		}
	}

	fmt.Println()
	printFreeDiskSpace()
	log.Println(colorstring.Green(" => vagrant box registered! [OK]"))

	fmt.Println()
	log.Println(colorstring.Green(" => Creating and booting vagrant VM at path"), destinationDirPath)

	if err := pathutil.EnsureDirExist(destinationDirPath); err != nil {
		return fmt.Errorf("Failed to create vagrant VM destination directory (path: %s), error: %s", destinationDirPath, err)
	}

	const vagrantFileContent = `# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "bitrise-replica-macos"
  config.vm.synced_folder ".", "/vagrant", :disabled => true
  config.ssh.insert_key = false
end
`

	if err := fileutil.WriteStringToFile(filepath.Join(destinationDirPath, "Vagrantfile"), vagrantFileContent); err != nil {
		return fmt.Errorf("Failed to write Vagrantfile into the destination directory, error: %s", err)
	}

	{
		cmd := cmdex.NewCommandWithStandardOuts("vagrant",
			"up",
		).SetDir(destinationDirPath)

		fmt.Println()
		log.Printf("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to run command, error: %s", err)
		}
	}

	printFreeDiskSpace()
	log.Println(colorstring.Green(" => vagrant VM created & ready! [OK]"))

	fmt.Println()
	log.Println(colorstring.Green(" => Creating an initial snapshot ..."))

	{
		cmd := cmdex.NewCommandWithStandardOuts("vagrant",
			"snapshot",
			"save", vagrantInitialSnapshotID,
		).SetDir(destinationDirPath)

		fmt.Println()
		log.Printf("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to run command, error: %s", err)
		}
	}

	printFreeDiskSpace()
	log.Println(colorstring.Green(" => Snapshot created! [OK]"))
	fmt.Println(colorstring.Yellow(" NOTE: you can restore this saved snapshot state of the virtual machine with:"))
	fmt.Println(" $ vagrant snapshot restore " + vagrantInitialSnapshotID)
	fmt.Println()

	return nil
}

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
	"github.com/bitrise-io/goinp/goinp"
	"github.com/spf13/cobra"
)

const (
	vagrantBoxName           = "bitrise-replica-macos"
	vagrantInitialSnapshotID = "bitrise-replica-initial"
)

var (
	flagIsSkipBoxReg = false
)

// vagrantCmd represents the vagrant command
var vagrantCmd = &cobra.Command{
	Use:   "vagrant DESTINATION_DIR_PATH VAGRANT_BOX_PATH",
	Short: `Create a vagrant VM, using the vagrant box`,
	Long: `Create a vagrant VM, using the vagrant box.

NOTE: You can create the vagrant box with: replica create box`,
	RunE: func(cmd *cobra.Command, args []string) error {
		destinationDirPath := ""
		vagrantBoxPath := ""

		if flagIsSkipBoxReg {
			if len(args) < 1 {
				return errors.New("no destination directory path provided")
			}
			destinationDirPath = args[0]
		} else {
			if len(args) < 2 {
				return errors.New("no vagrant box or destination directory path provided")
			}
			destinationDirPath = args[0]
			vagrantBoxPath = args[1]
		}

		return createAndProvisionVagrantVM(destinationDirPath, flagIsSkipBoxReg, vagrantBoxPath)
	},
}

func init() {
	createCmd.AddCommand(vagrantCmd)
	vagrantCmd.Flags().BoolVar(&flagIsSkipBoxReg, "skip-box-reg", false, "Skip the vagrant box registration (only use this if the box is already registered in vagrant!)")
}

func createAndProvisionVagrantVM(destinationDirPath string, isShouldSkipBoxReg bool, vagrantBoxPath string) error {
	if err := pathutil.EnsureDirExist(destinationDirPath); err != nil {
		return fmt.Errorf("Failed to create vagrant VM destination directory (path: %s), error: %s", destinationDirPath, err)
	}

	if isShouldSkipBoxReg {
		log.Println(colorstring.Yellow(" => Skipping the registration of the vagrant box"))
	} else {
		fmt.Println()
		log.Println(colorstring.Green(" => Registering the vagrant box:"), vagrantBoxPath)
		printFreeDiskSpace()

		if err := registerVagrantBox(vagrantBoxPath); err != nil {
			return fmt.Errorf("Failed to register vagrant box, error: %s", err)
		}

		fmt.Println()
		printFreeDiskSpace()
		log.Println(colorstring.Green(" => vagrant box registered! [OK]"))
	}

	fmt.Println()
	log.Println(colorstring.Green(" => Creating and booting vagrant VM at path:"), destinationDirPath)

	if err := createVagrantVM(destinationDirPath, true); err != nil {
		return fmt.Errorf("Failed to create Vagrant VM, error: %s", err)
	}

	printFreeDiskSpace()
	log.Println(colorstring.Green(" => vagrant VM created & ready! [OK]"))

	fmt.Println()
	log.Println(colorstring.Green(" => Creating an initial snapshot ..."))

	if err := createVagrantSnapshot(destinationDirPath, vagrantInitialSnapshotID); err != nil {
		return fmt.Errorf("Failed to create vagrant snapshot, error: %s", err)
	}

	printFreeDiskSpace()
	log.Println(colorstring.Green(" => Snapshot created! [OK]"))
	fmt.Println(colorstring.Yellow(" NOTE: you can restore this saved snapshot state of the virtual machine with:"))
	fmt.Println(" $ vagrant snapshot restore " + vagrantInitialSnapshotID)
	fmt.Println()

	fmt.Println()
	log.Println(colorstring.Green(" => Sync Xcode.app ..."))

	xcodeAppPath, err := goinp.AskForStringWithDefault(
		"Please specify an Xcode.app (path) to be synced into the virtual machine",
		"/Applications/Xcode.app")
	if err != nil {
		return fmt.Errorf("failed to get Xcode.app path, error: %s", err)
	}

	if err := uploadDir(destinationDirPath, xcodeAppPath, "/Applications/Xcode.app"); err != nil {
		return fmt.Errorf("failed to sync Xcode.app, error: %s", err)
	}

	printFreeDiskSpace()
	log.Println(colorstring.Green(" => Xcode.app sync DONE! [OK]"))
	fmt.Println()

	return nil
}

func uploadDir(vagrantVMDir, dirToUpload, targetPathInVM string) error {
	vagrantSSHConfigFilePth, err := saveVagrantSSHConfigIntoTmpFile(vagrantVMDir)
	if err != nil {
		return fmt.Errorf("failed to determin vagrant ssh configs, error: %s", err)
	}

	{
		cmd := cmdex.NewCommandWithStandardOuts("rsync",
			"-avhP",
			"-e", "ssh -F "+vagrantSSHConfigFilePth,
			filepath.Clean(dirToUpload)+"/", "default:"+filepath.Clean(targetPathInVM)+"/",
		).SetDir(vagrantVMDir)

		fmt.Println()
		log.Printf("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to run command, error: %s", err)
		}
	}

	return nil
}

func saveVagrantSSHConfigIntoTmpFile(vagrantVMDir string) (string, error) {
	tmpDirPth, err := pathutil.NormalizedOSTempDirPath("replica-vagrant-ssh")
	if err != nil {
		return "", fmt.Errorf("failed to create a temporary directory for vagrant ssh config file, error: %s", err)
	}
	vagrantSSHConfigFilePath := filepath.Join(tmpDirPth, "vagrant.ssh.config")

	cmd := cmdex.NewCommand("vagrant",
		"ssh-config",
	).SetDir(vagrantVMDir)

	fmt.Println()
	log.Printf("$ %s", cmd.PrintableCommandArgs())
	fmt.Println()

	sshConfigCmdOutput, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run command, error: %s", err)
	}

	if err := fileutil.WriteStringToFile(vagrantSSHConfigFilePath, sshConfigCmdOutput); err != nil {
		return "", fmt.Errorf("failed to write vagrant ssh config into file, error: %s", err)
	}

	return vagrantSSHConfigFilePath, nil
}

func createVagrantSnapshot(vagrantVMDir, snapshotID string) error {
	cmd := cmdex.NewCommandWithStandardOuts("vagrant",
		"snapshot",
		"save", snapshotID,
	).SetDir(vagrantVMDir)

	fmt.Println()
	log.Printf("$ %s", cmd.PrintableCommandArgs())
	fmt.Println()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to run command, error: %s", err)
	}

	return nil
}

func createVagrantVM(vagrantVMDirPath string, isVagrantDestroyBeforeCreate bool) error {
	if isVagrantDestroyBeforeCreate {
		{
			cmd := cmdex.NewCommandWithStandardOuts("vagrant",
				"destroy", "-f",
			).SetDir(vagrantVMDirPath)

			fmt.Println()
			log.Printf("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Failed to run command, error: %s", err)
			}
		}
	}

	const vagrantFileContent = `# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "bitrise-replica-macos"
  config.vm.synced_folder ".", "/vagrant", :disabled => true
  config.ssh.insert_key = false

  config.vm.provider "virtualbox" do |v|
    # v.linked_clone = true

    # v.memory = 4096
    # v.cpus = 2
  end
end
`

	if err := fileutil.WriteStringToFile(filepath.Join(vagrantVMDirPath, "Vagrantfile"), vagrantFileContent); err != nil {
		return fmt.Errorf("Failed to write Vagrantfile into the destination directory, error: %s", err)
	}

	{
		cmd := cmdex.NewCommandWithStandardOuts("vagrant",
			"up",
		).SetDir(vagrantVMDirPath)

		fmt.Println()
		log.Printf("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to run command, error: %s", err)
		}
	}

	return nil
}

func registerVagrantBox(vagrantBoxPath string) error {
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
	return nil
}

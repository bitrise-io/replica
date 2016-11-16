package cmd

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
)

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
		out, err := runToolVersionCommand("vagrant", "--version")
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
	{
		out, err := runToolVersionCommand("sysctl", "hw.model")
		if err != nil {
			return fmt.Errorf("Failed to get Mac hardware version, error: %s", err)
		}
		fmt.Println()
		fmt.Println(colorstring.Green("* Mac hardware version:"))
		fmt.Println(out)
	}

	fmt.Println()
	fmt.Println("NOTES: add your notes here")
	fmt.Println()
	fmt.Println("------------------------------------")
	fmt.Println()

	return nil
}

func printFreeDiskSpace() {
	var stat syscall.Statfs_t
	wd, err := os.Getwd()
	if err != nil {
		log.Println(colorstring.Red(" [!] Failed to get current working directory, error:"), err)
		return
	}
	if err := syscall.Statfs(wd, &stat); err != nil {
		log.Println(colorstring.Red(" [!] Failed to get file system stats, error:"), err)
		return
	}
	availableSpaceInBytes := stat.Bavail * uint64(stat.Bsize)
	const KB = uint64(1024)
	log.Println(" -> (i) Free disk space: ", availableSpaceInBytes/KB/KB/KB, "GB")
}

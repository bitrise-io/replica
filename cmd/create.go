package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"text/template"

	"github.com/DHowett/go-plist"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/templateutil"
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

func createVagrantBoxFromInstallMacOSApp(installMacOSAppPath string) error {
	log.Printf("installMacOSAppPath: %s", installMacOSAppPath)

	if err := createInstallDMGFromInstallMacOSApp(installMacOSAppPath); err != nil {
		return fmt.Errorf("Failed to create Install DMG, error: %s", err)
	}

	return nil
}

func createInstallDMGFromInstallMacOSApp(installMacOSAppPath string) error {
	supportDirPath := "./_support"

	accountUsername := "vagrant"
	// accountPassword := "vagrant"
	outDir := "./packer"

	installESDPath := filepath.Join(installMacOSAppPath, "Contents/SharedSupport/InstallESD.dmg")
	if isExist, err := pathutil.IsPathExists(installESDPath); err != nil {
		return fmt.Errorf("Failed to locate InstallESD.dmg, error: %s", err)
	} else if !isExist {
		return fmt.Errorf("InstallESD.dmg does not exist inside the installer at path: %s", installESDPath)
	}

	tmpDir, err := pathutil.NormalizedOSTempDirPath("replica")
	if err != nil {
		return fmt.Errorf("Failed to create temporary ESD mount directory, error: %s", err)
	}

	fmt.Println()
	log.Println(colorstring.Green(" => Attaching input OS X installer image with shadow file.."))
	tmpESDMountDir := filepath.Join(tmpDir, "mnt", "esd")
	if err := pathutil.EnsureDirExist(tmpESDMountDir); err != nil {
		return fmt.Errorf("Failed to create temporary ESD mount directory, error: %s", err)
	}
	{
		tmpESDShadowFilePath := filepath.Join(tmpDir, "esd-shadow")
		if isExist, err := pathutil.IsPathExists(tmpESDShadowFilePath); err != nil {
			return fmt.Errorf("Failed to check whether the temporary ESD shadow file already exists, error: %s", err)
		} else if isExist {
			return fmt.Errorf("Temporary ESD shadow file already exists at path: %s", tmpESDShadowFilePath)
		}

		cmd := cmdex.NewCommandWithStandardOuts("hdiutil",
			"attach", installESDPath,
			"-mountpoint", tmpESDMountDir,
			"-shadow", tmpESDShadowFilePath,
			"-nobrowse", "-owners", "on")
		fmt.Println()
		log.Printf("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to mount InstallESD into a temporary directory (path:%s), error: %s", tmpESDMountDir, err)
		}

		// cleanup
		defer func() {
			cmd := cmdex.NewCommandWithStandardOuts("hdiutil", "detach", "-quiet", "-force", tmpESDMountDir)
			fmt.Println()
			log.Printf("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()
			if err := cmd.Run(); err != nil {
				log.Printf(" [!] Failed to detach tmp ESD mount (path:%s), error: %s", tmpESDMountDir, err)
			}
		}()
	}

	fmt.Println()
	log.Println(colorstring.Green(" => Mounting BaseSystem.."))
	macOSVersion := MacOSVersionModel{}
	{
		baseSystemDMGPath := filepath.Join(tmpESDMountDir, "BaseSystem.dmg")
		if isExist, err := pathutil.IsPathExists(baseSystemDMGPath); err != nil {
			return fmt.Errorf("Failed to check whether BaseSystem.dmg exists (path:%s), error: %s", baseSystemDMGPath, err)
		} else if !isExist {
			return fmt.Errorf("BaseSystem.dmg does not exist (path:%s)", baseSystemDMGPath)
		}
		tmpBaseSystemMountDirPath := filepath.Join(tmpDir, "mnt", "basesystem")
		if err := pathutil.EnsureDirExist(tmpBaseSystemMountDirPath); err != nil {
			return fmt.Errorf("Failed to create temporary 'Base System' mount directory, error: %s", err)
		}
		cmd := cmdex.NewCommandWithStandardOuts("hdiutil",
			"attach", baseSystemDMGPath,
			"-mountpoint", tmpBaseSystemMountDirPath,
			"-nobrowse", "-owners", "on")
		fmt.Println()
		log.Printf("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to mount BaseSystem.dmg into a temporary directory (path:%s), error: %s", tmpBaseSystemMountDirPath, err)
		}

		// cleanup
		defer func() {
			cmd := cmdex.NewCommandWithStandardOuts("hdiutil", "detach", "-quiet", "-force", tmpBaseSystemMountDirPath)
			fmt.Println()
			log.Printf("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()
			if err := cmd.Run(); err != nil {
				log.Printf(" [!] Failed to detach tmp ESD mount (path:%s), error: %s", tmpBaseSystemMountDirPath, err)
			}
		}()

		systemVersionPlistFilePath := filepath.Join(tmpBaseSystemMountDirPath, "System/Library/CoreServices/SystemVersion.plist")
		macOSVer, err := readMacOSVersionFromPlist(systemVersionPlistFilePath)
		if err != nil {
			return fmt.Errorf("Failed to read MacOS version, error: %s", err)
		}
		log.Printf("macOSVer: %#v", macOSVer)
		macOSVersion = macOSVer
	}

	outDMGPath := filepath.Join(outDir, fmt.Sprintf("OSX_InstallESD_%s_%s.dmg", macOSVersion.Version, macOSVersion.Build))
	log.Printf("outDMGPath: %s", outDMGPath)

	fmt.Println()
	log.Println(colorstring.Green(" => Making firstboot installer pkg.."))
	{
		if err := pathutil.EnsureDirExist(filepath.Join(supportDirPath, "pkgroot/private/var/db/dslocal/nodes/Default/users")); err != nil {
			return fmt.Errorf("Failed to create pkg users dir, error: %s", err)
		}
		if err := pathutil.EnsureDirExist(filepath.Join(supportDirPath, "pkgroot/private/var/db/shadow/hash")); err != nil {
			return fmt.Errorf("Failed to create pkg hash dir, error: %s", err)
		}

		// Originally this was generate with: $ openssl base64 -in path/to/image.jpg
		userImagePath := filepath.Join(supportDirPath, "vagrant.jpg")
		imgContBytes, err := fileutil.ReadBytesFromFile(userImagePath)
		if err != nil {
			return fmt.Errorf("Failed to read user account image (path:%s), error: %s", userImagePath, err)
		}
		rawBase64UserImage := base64.StdEncoding.EncodeToString(imgContBytes)
		// inject newline at every 64th char, to match the output of $ openssl base64 -in path/to/image.jpg
		multilineBase64UserImage := ""
		for idx, c := range rawBase64UserImage {
			if idx%64 == 0 && idx != 0 {
				multilineBase64UserImage = multilineBase64UserImage + "\n"
			}
			multilineBase64UserImage = multilineBase64UserImage + string(c)
		}

		accountGeneratedUID := "11112222-3333-4444-AAAA-BBBBCCCCDDDD"
		userPlistContent, err := renderUserPlistTemplate(accountUsername, multilineBase64UserImage, accountGeneratedUID)
		if err != nil {
			return fmt.Errorf("Failed to render User.plist template, error: %s", err)
		}
		fmt.Println()
		fmt.Println(userPlistContent)
		fmt.Println()

		userPlistPath := filepath.Join(supportDirPath,
			"pkgroot/private/var/db/dslocal/nodes/Default/users", accountUsername+".plist")
		if err := fileutil.WriteStringToFile(userPlistPath, userPlistContent); err != nil {
			return fmt.Errorf("Failed to write User.plist into file, error: %s", err)
		}
		log.Println("User.plist (" + accountUsername + ".plist) saved into file - [OK]")
	}

	return nil
}

func renderUserPlistTemplate(accountUsername, accountImageBase64, accountGeneratedUID string) (string, error) {
	type UserPlistTemplateInventory struct {
		AccountUsername     string
		AccountImageBase64  string
		AccountGeneratedUID string
	}
	inv := UserPlistTemplateInventory{
		AccountUsername:     accountUsername,
		AccountImageBase64:  accountImageBase64,
		AccountGeneratedUID: accountImageBase64,
	}

	result, err := templateutil.EvaluateTemplateStringToString(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>authentication_authority</key>
	<array>
		<string>;ShadowHash;</string>
	</array>
	<key>generateduid</key>
	<array>
		<string>{{ .AccountGeneratedUID }}</string>
	</array>
	<key>gid</key>
	<array>
		<string>20</string>
	</array>
	<key>home</key>
	<array>
		<string>/Users/{{ .AccountUsername }}</string>
	</array>
	<key>jpegphoto</key>
	<array>
		<data>
			{{ .AccountImageBase64 }}
		</data>
	</array>
	<key>name</key>
	<array>
		<string>{{ .AccountUsername }}</string>
	</array>
	<key>passwd</key>
	<array>
		<string>********</string>
	</array>
	<key>realname</key>
	<array>
		<string>{{ .AccountUsername }}</string>
	</array>
	<key>shell</key>
	<array>
		<string>/bin/bash</string>
	</array>
	<key>uid</key>
	<array>
		<string>501</string>
	</array>
</dict>
</plist>
`,
		inv, template.FuncMap{})

	return result, err
}

// MacOSVersionModel ...
type MacOSVersionModel struct {
	Version string `plist:"ProductVersion"`
	Build   string `plist:"ProductBuildVersion"`
}

func readMacOSVersionFromPlist(plistPath string) (MacOSVersionModel, error) {
	f, err := os.Open(plistPath)
	if err != nil {
		return MacOSVersionModel{}, fmt.Errorf("Failed to open plist file (%s), error: %s", plistPath, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf(" [!] Failed to close plist file (%s), error: %s", plistPath, err)
		}
	}()
	var macOSVersion MacOSVersionModel
	if err := plist.NewDecoder(f).Decode(&macOSVersion); err != nil {
		return macOSVersion, fmt.Errorf("Failed to decode Plist file (%s) content, error: %s", plistPath, err)
	}
	return macOSVersion, nil
}

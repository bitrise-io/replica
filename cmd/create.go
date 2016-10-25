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
	// supportDirPath := "./_support"

	accountUsername := "vagrant"
	// accountPassword := "vagrant"
	outDir := "./packer"
	if err := pathutil.EnsureDirExist(outDir); err != nil {
		return fmt.Errorf("Failed to create output directory (path:%s), error: %s", outDir, err)
	}

	// ESD="$ESD/Contents/SharedSupport/InstallESD.dmg"
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

	defer func() {
		log.Println(colorstring.Yellow("If you want to clean up the temporary files created by replica,"))
		log.Println(colorstring.Yellow(" just delete the directory: "), tmpDir)
	}()

	fmt.Println()
	log.Println(colorstring.Green(" => Attaching input OS X installer image with shadow file.."))
	// MNT_ESD=$(/usr/bin/mktemp -d /tmp/veewee-osx-esd.XXXX)
	tmpESDMountDir := filepath.Join(tmpDir, "mnt", "esd")
	if err := pathutil.EnsureDirExist(tmpESDMountDir); err != nil {
		return fmt.Errorf("Failed to create temporary ESD mount directory, error: %s", err)
	}
	{
		// SHADOW_FILE=$(/usr/bin/mktemp /tmp/veewee-osx-shadow.XXXX)
		tmpESDShadowFilePath := filepath.Join(tmpDir, "esd-shadow")
		// rm "$SHADOW_FILE"
		if isExist, err := pathutil.IsPathExists(tmpESDShadowFilePath); err != nil {
			return fmt.Errorf("Failed to check whether the temporary ESD shadow file already exists, error: %s", err)
		} else if isExist {
			return fmt.Errorf("Temporary ESD shadow file already exists at path: %s", tmpESDShadowFilePath)
		}

		// hdiutil attach "$ESD" -mountpoint "$MNT_ESD" -shadow "$SHADOW_FILE" -nobrowse -owners on
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
			// hdiutil detach -quiet -force "$MNT_ESD"
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
	// msg_status "Mounting BaseSystem.."
	log.Println(colorstring.Green(" => Mounting BaseSystem.."))
	macOSVersion := MacOSVersionModel{}
	// BASE_SYSTEM_DMG="$MNT_ESD/BaseSystem.dmg"
	baseSystemDMGPath := filepath.Join(tmpESDMountDir, "BaseSystem.dmg")
	// MNT_BASE_SYSTEM=$(/usr/bin/mktemp -d /tmp/veewee-osx-basesystem.XXXX)
	tmpBaseSystemMountDirPath := filepath.Join(tmpDir, "mnt", "basesystem")
	{
		if isExist, err := pathutil.IsPathExists(baseSystemDMGPath); err != nil {
			return fmt.Errorf("Failed to check whether BaseSystem.dmg exists (path:%s), error: %s", baseSystemDMGPath, err)
		} else if !isExist {
			return fmt.Errorf("BaseSystem.dmg does not exist (path:%s)", baseSystemDMGPath)
		}

		if err := pathutil.EnsureDirExist(tmpBaseSystemMountDirPath); err != nil {
			return fmt.Errorf("Failed to create temporary 'Base System' mount directory, error: %s", err)
		}
		// hdiutil attach "$BASE_SYSTEM_DMG" -mountpoint "$MNT_BASE_SYSTEM" -nobrowse -owners on
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
			// hdiutil detach -quiet -force "$MNT_BASE_SYSTEM"
			cmd := cmdex.NewCommandWithStandardOuts("hdiutil", "detach", "-quiet", "-force", tmpBaseSystemMountDirPath)
			fmt.Println()
			log.Printf("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()
			if err := cmd.Run(); err != nil {
				log.Printf(" [!] Failed to detach tmp ESD mount (path:%s), error: %s", tmpBaseSystemMountDirPath, err)
			}
		}()

		// SYSVER_PLIST_PATH="$MNT_BASE_SYSTEM/System/Library/CoreServices/SystemVersion.plist"
		systemVersionPlistFilePath := filepath.Join(tmpBaseSystemMountDirPath, "System/Library/CoreServices/SystemVersion.plist")

		// DMG_OS_VERS=$(/usr/libexec/PlistBuddy -c 'Print :ProductVersion' "$SYSVER_PLIST_PATH")
		macOSVer, err := readMacOSVersionFromPlist(systemVersionPlistFilePath)
		if err != nil {
			return fmt.Errorf("Failed to read MacOS version, error: %s", err)
		}
		// msg_status "OS X version detected: 10.$DMG_OS_VERS_MAJOR.$DMG_OS_VERS_MINOR, build $DMG_OS_BUILD"
		log.Printf("OS X version detected: %#v", macOSVer)
		macOSVersion = macOSVer
	}

	// OUTPUT_DMG="$OUT_DIR/OSX_InstallESD_${DMG_OS_VERS}_${DMG_OS_BUILD}.dmg"
	outDMGPath := filepath.Join(outDir, fmt.Sprintf("OSX_InstallESD_%s_%s.dmg", macOSVersion.Version, macOSVersion.Build))
	log.Printf("outDMGPath: %s", outDMGPath)
	if isExist, err := pathutil.IsPathExists(outDMGPath); err != nil {
		return fmt.Errorf("Failed to check whether the output DMG file already exists, error: %s", err)
	} else if isExist {
		return fmt.Errorf("Output DMG already exists (at path: %s) - covardly refusing to overwrite it", outDMGPath)
	}

	fmt.Println()
	// msg_status "Making firstboot installer pkg.."
	log.Println(colorstring.Green(" => Making firstboot installer pkg.."))
	builtPkgPath := ""
	{
		tmpInstallerPkgPath := filepath.Join(tmpDir, "pkginst")
		if err := pathutil.EnsureDirExist(tmpInstallerPkgPath); err != nil {
			return fmt.Errorf("Failed to create tmp installer pkg dir, error: %s", err)
		}
		log.Println(" ==> Created temporary installer pkg directory at path: ", tmpInstallerPkgPath)

		pkgBuildPkgRootPath := filepath.Join(tmpInstallerPkgPath, "pkgroot")

		// mkdir -p "$SUPPORT_DIR/pkgroot/private/var/db/dslocal/nodes/Default/users"
		if err := pathutil.EnsureDirExist(filepath.Join(pkgBuildPkgRootPath, "private/var/db/dslocal/nodes/Default/users")); err != nil {
			return fmt.Errorf("Failed to create pkg users dir, error: %s", err)
		}
		// mkdir -p "$SUPPORT_DIR/pkgroot/private/var/db/shadow/hash"
		if err := pathutil.EnsureDirExist(filepath.Join(pkgBuildPkgRootPath, "private/var/db/shadow/hash")); err != nil {
			return fmt.Errorf("Failed to create pkg hash dir, error: %s", err)
		}

		// BASE64_IMAGE=$(openssl base64 -in "$IMAGE_PATH")
		userImagePath := filepath.Join("./data", "vagrant.jpg")
		imgContBytes, err := fileutil.ReadBytesFromFile(userImagePath)
		if err != nil {
			return fmt.Errorf("Failed to read user account image (path:%s), error: %s", userImagePath, err)
		}
		// Originally this was generate with: $ openssl base64 -in path/to/image.jpg
		rawBase64UserImage := base64.StdEncoding.EncodeToString(imgContBytes)
		// inject newline at every 64th char, to match the output of $ openssl base64 -in path/to/image.jpg
		multilineBase64UserImage := ""
		for idx, c := range rawBase64UserImage {
			if idx%64 == 0 && idx != 0 {
				multilineBase64UserImage = multilineBase64UserImage + "\n"
			}
			multilineBase64UserImage = multilineBase64UserImage + string(c)
		}

		// render_template "$SUPPORT_DIR/user.plist" > "$SUPPORT_DIR/pkgroot/private/var/db/dslocal/nodes/Default/users/$USER.plist"
		// USER_GUID=$(/usr/libexec/PlistBuddy -c 'Print :generateduid:0' "$SUPPORT_DIR/user.plist")
		accountGeneratedUID := "11112222-3333-4444-AAAA-BBBBCCCCDDDD"
		userPlistContent, err := renderUserPlistTemplate(accountUsername, multilineBase64UserImage, accountGeneratedUID)
		if err != nil {
			return fmt.Errorf("Failed to render User.plist template, error: %s", err)
		}
		// fmt.Println()
		// fmt.Println(userPlistContent)
		// fmt.Println()

		userPlistPath := filepath.Join(pkgBuildPkgRootPath,
			"private/var/db/dslocal/nodes/Default/users", accountUsername+".plist")
		if err := fileutil.WriteStringToFile(userPlistPath, userPlistContent); err != nil {
			return fmt.Errorf("Failed to write User.plist into file, error: %s", err)
		}
		log.Println("User.plist (" + accountUsername + ".plist) saved into file - [OK]")

		// "$SUPPORT_DIR/generate_shadowhash" "$PASSWORD" > "$SUPPORT_DIR/pkgroot/private/var/db/shadow/hash/$USER_GUID"
		// # Generate a shadowhash from the supplied password
		// user shadow hash (password)
		// generate one with _support/generate_shadowhash: _support/generate_shadowhash PASSWORD
		// this one is for "vagrant"
		accountPasswordShadowHashFilePath := filepath.Join(
			pkgBuildPkgRootPath, "private/var/db/shadow/hash", accountGeneratedUID)
		// if err := fileutil.WriteStringToFile(accountPasswordShadowHashFilePath, accountPasswordShadowHash); err != nil {
		// 	return fmt.Errorf("Failed to write account password shadow hash into file (%s), error: %s", accountPasswordShadowHashFilePath, err)
		// }
		{
			cmd := cmdex.NewCommandWithStandardOuts(
				"cp",
				"./data/usr-password-shadow",
				accountPasswordShadowHashFilePath,
			)
			fmt.Println()
			log.Printf("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Failed to run command, error: %s", err)
			}
		}

		//
		// cat "$SUPPORT_DIR/pkg-postinstall" \
		// | sed -e "s/__USER__PLACEHOLDER__/${USER}/" \
		// | sed -e "s/__DISABLE_REMOTE_MANAGEMENT__/${DISABLE_REMOTE_MANAGEMENT}/" \
		// | sed -e "s/__DISABLE_SCREEN_SHARING__/${DISABLE_SCREEN_SHARING}/" \
		// | sed -e "s/__DISABLE_SIP__/${DISABLE_SIP}/" \
		// > "$SUPPORT_DIR/tmp/Scripts/postinstall"
		//
		disableRemoteManagement := true
		disableScreenSharing := true
		disableSIP := false
		postInstScriptCont, err := renderPostInstallScriptTemplate(accountUsername, disableRemoteManagement, disableScreenSharing, disableSIP)
		if err != nil {
			return fmt.Errorf("Failed to render post install script template, error: %s", err)
		}

		postInstallScriptDirPath := filepath.Join(tmpInstallerPkgPath, "tmp/Scripts")
		// mkdir -p "$SUPPORT_DIR/tmp/Scripts"
		if err := pathutil.EnsureDirExist(postInstallScriptDirPath); err != nil {
			return fmt.Errorf("Failed to create post install Scripts directory (path:%s), error: %s", postInstallScriptDirPath, err)
		}
		postInstallScriptPath := filepath.Join(postInstallScriptDirPath, "postinstall")
		if err := fileutil.WriteStringToFile(postInstallScriptPath, postInstScriptCont); err != nil {
			return fmt.Errorf("Failed to write Post Install script into file, error: %s", err)
		}
		log.Println("Post Install script saved into file - [OK]")
		// chmod a+x "$SUPPORT_DIR/tmp/Scripts/postinstall"
		os.Chmod(postInstallScriptPath, 0755)
		log.Println("Post Install script made executable - [OK]")

		fmt.Println()
		log.Println(colorstring.Green(" ==> Building it ..."))
		// BUILT_COMPONENT_PKG="$SUPPORT_DIR/tmp/config-component.pkg"
		builtComponentPkgPath := filepath.Join(tmpInstallerPkgPath, "config-component.pkg")
		{
			// pkgbuild --quiet \
			// 	--root "$SUPPORT_DIR/pkgroot" \
			// 	--scripts "$SUPPORT_DIR/tmp/Scripts" \
			// 	--identifier com.vagrantup.config \
			// 	--version 0.1 \
			// 	"$BUILT_COMPONENT_PKG"
			cmd := cmdex.NewCommandWithStandardOuts("pkgbuild",
				"--quiet",
				"--root", pkgBuildPkgRootPath,
				"--scripts", postInstallScriptDirPath,
				"--identifier", "com.vagrantup.config",
				"--version", "0.1",
				builtComponentPkgPath,
			)
			fmt.Println()
			log.Printf("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Failed to build package, error: %s", err)
			}
		}

		fmt.Println()
		log.Println(colorstring.Green(" ==> Packaging it ..."))
		// BUILT_PKG="$SUPPORT_DIR/tmp/config.pkg"
		builtPkgPath = filepath.Join(tmpInstallerPkgPath, "config.pkg")
		{
			// productbuild \
			// 	--package "$BUILT_COMPONENT_PKG" \
			// 	"$BUILT_PKG"
			cmd := cmdex.NewCommandWithStandardOuts("productbuild",
				"--package", builtComponentPkgPath,
				builtPkgPath,
			)
			fmt.Println()
			log.Printf("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Failed to build package, error: %s", err)
			}
		}

		// # We'd previously mounted this to check versions
		// hdiutil detach "$MNT_BASE_SYSTEM"
		{
			cmd := cmdex.NewCommandWithStandardOuts("hdiutil",
				"detach", tmpBaseSystemMountDirPath,
			)
			fmt.Println()
			log.Printf("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Failed to run command, error: %s", err)
			}
		}

		// return errors.New("TEST")
	}

	fmt.Println()
	// BASE_SYSTEM_DMG_RW="$(/usr/bin/mktemp /tmp/veewee-osx-basesystem-rw.XXXX).dmg"
	baseSystemDMGRWPath := filepath.Join(tmpDir, "osx-basesystem-rw.dmg")
	// msg_status "Creating empty read-write DMG located at $BASE_SYSTEM_DMG_RW.."
	log.Println(colorstring.Green(" => Creating empty read-write DMG located at " + baseSystemDMGRWPath + ".."))
	// MNT_BASE_SYSTEM="/Volumes/OS X Base System"
	mountedBaseSystemPath := "/Volumes/OS X Base System"
	{
		// hdiutil create -o "$BASE_SYSTEM_DMG_RW" -size 10g -layout SPUD -fs HFS+J
		{
			cmd := cmdex.NewCommandWithStandardOuts("hdiutil",
				"create", "-o", baseSystemDMGRWPath,
				"-size", "10g", "-layout", "SPUD", "-fs", "HFS+J",
			)
			fmt.Println()
			log.Printf("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Failed to run command, error: %s", err)
			}
		}

		tmpBaseSystemDMGRWMountDirPath := filepath.Join(tmpDir, "mnt", "dmg-basesystem-dmg")
		if err := pathutil.EnsureDirExist(tmpBaseSystemDMGRWMountDirPath); err != nil {
			return fmt.Errorf("Failed to create temporary 'Base System' mount directory, error: %s", err)
		}

		// hdiutil attach "$BASE_SYSTEM_DMG_RW" -mountpoint "$MNT_BASE_SYSTEM" -nobrowse -owners on
		{
			cmd := cmdex.NewCommandWithStandardOuts("hdiutil",
				"attach", baseSystemDMGRWPath,
				"-mountpoint", tmpBaseSystemDMGRWMountDirPath,
				"-nobrowse", "-owners", "on",
			)
			fmt.Println()
			log.Printf("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Failed to run command, error: %s", err)
			}
		}
		// cleanup
		defer func() {
			cmd := cmdex.NewCommandWithStandardOuts("hdiutil", "detach", "-quiet", "-force", tmpBaseSystemDMGRWMountDirPath)
			fmt.Println()
			log.Printf("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()
			if err := cmd.Run(); err != nil {
				log.Printf(" [!] Failed to detach tmp ESD mount (path:%s), error: %s", tmpBaseSystemDMGRWMountDirPath, err)
			}
		}()

		log.Println(" ==> Restoring ('asr restore') the BaseSystem to the read-write DMG..")
		// This asr restore was needed as of 10.11 DP7 and up. See
		// https://github.com/timsutton/osx-vm-templates/issues/40
		//
		// Note that when the restore completes, the volume is automatically re-mounted
		// and not with the '-nobrowse' option. It's an annoyance we could possibly fix
		// in the future..

		// asr restore --source "$BASE_SYSTEM_DMG" --target "$MNT_BASE_SYSTEM" --noprompt --noverify --erase
		{
			cmd := cmdex.NewCommandWithStandardOuts("asr",
				"restore", "--source", baseSystemDMGPath,
				"--target", tmpBaseSystemDMGRWMountDirPath,
				"--noprompt", "--noverify", "--erase",
			)
			fmt.Println()
			log.Printf("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Failed to run command, error: %s", err)
			}
		}

		// rm -r "$MNT_BASE_SYSTEM"
		{
			cmd := cmdex.NewCommandWithStandardOuts("rm",
				"-r", tmpBaseSystemDMGRWMountDirPath,
			)
			fmt.Println()
			log.Printf("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Failed to run command, error: %s", err)
			}
		}

		{
			// BASESYSTEM_OUTPUT_IMAGE="$OUTPUT_DMG"
			// outDMGPath

			// PACKAGES_DIR="$MNT_BASE_SYSTEM/System/Installation/Packages"
			packagesDir := filepath.Join(mountedBaseSystemPath, "System/Installation/Packages")

			// rm "$PACKAGES_DIR"
			if err := os.Remove(packagesDir); err != nil {
				return fmt.Errorf("Failed to remove mounted Packages dir (path:%s), error: %s", packagesDir, err)
			}

			// msg_status "Moving 'Packages' directory from the ESD to BaseSystem.."
			log.Println(" ==> Moving 'Packages' directory from the ESD to BaseSystem..")

			// sudo mv -v "$MNT_ESD/Packages" "$MNT_BASE_SYSTEM/System/Installation/"
			{
				cmd := cmdex.NewCommandWithStandardOuts("sudo",
					"mv", "-v",
					filepath.Join(tmpESDMountDir, "Packages"),
					filepath.Join(mountedBaseSystemPath, "System/Installation")+"/",
				)
				fmt.Println()
				log.Printf("$ %s", cmd.PrintableCommandArgs())
				fmt.Println()
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("Failed to run command, error: %s", err)
				}
			}

			// # This isn't strictly required for Mavericks, but Yosemite will consider the
			// # installer corrupt if this isn't included, because it cannot verify BaseSystem's
			// # consistency and perform a recovery partition verification
			// msg_status "Copying in original BaseSystem dmg and chunklist.."
			log.Println(" ==> Copying in original BaseSystem dmg and chunklist..")

			// cp "$MNT_ESD/BaseSystem.dmg" "$MNT_BASE_SYSTEM/"
			{
				cmd := cmdex.NewCommandWithStandardOuts("cp",
					filepath.Join(tmpESDMountDir, "BaseSystem.dmg"),
					mountedBaseSystemPath+"/",
				)
				fmt.Println()
				log.Printf("$ %s", cmd.PrintableCommandArgs())
				fmt.Println()
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("Failed to run command, error: %s", err)
				}
			}
			// cp "$MNT_ESD/BaseSystem.chunklist" "$MNT_BASE_SYSTEM/"
			{
				cmd := cmdex.NewCommandWithStandardOuts("cp",
					filepath.Join(tmpESDMountDir, "BaseSystem.chunklist"),
					mountedBaseSystemPath+"/",
				)
				fmt.Println()
				log.Printf("$ %s", cmd.PrintableCommandArgs())
				fmt.Println()
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("Failed to run command, error: %s", err)
				}
			}

			// msg_status "Adding automated components.."
			fmt.Println()
			log.Println(colorstring.Green(" => Adding automated components.."))
			{
				// CDROM_LOCAL="$MNT_BASE_SYSTEM/private/etc/rc.cdrom.local"
				cdromDotLocalFilePath := filepath.Join(mountedBaseSystemPath, "private/etc/rc.cdrom.local")
				// cat > $CDROM_LOCAL << EOF
				// diskutil eraseDisk jhfs+ "Macintosh HD" GPTFormat disk0
				// if [ "\$?" == "1" ]; then
				//     diskutil eraseDisk jhfs+ "Macintosh HD" GPTFormat disk1
				// fi
				// EOF
				cdromFileCont := `diskutil eraseDisk jhfs+ "Macintosh HD" GPTFormat disk0
				if [ "\$?" == "1" ]; then
				    diskutil eraseDisk jhfs+ "Macintosh HD" GPTFormat disk1
				fi`
				if err := fileutil.WriteStringToFile(cdromDotLocalFilePath, cdromFileCont); err != nil {
					return fmt.Errorf("Failed to write rc.cdrom.local content into file, error: %s", err)
				}
				// chmod a+x "$CDROM_LOCAL"
				os.Chmod(cdromDotLocalFilePath, 0755)

				{
					// mkdir "$PACKAGES_DIR/Extras"
					packagesExtrasDirPath := filepath.Join(packagesDir, "Extras")
					if err := pathutil.EnsureDirExist(packagesExtrasDirPath); err != nil {
						return fmt.Errorf("Failed to create Packages/Extras, error: %s", err)
					}

					// cp "$SUPPORT_DIR/minstallconfig.xml" "$PACKAGES_DIR/Extras/"
					minstallconfigXMLContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>InstallType</key>
	<string>automated</string>
	<key>Language</key>
	<string>en</string>
	<key>Package</key>
	<string>/System/Installation/Packages/OSInstall.collection</string>
	<key>Target</key>
	<string>/Volumes/Macintosh HD</string>
	<key>TargetName</key>
	<string>Macintosh HD</string>
</dict>
</plist>
`
					fpth := filepath.Join(packagesExtrasDirPath, "minstallconfig.xml")
					if err := fileutil.WriteStringToFile(fpth, minstallconfigXMLContent); err != nil {
						return fmt.Errorf("Failed to write 'minstallconfig.xml' into file, error: %s", err)
					}
				}
				// cp "$SUPPORT_DIR/OSInstall.collection" "$PACKAGES_DIR/"
				{
					osInstallCollectioncont := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<array>
	<string>/System/Installation/Packages/OSInstall.mpkg</string>
	<string>/System/Installation/Packages/OSInstall.mpkg</string>
    <string>/System/Installation/Packages/config.pkg</string>
</array>
</plist>
`
					fpth := filepath.Join(packagesDir, "OSInstall.collection")
					if err := fileutil.WriteStringToFile(fpth, osInstallCollectioncont); err != nil {
						return fmt.Errorf("Failed to write 'OSInstall.collection' into file, error: %s", err)
					}
				}

				// cp "$BUILT_PKG" "$PACKAGES_DIR/"
				{
					cmd := cmdex.NewCommandWithStandardOuts("cp",
						builtPkgPath,
						packagesDir+"/",
					)
					fmt.Println()
					log.Printf("$ %s", cmd.PrintableCommandArgs())
					fmt.Println()
					if err := cmd.Run(); err != nil {
						return fmt.Errorf("Failed to run command, error: %s", err)
					}
				}
				// rm -rf "$SUPPORT_DIR/tmp"
			}
		}
	}

	// msg_status "Unmounting BaseSystem.."
	log.Println(colorstring.Green(" => Unmounting BaseSystem.."))
	// hdiutil detach "$MNT_BASE_SYSTEM"
	{
		cmd := cmdex.NewCommandWithStandardOuts("hdiutil",
			"detach", mountedBaseSystemPath,
		)
		fmt.Println()
		log.Printf("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to run command, error: %s", err)
		}
	}

	// msg_status "On Mavericks and later, the entire modified BaseSystem is our output dmg."
	// hdiutil convert -format UDZO -o "$OUTPUT_DMG" "$BASE_SYSTEM_DMG_RW"
	{
		cmd := cmdex.NewCommandWithStandardOuts("hdiutil",
			"convert", "-format", "UDZO",
			"-o", outDMGPath,
			baseSystemDMGRWPath,
		)
		fmt.Println()
		log.Printf("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to run command, error: %s", err)
		}
	}

	// msg_status "Checksumming output image.."
	// MD5=$(md5 -q "$OUTPUT_DMG")
	// msg_status "MD5: $MD5"

	// msg_status "Done. Built image is located at $OUTPUT_DMG. Add this iso and its checksum to your template."
	log.Println(colorstring.Green("Done. Built image is located at " + outDMGPath + ". Add this iso and its checksum to your template."))

	return nil
}

func renderPostInstallScriptTemplate(accountUsername string, disableRemoteManagement, disableScreenSharing, disableSIP bool) (string, error) {
	type TemplateInventory struct {
		AccountUsername         string
		DisableRemoteManagement int
		DisableScreenSharing    int
		DisableSIP              int
	}
	inv := TemplateInventory{
		AccountUsername:         accountUsername,
		DisableRemoteManagement: 0,
		DisableScreenSharing:    0,
		DisableSIP:              0,
	}
	if disableRemoteManagement {
		inv.DisableRemoteManagement = 1
	}
	if disableScreenSharing {
		inv.DisableScreenSharing = 1
	}
	if disableSIP {
		inv.DisableSIP = 1
	}

	result, err := templateutil.EvaluateTemplateStringToString(`#!/bin/sh
USER="{{ .AccountUsername }}"
OSX_VERS=$(sw_vers -productVersion | awk -F "." '{print $2}')
PlistBuddy="/usr/libexec/PlistBuddy"

target_ds_node="${3}/private/var/db/dslocal/nodes/Default"
# Override the default behavior of sshd on the target volume to be not disabled
if [ "$OSX_VERS" -ge 10 ]; then
    OVERRIDES_PLIST="$3/private/var/db/com.apple.xpc.launchd/disabled.plist"
    $PlistBuddy -c 'Delete :com.openssh.sshd' "$OVERRIDES_PLIST"
    $PlistBuddy -c 'Add :com.openssh.sshd bool False' "$OVERRIDES_PLIST"
    if [ {{ .DisableScreenSharing }} = 0 ]; then
        $PlistBuddy -c 'Delete :com.apple.screensharing' "$OVERRIDES_PLIST"
        $PlistBuddy -c 'Add :com.apple.screensharing bool False' "$OVERRIDES_PLIST"
    fi
else
    OVERRIDES_PLIST="$3/private/var/db/launchd.db/com.apple.launchd/overrides.plist"
    $PlistBuddy -c 'Delete :com.openssh.sshd' "$OVERRIDES_PLIST"
    $PlistBuddy -c 'Add :com.openssh.sshd:Disabled bool False' "$OVERRIDES_PLIST"
    if [ {{ .DisableScreenSharing }} = 0 ]; then
        $PlistBuddy -c 'Delete :com.apple.screensharing' "$OVERRIDES_PLIST"
        $PlistBuddy -c 'Add :com.apple.screensharing:Disabled bool False' "$OVERRIDES_PLIST"
    fi
fi

# Add user to sudoers
cp "$3/etc/sudoers" "$3/etc/sudoers.orig"
echo "$USER ALL=(ALL) NOPASSWD: ALL" >> "$3/etc/sudoers"

# Add user to admin group memberships (even though GID 80 is enough for most things)
USER_GUID=$($PlistBuddy -c 'Print :generateduid:0' "$target_ds_node/users/$USER.plist")
USER_UID=$($PlistBuddy -c 'Print :uid:0' "$target_ds_node/users/$USER.plist")
$PlistBuddy -c 'Add :groupmembers: string '"$USER_GUID" "$target_ds_node/groups/admin.plist"

# Add user to SSH SACL group membership
ssh_group="${target_ds_node}/groups/com.apple.access_ssh.plist"
$PlistBuddy -c 'Add :groupmembers array' "${ssh_group}"
$PlistBuddy -c 'Add :groupmembers:0 string '"$USER_GUID"'' "${ssh_group}"
$PlistBuddy -c 'Add :users array' "${ssh_group}"
$PlistBuddy -c 'Add :users:0 string '$USER'' "${ssh_group}"

# Enable Remote Desktop and configure user with full privileges
if [ {{ .DisableRemoteManagement }} = 0 ]; then
    echo "enabled" > "$3/private/etc/RemoteManagement.launchd"
    $PlistBuddy -c 'Add :naprivs array' "$target_ds_node/users/$USER.plist"
    $PlistBuddy -c 'Add :naprivs:0 string -1073741569' "$target_ds_node/users/$USER.plist"
fi

if [ {{ .DisableSIP }} = 1 ]; then
    csrutil disable
fi

# Pre-create user folder so veewee will have somewhere to scp configinfo to
mkdir -p "$3/Users/$USER/Library/Preferences"

# Suppress annoying iCloud welcome on a GUI login
$PlistBuddy -c 'Add :DidSeeCloudSetup bool true' "$3/Users/$USER/Library/Preferences/com.apple.SetupAssistant.plist"
$PlistBuddy -c 'Add :LastSeenCloudProductVersion string 10.'"$OSX_VERS" "$3/Users/$USER/Library/Preferences/com.apple.SetupAssistant.plist"

# Fix ownership now that the above has made a Library folder as root
chown -R "$USER_UID":20 "$3/Users/$USER"

# Disable Diagnostics submissions prompt if 10.10
# http://macops.ca/diagnostics-prompt-yosemite
if [ "$OSX_VERS" -ge 10 ]; then
    # Apple's defaults
    SUBMIT_TO_APPLE=YES
    SUBMIT_TO_APP_DEVELOPERS=NO

    CRASHREPORTER_SUPPORT="$3/Library/Application Support/CrashReporter"
    CRASHREPORTER_DIAG_PLIST="${CRASHREPORTER_SUPPORT}/DiagnosticMessagesHistory.plist"
    if [ ! -d "${CRASHREPORTER_SUPPORT}" ]; then
        mkdir "${CRASHREPORTER_SUPPORT}"
        chmod 775 "${CRASHREPORTER_SUPPORT}"
        chown root:admin "${CRASHREPORTER_SUPPORT}"
    fi
    for key in AutoSubmit AutoSubmitVersion ThirdPartyDataSubmit ThirdPartyDataSubmitVersion; do
        $PlistBuddy -c "Delete :$key" "${CRASHREPORTER_DIAG_PLIST}" 2> /dev/null
    done
    $PlistBuddy -c "Add :AutoSubmit bool ${SUBMIT_TO_APPLE}" "${CRASHREPORTER_DIAG_PLIST}"
    $PlistBuddy -c "Add :AutoSubmitVersion integer 4" "${CRASHREPORTER_DIAG_PLIST}"
    $PlistBuddy -c "Add :ThirdPartyDataSubmit bool ${SUBMIT_TO_APP_DEVELOPERS}" "${CRASHREPORTER_DIAG_PLIST}"
    $PlistBuddy -c "Add :ThirdPartyDataSubmitVersion integer 4" "${CRASHREPORTER_DIAG_PLIST}"
fi

# Disable loginwindow screensaver to save CPU cycles
$PlistBuddy -c 'Add :loginWindowIdleTime integer 0' "$3/Library/Preferences/com.apple.screensaver.plist"

# Disable the welcome screen
touch "$3/private/var/db/.AppleSetupDone"
`, inv, template.FuncMap{})

	return result, err
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
		AccountGeneratedUID: accountGeneratedUID,
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
`, inv, template.FuncMap{})

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

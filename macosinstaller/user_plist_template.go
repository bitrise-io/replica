package macosinstaller

import (
	"text/template"

	"github.com/bitrise-io/go-utils/templateutil"
)

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

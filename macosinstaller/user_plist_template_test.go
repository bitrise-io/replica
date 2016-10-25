package macosinstaller

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_renderUserPlistTemplate(t *testing.T) {
	result, err := renderUserPlistTemplate("ACCUSRNAME", "ACCIMGB64", "ACCGENUID")
	require.NoError(t, err)
	require.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>authentication_authority</key>
	<array>
		<string>;ShadowHash;</string>
	</array>
	<key>generateduid</key>
	<array>
		<string>ACCGENUID</string>
	</array>
	<key>gid</key>
	<array>
		<string>20</string>
	</array>
	<key>home</key>
	<array>
		<string>/Users/ACCUSRNAME</string>
	</array>
	<key>jpegphoto</key>
	<array>
		<data>
			ACCIMGB64
		</data>
	</array>
	<key>name</key>
	<array>
		<string>ACCUSRNAME</string>
	</array>
	<key>passwd</key>
	<array>
		<string>********</string>
	</array>
	<key>realname</key>
	<array>
		<string>ACCUSRNAME</string>
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
`, result)
}

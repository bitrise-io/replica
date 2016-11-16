#!/bin/bash
set -ex

echo "Enabling automatic GUI login for the '$USERNAME' user.."

python /private/tmp/set_kcpassword.py "$PASSWORD"

/usr/bin/defaults write /Library/Preferences/com.apple.loginwindow autoLoginUser "$USERNAME"

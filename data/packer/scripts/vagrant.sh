#!/bin/sh
set -ex
OSX_VERS=$(sw_vers -productVersion | awk -F "." '{print $2}')

# Set computer/hostname
COMPNAME=osx-10_${OSX_VERS}
scutil --set ComputerName ${COMPNAME}
scutil --set HostName ${COMPNAME}.vagrantup.com

echo "Installing vagrant keys for $USERNAME user"
mkdir "/Users/$USERNAME/.ssh"
chmod 700 "/Users/$USERNAME/.ssh"
curl -L 'https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant.pub' > "/Users/$USERNAME/.ssh/authorized_keys"
chmod 600 "/Users/$USERNAME/.ssh/authorized_keys"
chown -R "$USERNAME" "/Users/$USERNAME/.ssh"

# Create a group and assign the user to it
dseditgroup -o create "$USERNAME"
dseditgroup -o edit -a "$USERNAME" "$USERNAME"

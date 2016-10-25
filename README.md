# replica

Create a CI environment (with Docker / virtual machine)

_Based on the awesome [https://github.com/timsutton/osx-vm-templates](https://github.com/timsutton/osx-vm-templates) project,
and [https://spin.atomicobject.com/2015/11/17/vagrant-osx/](https://spin.atomicobject.com/2015/11/17/vagrant-osx/) blog post._


## Links

* [Developing on OS X Inside Vagrant - automated MacOS vagrant box creation](https://spin.atomicobject.com/2015/11/17/vagrant-osx/)
    * Using: https://github.com/timsutton/osx-vm-templates
* [New Adventures in Automating OS X Installs with startosinstall](https://macops.ca/new-adventures-in-automating-os-x-installs-with-startosinstall)

## InstallDMG location:

```
"Install OS X..." / "Install macOS ..." app -> Show package content -> Contents -> SharedSupport -> InstallESD.dmg
```


## TODO

- auto add the created box into vagrant
- tool versions: auto save into file if create is successful
- save vagrant box into _out, and maybe expose commands to only do parts (prep | packer)
- delete tmp dir, unless error or flag passed
- password: try to write it into file from string

- elimintate `cd`s - generate the files right where it have to be
- annotate the code, based on the original
- remove `veewee` from everywhere


## Tested tool versions

You can find the tool versions we tested with in the `TESTED_TOOL_VERSIONS.md`
file. __Feel free to add your tool versions report to the list!__

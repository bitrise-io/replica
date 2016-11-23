# Tested tool versions

```
* VirtualBox version:
5.1.8r111374

* vagrant version:
Vagrant 1.8.6

* packer version:
Packer v0.12.0

* Host MacOS version:
ProductName:	Mac OS X
ProductVersion:	10.12.1
BuildVersion:	16B2555

* Mac hardware version:
hw.model: MacBookPro11,1

NOTES: works perfectly - vagrant 1.8.7 has issues with local boxes, don't use that
or search for a fix, but 1.8.6 works.
That said, we had kernel panic with this config once, while we were working
with the virtual machine (not during the creation, but after that, when the
vagrant VM was created, booted and we tested a couple of things in it).
```

```
* VirtualBox version:
5.1.8r111374

* vagrant version:
Vagrant 1.8.6

* packer version:
Packer v0.11.0

* Host MacOS version:
ProductName:	Mac OS X
ProductVersion:	10.12.1
BuildVersion:	16B2555

* Mac hardware version:
hw.model: MacBookPro11,1

NOTES: so far worked perfectly
```

```
* VirtualBox version:
5.1.8r111374

* vagrant version:
Vagrant 1.8.6

* packer version:
Packer v0.10.2

* Host MacOS version:
ProductName:	Mac OS X
ProductVersion:	10.12.1
BuildVersion:	16B2555

* Mac hardware version:
hw.model: MacBookPro11,1

NOTES: so far worked perfectly
```

```
* VirtualBox version:
5.1.6r110634

* vagrant version:
Vagrant 1.8.6

* packer version:
Packer v0.10.2

* Host MacOS version:
ProductName:	Mac OS X
ProductVersion:	10.12
BuildVersion:	16A323

* Mac hardware version:
hw.model: MacBookPro11,1

NOTES: we saw occasional kernel panics / reboots (on the host Mac!)
with this combination, during the box preparation process
```
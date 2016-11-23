## Changelog (Current version: 0.9.5)

-----------------

## 0.9.5 (2016 Nov 23)

### Release Notes

* `replica create vagrant` now auto creates a vagrant/VirtualBox snapshot after an initial, successful boot,
  so that you can revert to this "vanilla" state any time you want to
* macOS Sierra: suppressing the initial Siri prompt (first login)
* `replica create` now includes `replica create vagrant`, if you allow it when prompted (after the box is
  successfully created)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-tools/replica/releases/download/0.9.5/replica-$(uname -s)-$(uname -m) > /usr/local/bin/replica
```

Then:

```
chmod +x /usr/local/bin/replica
```

That's all, you're ready to go!

### Release Commits - 0.9.4 -> 0.9.5

* [46b84c1] Viktor Benei - auto create vagrant if user allows it, in full `replica create` (#6) (2016 Nov 23)
* [073e843] Viktor Benei - version number 0.9.5 (2016 Nov 22)
* [8607c3e] Viktor Benei - Feature/siri prompt suppress (#5) (2016 Nov 21)
* [ccb49fd] Viktor Benei - Feature/vagrant rev (#4) (2016 Nov 21)
* [f56df31] Viktor Benei - replica create vagrant - auto create a snapshot (#3) (2016 Nov 20)
* [b248a73] Viktor Benei - tested tool versions addition (2016 Nov 20)
* [a349eaf] Viktor Benei - install section (2016 Nov 19)


## 0.9.4 (2016 Nov 16)

### Release Notes

* New command: `replica create vagrant`, to create and boot
  a `vagrant` VM.

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-tools/replica/releases/download/0.9.4/replica-$(uname -s)-$(uname -m) > /usr/local/bin/replica
```

Then:

```
chmod +x /usr/local/bin/replica
```

That's all, you're ready to go!

### Release Commits - 0.9.3 -> 0.9.4

* [15c919d] Viktor Benei - v0.9.4 - version (2016 Nov 16)
* [f7a8443] Viktor Benei - base `replica create vagrant` command (#2) (2016 Nov 16)


## 0.9.3 (2016 Nov 16)

### Release Notes

* Completely stand alone binary, with embedded resources
* new `create` sub commands:
    * `replica create dmg` : creates the auto-install DMG, then stops
    * `replica create box` : creates the `vagrant` box from an input auto-install DMG
    * `replica create` performs both, but you can now perform these separately
      with the new sub commands

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-tools/replica/releases/download/0.9.3/replica-$(uname -s)-$(uname -m) > /usr/local/bin/replica
```

Then:

```
chmod +x /usr/local/bin/replica
```

That's all, you're ready to go!

### Release Commits - 0.9.2 -> 0.9.3

* [12adced] Viktor Benei - v0.9.3 - version (2016 Nov 16)
* [07b1793] Viktor Benei - Feature/completely stand alone base release (#1) (2016 Nov 16)


## 0.9.2 (2016 Nov 15)

### Release Notes

* One more release fix

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-tools/replica/releases/download/0.9.2/replica-$(uname -s)-$(uname -m) > /usr/local/bin/replica
```

Then:

```
chmod +x /usr/local/bin/replica
```

That's all, you're ready to go!

### Release Commits - 0.9.1 -> 0.9.2

* [a8e331d] Viktor Benei - v0.9.2 - version (2016 Nov 15)
* [9ccbf8f] Viktor Benei - one more release fix (2016 Nov 15)


## 0.9.1 (2016 Nov 15)

### Release Notes

* Fixing the release process / workflow

### Release Commits - v0.9.0 -> 0.9.1

* [c5a6bed] Viktor Benei - v0.9.1 - prep (2016 Nov 15)
* [f63c111] Viktor Benei - create release fix (2016 Nov 15)
* [252065b] Viktor Benei - prepare, create and publish release (2016 Nov 15)


-----------------

Updated: 2016 Nov 23
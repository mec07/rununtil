# rununtil
Go library to run a function until a kill signal is recieved.

See the docs: https://godoc.org/github.com/mec07/rununtil

## Changelog

The [CHANGELOG.md](./CHANGELOG.md) file tracks all of the changes and each release.
We are managing it using the helpful [changelog-tool](https://github.com/ponylang/changelog-tool).
On Mac you can install it with brew: `brew install kaluza-tech/devint/changelog-tool`.
On linux (or any other platform) you have to install the pony language.
Then just follow the instructions on the github page for installing the changelog-tool:
https://github.com/ponylang/changelog-tool#installation

The use of the tool is straightforward.
To create a new changelog (don't run this in this repo because then you'll replace the current changelog with a new one!):
```
changelog-tool new
```
To start recording a new entry:
```
changelog-tool unreleased -e
```
The `-e` means update the changelog file in place.
Then manually edit the changelog to add your changes in the unreleased section.
When you're ready you can then "release" it by executing:
```
changelog-tool release 0.0.1 -e
```
Replace `0.0.1` with the new version.

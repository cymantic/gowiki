# gowiki

A wiki you can run locally, backed up with a git repo for versioning.

## TL;DR

```
./gowiki -data=/var/gowiki/data -init=https://github.com/cymantic/gowiki-data.git
```

## Installation

### Requirements
For Git Integration this uses the [git2go][git2go] library which will require [libgit2][libgit2] installing.

### Configuration
To push to a remote repo as origin with a (github deploy) key, you must configure an ssh key and set the following environment variables:
  * `GOWIKI_GIT_USERNAME` (most likely to be `git`)
  * `GOWIKI_GIT_SSH_KEY_PATH` (file path to the ssh key, assumes public key is there with `.pub`)
  * `GOWIKI_GIT_PASSPHRASE` 

### Initial Run
  * To set the directory for the wiki data pass `-data=<path to data>`
  * To create the initial wiki data pass `-init=https://github.com/cymantic/gowiki-data.git`
  * To setup an origin to push changes to pass `-origin=git@github.com:<username>/mywiki-data.git`
  
## Using without Git
Copy the files from `https://github.com/cymantic/gowiki-data.git` to the data directory.

Start with `./gowiki -data=/var/gowiki/data`


[git2go]: https://github.com/libgit2/git2go
[libgit2]: https://libgit2.github.com/

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
Allows you to start a new empty wiki.

  * Set the directory for the wiki data: `-data=<path to data>`
  * Create the initial wiki from gowiki-data repo: `-init=https://github.com/cymantic/gowiki-data.git`
  * Setup an origin to push changes to: `-origin=git@github.com:<username>/mywiki-data.git`
  
### Clone Run
Allows you to start a wiki from an existing gowiki git repo.

  * Set the directory for the wiki data: `-data=<path to data>`
  * Clone from an existing gowiki data repo: `-clone=git@github.com:<username>/mywiki-data.git`
  
  
### Update Run
After init or clone, when just starting `gowiki`, if there is an origin, `gowiki` will pull the latest changes, but cannot resolve merge conflicts
  
## Using without Git
Copy the files from `https://github.com/cymantic/gowiki-data.git` to the data directory.

Start with `./gowiki -data=/var/gowiki/data`


[git2go]: https://github.com/libgit2/git2go
[libgit2]: https://libgit2.github.com/

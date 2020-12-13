[![Build](https://github.com/driessamyn/deduplicater/workflows/Build/badge.svg)](https://github.com/driessamyn/deduplicater/actions?query=workflow%3ABuild)
[![codecov](https://codecov.io/gh/driessamyn/deduplicater/branch/main/graph/badge.svg?token=1iZhvHQt29)](https://codecov.io/gh/driessamyn/deduplicater)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# de-duplicater

Find and manage duplicate files
Utility to help identify and manage duplicate files on the file system.
This uses a 2 step process:

1. Create an index: this is a quite slow (depending on the number and size of files) process where it creates an index of all files in the given folder. During the creation of thi index, the script will calculate one or more hashes that can be used to identify duplicates (see next step)
1. Once the index is created, the script will help identify duplicate files, following one or more strategies and based on information captured in the previous step.

## Install

Download the [release binaries for MacOS, Windows or Linux](https://github.com/driessamyn/deduplicater/releases).

Or use [homebrew](https://brew.sh/):

`brew install driessamyn/homewbrew-repo/deduplicater`.

## Usage

### Create index

Create and index of all files in `/mnt/c/Users/bob/Pictures` and store the index in `/mnt/c/Users/bob/Pictures`.
A file called `.duplicate-index.json` will be placed in `/mnt/c/Users/bob/Pictures`.
Currently `md5` is the only hashing option.
```bash
deduplicater index --md5 -d "/mnt/c/Users/bob/Pictures" -f "/mnt/c/Users/bob/Pictures"
```

### Find and remove duplicates

Use the index to identify duplicate files.
Currently only the `md5` hash is supported to identify duplicates.
``` bash
deduplicater find --md5 -f "/mnt/c/Users/bob/Pictures"
```

If duplicates are found, they can optionally be removed.

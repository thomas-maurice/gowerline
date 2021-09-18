# from https://gist.github.com/jpmens/6248478

# -*- coding: utf-8 -*-
# Author: Douglas Creager <dcreager@dcreager.net>
# This file is placed into the public domain.

# Calculates the current version number.  If possible, this is the
# output of "git describe", modified to conform to the versioning
# scheme that setuptools uses.  If “git describe” returns an error
# (most likely because we're in an unpacked copy of a release tarball,
# rather than in a git working copy), then we fall back on reading the
# contents of the RELEASE-VERSION file.
#
# To use this script, simply import it your setup.py file, and use the
# results of get_git_version() as your package version:
#
# from version import *
#
# setup(
#     version=get_git_version(),
#     .
#     .
#     .
# )
#
#
# This will automatically update the RELEASE-VERSION file, if
# necessary.  Note that the RELEASE-VERSION file should *not* be
# checked into git; please add it to your top-level .gitignore file.
#
# You'll probably want to distribute the RELEASE-VERSION file in your
# sdist tarballs; to do this, just create a MANIFEST.in file that
# contains the following line:
#
#   include RELEASE-VERSION

__all__ = ("get_git_version")

from subprocess import Popen, PIPE
import os.path


def call_git_describe(abbrev):
    try:
        if not os.path.isdir(".git"):
            raise Exception("not in a git repo")

        p = Popen(['bash', '-c', "make version | grep Version | awk '{ print $3 }' | sed -e 's/,$//'"],
                  stdout=PIPE, stderr=PIPE)
        p.stderr.close()
        line = p.stdout.readlines()[0]
        return bytes(line.strip())

    except:
        versionFile = open("VERSION", "r")
        return versionFile.read().strip()


def write_release_version(version):
    f = open("RELEASE-VERSION", "w")
    f.write("%s\n" % version)
    f.close()


def get_git_version(abbrev=7):
    version = call_git_describe(abbrev)

    if version is None:
        raise ValueError("Cannot find the version number!")

    if type(version) == bytes:
        return version.decode()
    return version


if __name__ == "__main__":
    print(get_git_version())

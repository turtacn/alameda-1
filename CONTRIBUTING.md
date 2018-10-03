# Contributing to CoreDNS
Welcome! The Alameda project is [Apache 2.0 licensed](LICENSE) and accepts contributions via GitHub
pull requests. This document outlines some of the conventions on development
workflow, commit message formatting, contact points and other resources to make
it easier to get your contribution accepted.

## Developer Certificate Of Origin
By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [Developer Certificate of Origin (DCO)](https://developercertificate.org/) file for details.

Contributors sign-off that they adhere to these requirements by adding a
Signed-off-by line to commit messages. For example:

    This is my commit message

    Signed-off-by: Random J Developer <random@developer.example.org>

Git even has a -s command line option to append this automatically to your
commit message:

    git commit -s -m 'This is my commit message'

If you have already made a commit and forgot to include the sign-off, you can amend your last commit
to add the sign-off with the following command, which can then be force pushed.


    git commit --amend -s


We use a [DCO bot](https://github.com/apps/dco) to enforce the DCO on each pull
request and branch commits.

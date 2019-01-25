# Contributing to Alameda 
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

## Getting Started

If you have questions, comments, or requests feel free to post on the mailing list or
create an issue on GitHub.

If you want to contribute code and you are new to the Go programming language, check out
the [DEVELOPMENT.md](./docs/devel/DEVELOPMENT.md) reference for help getting started.

We currently welcome contributions of all kinds. For example:

- Development of features, bug fixes, and other improvements.
- Documentation including reference material and examples.
- Bug and feature reports.

## Issues

Issues are used as the primary method for tracking anything to do with the Alameda project.

### Issue Types

There are 4 types of issues (each with their own corresponding [label](#labels)):
- Question: These are support or functionality inquiries that we want to have a record of for
future reference. Generally these are questions that are too complex or large to store in the
Slack channel or have particular interest to the community as a whole. Depending on the discussion,
these can turn into "Feature" or "Bug" issues.
- Proposal: Used for items (like this one) that propose a new ideas or functionality that require
a larger community discussion. This allows for feedback from others in the community before a
feature is actually  developed. This is not needed for small additions. Final word on whether or
not a feature needs a proposal is up to the core maintainers. All issues that are proposals should
both have a label and an issue title of "Proposal: [the rest of the title]." A proposal can become
a "Feature" and does not require a milestone.
- Features: These track specific feature requests and ideas until they are complete. They can evolve
from a "Proposal" or can be submitted individually depending on the size.
- Bugs: These track bugs with the code or problems with the documentation (i.e. missing or incomplete)

### Issue Lifecycle

The issue lifecycle is mainly driven by the core maintainers, but is good information for those
contributing to Alameda. All issue types follow the same general lifecycle. Differences are noted below.
1. Issue creation
2. Triage
    - The maintainer in charge of triaging will apply the proper labels for the issue. This
    includes labels for priority, type, and metadata (such as "starter"). The only issue
    priority we will be tracking is whether or not the issue is "critical." If additional
    levels are needed in the future, we will add them.
    - (If needed) Clean up the title to succinctly and clearly state the issue. Also ensure
    that proposals are prefaced with "Proposal".
    - Add the issue to the correct milestone. If any questions come up, don't worry about
    adding the issue to a milestone until the questions are answered.
    - We attempt to do this process at least once per work day.
3. Discussion
    - "Feature" and "Bug" issues should be connected to the PR that resolves it.
    - Whoever is working on a "Feature" or "Bug" issue (whether a maintainer or someone from
    the community), should either assign the issue to them self or make a comment in the issue
    saying that they are taking it.
    - "Proposal" and "Question" issues should stay open until resolved or if they have not been
    active for more than 30 days. This will help keep the issue queue to a manageable size and
    reduce noise. Should the issue need to stay open, the `keep open` label can be added.
4. Issue closure

## Contribution process
This is a rough outline of what a contributor's workflow looks like:

- Create a branch from where you want to base your work (usually master).
- Make your changes and arrange them in readable commits.
- Make sure your commit messages are in the proper format (see below).
- Push your changes to the branch in your fork of the repository.
- Make sure all tests pass, and add any new tests as appropriate.
- Submit a pull request to the original repository.


## Pull Requests (PR)

Like any good open source project, we use Pull Requests to track code changes

### PR Lifecycle

1. PR creation
    - We more than welcome PRs that are currently in progress. They are a great way to keep track of
    important work that is in-flight, but useful for others to see. If a PR is a work in progress,
    it **must** be prefaced with "WIP: [title]". Once the PR is ready for review, remove "WIP" from
    the title.
    - It is preferred, but not required, to have a PR tied to a specific issue.
2. Triage
    - The maintainer in charge of triaging will apply the proper labels for the issue. This should
    include at least a size label, `bug` or `feature`, and `awaiting review` once all labels are applied.
    See the [Labels section](#labels) for full details on the definitions of labels
    - Add the PR to the correct milestone. This should be the same as the issue the PR closes.
3. Assigning reviews
    - Once a review has the `awaiting review` label, maintainers will review them as schedule permits.
    The maintainer who takes the issue should self-request a review.
    - Reviews from others in the community, especially those who have encountered a bug or have
    requested a feature, are highly encouraged, but not required. Maintainer reviews **are** required
    before any merge
    - Any PR with the `size/large` label requires 2 review approvals from maintainers before it can be
    merged. Those with `size/medium` are per the judgement of the maintainers
4. Reviewing/Discussion
    - Once a maintainer begins reviewing a PR, they will remove the `awaiting review` label and add
    the `in progress` label so the person submitting knows that it is being worked on. This is
    especially helpful when the review may take awhile.
    - All reviews will be completed using Github review tool.
    - A "Comment" review should be used when there are questions about the code that should be
    answered, but that don't involve code changes. This type of review does not count as approval.
    - A "Changes Requested" review indicates that changes to the code need to be made before they will be merged.
    - Reviewers should update labels as needed (such as `needs rebase`)
5. Address comments by answering questions or changing code
6. Merge or close
    - PRs should stay open until merged or if they have not been active for more than 30 days.
    This will help keep the PR queue to a manageable size and reduce noise. Should the PR need
    to stay open (like in the case of a WIP), the `keep open` label can be added.
    - If the owner of the PR is listed in `OWNERS`, that user **must** merge their own PRs
    or explicitly request another OWNER do that for them.
    - If the owner of a PR is _not_ listed in `OWNERS`, any core committer may
    merge the PR once it is approved.

#### Documentation PRs

Documentation PRs will follow the same lifecycle as other PRs. They will also be labeled with the
`docs` label. For documentation, special attention will be paid to spelling, grammar, and clarity
(whereas those things don't matter *as* much for comments in code).

## The Triager

Each week, one of the core maintainers will serve as the designated "triager" starting after the
public standup meetings on Thursday. This person will be in charge triaging new PRs and issues
throughout the work week.

## Labels

The following tables define all label types used for Alamed. It is split up by category.

### Common

| Label | Description |
| ----- | ----------- |
| `bug` | Marks an issue as a bug or a PR as a bugfix |
| `critical` | Marks an issue or PR as critical. This means that addressing the PR or issue is top priority and will be handled first by maintainers |
| `docs` | Indicates the issue or PR is a documentation change |
| `duplicate` | Indicates that the issue or PR is a duplicate of another |
| `feature` | Marks the issue as a feature request or a PR as a feature implementation |
| `keep open` | Denotes that the issue or PR should be kept open past 30 days of inactivity |
| `refactor` | Indicates that the issue is a code refactor and is not fixing a bug or adding additional functionality |

### Issue Specific

| Label | Description |
| ----- | ----------- |
| `help wanted` | This issue is one the core maintainers cannot get to right now and would appreciate help with |
| `proposal` | This issue is a proposal |
| `question/support` | This issue is a support request or question |
| `starter` | This issue is a good for someone new to contributing to Alameda|
| `wont fix` | The issue has been discussed and will not be implemented (or accepted in the case of a proposal) |

### PR Specific

| Label | Description |
| ----- | ----------- |
| `awaiting review` | The PR has been triaged and is ready for someone to review |
| `breaking` | The PR has breaking changes (such as API changes) |
| `in progress` | Indicates that a maintainer is looking at the PR, even if no review has been posted yet |
| `needs pick` | Indicates that the PR needs to be picked into a feature branch (generally bugfix branches). Once it has been, the `picked` label should be applied and this one removed |
| `needs rebase` | A helper label used to indicate that the PR needs to be rebased before it can be merged. Used for easy filtering |
| `picked` | This PR has been picked into a feature branch |

#### Size labels

Size labels are used to indicate how "dangerous" a PR is. The guidelines below are used to assign the
labels, but ultimately this can be changed by the maintainers. For example, even if a PR only makes
30 lines of changes in 1 file, but it changes key functionality, it will likely be labeled as `size/large`
because it requires sign off from multiple people. Conversely, a PR that adds a small feature, but requires
another 150 lines of tests to cover all cases, could be labeled as `size/small` even though the number
lines is greater than defined below.

| Label | Description |
| ----- | ----------- |
| `size/small` | Anything less than or equal to 4 files and 150 lines. Only small amounts of manual testing may be required |
| `size/medium` | Anything greater than `size/small` and less than or equal to 8 files and 300 lines. Manual validation should be required. |
| `size/large` | Anything greater than `size/medium`. This should be thoroughly tested before merging and always requires 2 approvals. This also should be applied to anything that is a significant logic change. |

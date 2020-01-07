# Contributing
When making changes, please first either find the corresponding Github issue or create it if it does not exist. When creating an issue, add appropriate labels to it.

DO NOT COMMIT TO MASTER, please branch and PR any changes to master. This gives an opportunity for others to both code review your work and also see changes going into master, as opposed to new code simply appearing.

## Branch Guidelines
As you now have a Github issue tied to your changes, name your branch after your Github issue number, and optionally a few words after that describe the issue.

### Example
I make an issue to fix a bug with the "write" button. My issue shows up as #14, so I will create a branch titled

`14/fix-write-button`

## PR Guidelines
You may open a PR at any time once the first commit has been made. If the PR is not yet ready for review or merge, simply add the "WIP" (Work In Progress) label to it. Putting it up early allows other developers to see what you are working on and how far along it is, as well as allowing for architectural or design errors to be caught early on.

### When creating the PR
1. Give it a meaningful description that helps another developer understand what you have changed *and why*. It will typically be similar to the issue title, or a bit more descriptive.
2. Add any appropriate labels
3. Assign a reviewer and set the assignee to yourself

### Before removing the WIP label
1. Run `go fmt` if you worked on any Go code.
2. Ensure there are no linting errors for any Javascript code.
3. Ensure your code compiles and runs as expected!
4. Sync your branch with master. This can be done with:
```
git checkout master
git pull
git checkout -
git merge master
```
5. Now create the PR and add the "Ready for Review" label!

### Example
Continuing on our example from before, our PR could be named

`[14] Fix async race condition with the write button`

## Commit Guidelines
Be sure to prefix every commit with your issue number. For example, `[14] Lint race condition fix`

This section is a lot more optional, but in general please follow general commit message practices. Common guidelines outlined here: https://chris.beams.io/posts/git-commit/
In summary,
1. Separate subject from body with a blank line (if you have a body, which is optional)
2. Limit the subject line to 50 characters
3. Capitalize the subject line
4. Do not end the subject line with a period
5. Use the imperative mood in the subject line
6. Wrap the body at 72 characters
7. Use the body to explain what and why vs. how

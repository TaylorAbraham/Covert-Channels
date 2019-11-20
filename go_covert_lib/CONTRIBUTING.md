# Contributing

When making changes, please first either find the corresponding JIRA issue or create it if it does not exist.

DO NOT COMMIT TO MASTER, please branch and PR any changes to master. This gives an opportunity for others to both code review your work and also see changes going into master, as opposed to new code simply appearing.

## Branch Naming

As you now have a JIRA issue tied to your changes, name your branch after your JIRA ticket number, and optionally a word or two after that describing the ticket.

`CC-42/fix-submit-button`

## PR Guidelines

Before creating a PR, ensure to do the following:
1. Run `go fmt` if you worked on any go code.
2. Ensure your code compiles and runs as expected!
3. Sync your branch with master. This can be done with `git checkout master`, `git pull`, `git checkout -`, and `git merge master` in that order, or any method you understand the implications of.

When creating the PR:
1. Give it a meaningful description that helps another developer understand what you have changed *and why*
2. Add any appropriate labels
3. Assign a reviewer and set the assignee to yourself

## Commit Guidelines

This section is a lot more optional, but in general please follow general commit message practices. Common guidelines outlined here: https://chris.beams.io/posts/git-commit/
In summary,
1. Separate subject from body with a blank line
2. Limit the subject line to 50 characters
3. Capitalize the subject line
4. Do not end the subject line with a period
5. Use the imperative mood in the subject line
6. Wrap the body at 72 characters
7. Use the body to explain what and why vs. how

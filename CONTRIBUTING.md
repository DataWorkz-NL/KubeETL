# Contributing

Thanks for showing interest in KubeETL. We welcome contributions of all kinds. For example:

- New features, bug fixes, and other improvements to the code.
- Documentation improvements.
- Examples of usage.
- Bug reports.
- Feature requests.

If you have any questions feel free to create a Github issue.

## Contribution process

Small changes can be submitted directly via a Pull Request. You can expect the maintainers to react within a reasonable timespan on the PR.

Larger changes should be accompanied by an associated Github issue. The issue should detail what use cases the change covers and what the proposed solution is.

## Code changes

For code contributions, adhere to the following guidelines:

- Most code changes should be accompanied by unit tests.
- For larger changes, integration tests should be included.
- All tests must pass.
- The CI/CD pipeline must succeed.
- All commits must be signed off (see the next section for details).

Before the code change will be merged, it will be reviewed by one of the maintainers. Once the change has been approved, one of the maintainers will squash merge the Pull Request.

## Commit sign off

The KubeETL project requires that all contributions are signed off. The [Developer Certificate of Origin (DCO)](https://developercertificate.org/) is a simple way to certify that you wrote or have the right to submit the code you are contributing to the project.

You can sign-off on your commits by adding the following at the end of your commit:

```text
Signed-off-by: Random Contributor <random@contributor.example.org>
```

Git has a -s command line option to do this automatically: `git commit -s`.

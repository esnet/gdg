# Contributing

Contributions are welcome, and they are greatly appreciated! Every little bit helps, and credit will always be given.

You can contribute in many ways:

## Types of Contributions

### Report Bugs

Report bugs at https://github.com/esnet/gdg/issues.

If you are reporting a bug, please include:

* Your operating system name and version.
* Any details about your local setup that might be helpful in troubleshooting.
* Detailed steps to reproduce the bug.

### Code Submissions

Any code submitted to enhance this project is great appreciated, but here's a check list to look at which will help ensure a successful code review.

1. Make sure the code works, compiles and so on.  
2. We have a docker-compose file that will bring up an instance of grafana.  A variety of integration tests exists under the `integration_tests` folder.  Ideally each new entity we introduce should have tests that go with it.  Please make sure you have a test for you code submission.
3. Configuration Changes: Please update the conf/gdg-example.yml if you are introducing any new configs 
4. Document the code, not every line needs docs, but explaining what the function does is helpful for those that follow.
5. The generated docs live under `documentation/content/en/docs/` Please update the md files to reflect your changes to let others know how to use the tool.  `usage_guide.md` is likely the only file you'll need to update.
    - documentation/content/en/docs/releases contains a list of changes, if you are adding a new feature, please add a short description of the upcoming new feature.
    - documentation/content/en/docs/usage_guide.md provides simple examples.  If you're adding a new entity please document its behavior.
6. If you've introduced a major change that affects the configuration, please take the time to update the context new wizard to ensure it's captured.


### AI-Assisted Contributions

We welcome contributions made with the help of AI tools (Claude, Copilot, ChatGPT, Cursor, and similar). These tools are fine to use, but they don't change who is responsible for the work.

**You are ultimately responsible for everything you submit.** If you open a pull request, issue, or comment, you are the author of record, whether you typed every line or an AI generated it. "The AI wrote it" is not an acceptable explanation for a bug, a license problem, or a low-quality submission.

If you use AI tools, please follow these guidelines:

1. **Understand your changes.** You should be able to explain every line you submit and why it is there. If you can't explain it in review, don't submit it.
2. **Review and test it yourself.** Read the generated code, run it, and make sure it works and passes `make test`. AI output can look plausible and still be wrong, so verify behavior rather than assuming it.
3. **Follow the project's conventions.** AI-generated code must meet the same standards as any other contribution: tests, docs, config updates, and the checklist under "Code Submissions".
4. **Check licensing and provenance.** By submitting, you confirm you have the right to contribute the code under this project's license. Be aware that wholly AI-generated output may not be copyrightable and may reproduce code from other sources, so review it, adapt it, and don't submit anything you can't vouch for.
5. **Keep submissions reviewable.** Large, unreviewed, or speculative AI-generated changes put an unfair burden on maintainers, who are volunteers. Keep pull requests focused and sized so they can be reviewed. A contribution should be worth more to the project than the time it takes to review it.
6. **Disclose significant AI assistance.** We encourage you to note AI involvement in your commit message using a trailer, for example `Assisted-by: <tool and version>` for minor help or `Generated-by: <tool and version>` for substantially AI-written code. Disclosure helps reviewers and keeps licensing provenance clear.
7. **Respect maintainers' time.** Submissions that are clearly unreviewed or that the author can't explain (for example, references to APIs or flags that don't exist, or answers that don't address review feedback) may be closed without further review. This applies regardless of whether AI was involved.

### Submit Feedback

The best way to send feedback is to file an issue at https://github.com/esnet/gdg/issues.

If you are proposing a feature:

* Explain in detail how it would work.
* Keep the scope as narrow as possible, to make it easier to implement.
* Remember that this is a volunteer-driven project, and that contributions
  are welcome :)

## Get Started!

Ready to contribute? Here's how to set up `gdg` for local development.

1. Fork the `gdg` repo on GitHub.
2. Clone your fork locally::
```bash
    $ git clone git@github.com:your_name_here/gdg.git
```
3. Create a branch for local development::
```bash
    $ git checkout -b name-of-your-bugfix-or-feature
```
   Now you can make your changes locally.

4. When you're done making changes, check that your changes pass the tests::
```bash
    $ make test
```
6. Commit your changes and push your branch to GitHub::
```bash
    $ git add .
    $ git commit -m "Your detailed description of your changes."
    $ git push origin name-of-your-bugfix-or-feature
```
7. Submit a pull request through the GitHub website.

Pull Request Guidelines
-----------------------

Before you submit a pull request, check that it meets these guidelines:

1. The pull request should include tests.
2. If the pull request adds functionality, the docs should be updated. Put
   your new functionality into a function with a docstring, and add the
   feature to the list in README.md.
3. If AI tools were used, see [AI-Assisted Contributions](#ai-assisted-contributions) above. You remain responsible for the submission.

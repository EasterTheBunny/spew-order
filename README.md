[![Netlify Status](https://api.netlify.com/api/v1/badges/bfa51a4c-7883-4f24-86c2-1f58a79ae0c9/deploy-status)](https://app.netlify.com/sites/ciphermtn/deploys)

# CI/CD Pipeline
The pipeline involves CircleCI and includes 3 main environments: ci, uat, and prod. The git repo has 3 corresponding branches with the same name. Merging a PR from ci to uat, for example, would move that change to the uat environment per the configured triggers in the pipeline.

## Environment Configurations
Each environment has a set of configuration values stored as encrypted files in the repo managed by `git-secrets` and `gpg`. An appropriate gpg key for the desired environment is required to read or edit these files. The following is an example setup for unlocking the ci environment configuration files.

```
$ export SECRETS_EXTENSION=".secret-ci"
$ export SECRETS_DIR=".gitsecret-ci"
$ git secret reveal
```

DO NOT REMOVE THESE FILES FROM `.gitignore`!!!

Once edits are done, do the following:

```
$ git secret hide
```

To add a new accessor to the environment, run the following:

```
$ git secret tell <user@email.com>
```

For more information, visit the documentation: [git-secret docs](https://git-secret.io/#commands) [gpg docs](https://keyring.debian.org/creating-key.html)

To set up default credentials, do the following:
```
$ export GOOGLE_APPLICATION_CREDENTIALS=/workspaces/spew-order/configurations/key.json
```
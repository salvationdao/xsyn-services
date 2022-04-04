# Github Actions

## Testing

<https://github.com/nektos/act> allows you to run actions locally.
It supports most steps that github can run.  
Some steps like `uses: actions/upload-artifact@v2` can not be run but can be skipped in testing like so ...

```yaml
- name: save artifact
  uses: actions/upload-artifact@v2
  if: ${{ !env.ACT }} # if running in local test environment
  with:
    name: ${{ steps.artifact.outputs.filename }}
    path: ${{ steps.artifact.outputs.filename }}
```

Some steps like `run: rsync ....` should not be run while testing and should be ignored like above.

List jobs

```sh
$ act --list
Stage  Job ID   Job name  Workflow name       Workflow file       Events
0      build    build     Staging Deployment  deploy-staging.yml  push
0      Release  Release   Publish Release     release.yml         push
1      deploy   deploy    Staging Deployment  deploy-staging.yml  push
```

Run a specific job

```sh
$ act --job build # case sensitive
[Staging Deployment/build] 🚀  Start image=ghcr.io/catthehacker/ubuntu:act-20.04
...
```

### Common tests

```sh
# staging build job
$ act --job build --secret GH_DEPLOY_BRIDGE_PK="$(cat ~/.ssh/work/GH_DEPLOY_BRIDGE_PK)" --secret GH_DEPLOY_HUB_PK="$(cat ~/.ssh/work/GH_DEPLOY_HUB_PK)" --secret GH_DEPLOY_LOGHELPERS_PK="$(cat ~/.ssh/work/GH_DEPLOY_LOGHELPERS_PK)" --secret GH_DEPLOY_SALE_PK="$(cat ~/.ssh/work/GH_DEPLOY_SALE_PK)"
[Staging Deployment/build] 🚀  Start image=ghcr.io/catthehacker/ubuntu:act-20.04
...
```

```sh
# release job
$ act --job Release --secret GH_DEPLOY_BRIDGE_PK="$(cat ~/.ssh/work/GH_DEPLOY_BRIDGE_PK)" --secret GH_DEPLOY_HUB_PK="$(cat ~/.ssh/work/GH_DEPLOY_HUB_PK)" --secret GH_DEPLOY_LOGHELPERS_PK="$(cat ~/.ssh/work/GH_DEPLOY_LOGHELPERS_PK)" --secret GH_DEPLOY_SALE_PK="$(cat ~/.ssh/work/GH_DEPLOY_SALE_PK)"
[Staging Deployment/build] 🚀  Start image=ghcr.io/catthehacker/ubuntu:act-20.04
...
```

## Build and Deploy Staging

### Required Secrets

```sh
# Private Repo Access
# Key pairs available in 1password / DevOps vault
GH_DEPLOY_BRIDGE_PK     # Deploy private key for https://github.com/ninja-syndicate/supremacy-bridge
GH_DEPLOY_HUB_PK        # Deploy private key for https://github.com/ninja-syndicate/hub
GH_DEPLOY_LOGHELPERS_PK # Deploy private key for https://github.com/ninja-software/log_helpers
GH_DEPLOY_SALE_PK       # Deploy private key for https://github.com/ninja-software/sale/

# Deployment target
STAGING_SSH_HOST # The host name of the target
STAGING_SSH_KNOWN_HOSTS # ssh-keyscan -t ED25519 -p STAGING_SSH_PORT STAGING_SSH_HOST
STAGING_SSH_PKEY
STAGING_SSH_PORT
STAGING_SSH_USER
```

### Jobs

#### Build

1. Set up private repo access using multiple deploy keys so the keys have read only access to the private repos.
    [multi key private go repos](https://gist.github.com/jrapoport/d12f60029eef017354d0ec982b918258).

    The basic steps are
    1. Create a host alias for github in `.ssh/config` with the private key file location
    2. Create the private key file
    3. Add the full repo url (ie `https://github.com/ninja-syndicate/hub`) to the git config override

2. Get the most recent version, increment by one and create a new tag, depending on branch.  
   if develop, use git hash `$PREVIOUS_TAG-dev-$GITHASH`.  
   if staging, use `$PREVIOUS_TAG-rc.1`, incrementing the `rc.` number.  
   else fail.  

   Also set any version related environment variables  
   **DEBUGGING**  
   - Error `fatal: tag 'x.y.z-rc.n' already exists`  
     Reason: There is a commit with 2 tags
     Response: Switch to staging and increment the tag on a untagged commit.
     `git switch staging && git commit --allow-empty -m 'tag error fix' && git tag x.y.z-rc.n+1 && git commit --allow-empty -m 'tag error fix'`
   - Error `fatal: No names found, cannot describe anything.`  
     Reason: There are no existing tags  
     Response: ensure that `actions/checkout@v2` has `fetch-depth: "0"`  
     Response alternitive: manually tag a commit on the same branch that the CI runs this  script on
   - Error `fatal: something something GITHASH`  
     This happens occasionally and I can't reproduce.
     Response: Contact Nathan so it can be resolved and documented.

3. Set private repo environment variables, build tools.  

4. Copy required files like configs, samples and migrations.

5. Build API server with version related environment variables.

6. Generate build info.

7. Push new tag to github.

8. Save build output to github actions.

#### Deploy

1. Wait for build to finish.

2. Retrive build output from previous job.

3. Set up ssh config using the deployment target secrets.

4. Rsync the files to `/usr/share/ninja_syndicate/passport-api_${{env.GITVERSION}}`

5. Copy the environment file from `passport-api_online` to new the new version.

6. BROKEN: backup the database

7. Update the `passport-api_online` symbolic link to the new version

8. Drop the database

9. Run migrate up

10. Restart services

## Publish Release

### Required Secrets

```sh
# Private Repo Access
# Key pairs available in 1password / DevOps vault
GH_DEPLOY_BRIDGE_PK     # Deploy private key for https://github.com/ninja-syndicate/supremacy-bridge
GH_DEPLOY_HUB_PK        # Deploy private key for https://github.com/ninja-syndicate/hub
GH_DEPLOY_LOGHELPERS_PK # Deploy private key for https://github.com/ninja-software/log_helpers
GH_DEPLOY_SALE_PK       # Deploy private key for https://github.com/ninja-software/sale/
```

### Jobs

#### Release

1. Set up private repo access using multiple deploy keys so the keys have read only access to the private repos.
    [multi key private go repos](https://gist.github.com/jrapoport/d12f60029eef017354d0ec982b918258).

    The basic steps are
    1. Create a host alias for github in `.ssh/config` with the private key file location
    2. Create the private key file
    3. Add the full repo url (ie `https://github.com/ninja-syndicate/hub`) to the git config override

2. Set private repo environment variables, build tools.  

3. Copy required files like configs, samples and migrations.

4. Get build metadata like current commit hash, build date, branch, commit tag

5. Build API server with version related environment variables.

6. Generate build info.

7. tar.gz build result.

9. Publish draft release with the tar.gz.

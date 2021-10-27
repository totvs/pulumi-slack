# Permissions
We need to configure some scopes to manage resources (Bot tokens):

1. Open https://api.slack.com/apps
2. Create new App
3. Configure App Name and Workspace to develop
4. Use the manifest below
```yaml
_metadata:
  major_version: 1
  minor_version: 1
display_information:
  name: pulumi-resource-slack
features:
  bot_user:
    display_name: pulumi-resource-slack
    always_online: false
oauth_config:
  scopes:
    bot:
      - channels:manage
      - groups:write
      - im:write
      - mpim:write
      - groups:read
      - channels:read
      - im:read
      - mpim:read
settings:
  org_deploy_enabled: false
  socket_mode_enabled: false
  token_rotation_enabled: false
```
5. Open menu Oauth & Permissions
6. Click on Install to Workspace and allow requested permissions
7. Copy the bot user oauth token to be used by provider

## Other tips
If you are facing problems to manage channel members, even if you add permission scopes to the bot user, check the channel management.

https://YOUR-WORKSPACE-HERE.slack.com/admin/settings#channel_management_restrictions
 
## Environment
You need to set Slack token with:

```sh
export SLACK_TOKEN xoxb-2271973586641-3369578963123-hngThgT5dq4W7vmHdzd91T3H
```
or
```sh
pulumi config set --secret slack:config:token xoxb-2271973586641-3369578963123-hngThgT5dq4W7vmHdzd91T3H
```

## Pre-requisites to develop

Install the `pulumictl` cli from the [releases](https://github.com/pulumi/pulumictl/releases) page or follow the [install instructions](https://github.com/pulumi/pulumictl#installation)

> NB: Usage of `pulumictl` is optional. If not using it, hard code the version in the [Makefile](Makefile) of when building explicitly pass version as `VERSION=0.0.1 make build`

## Build and Test

```bash
# build and install the resource provider plugin
$ make build install

# test
$ cd examples/ts
$ yarn link @pulumi/slack
$ yarn install
$ pulumi stack init test
$ pulumi up
```

# Usage

Install Plugin

```bash
pulumi plugin install resource slack 0.0.1 -f /tmp/pulumi-resource-slack-xxxxx.tar.gz
```

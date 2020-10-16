# aksctl

aksctl is a proof of concept tool that allows you to create [Azure AKS](https://azure.microsoft.com/en-us/services/kubernetes-service/) clusters quickly and easily from the command line.

It was inspired by its namesake, [eksctl](https://github.com/weaveworks/eksctl) and its purpose is to show how easy it is to create useful provisioning tools with the [Pulumi](https://pulumi.com/) [Automation API](pulumi.com/)

> :warning: **This is a proof of concept, and not designed to be used anywhere near your production environment**: Look at the code and be inspired!

## Background

eksctl is a fantastic tool for creating EKS clusters with a wonderful command line UX, but it only works in AWS for EKS!

It's possible that, we could create an eksctl like tool for AKS using the native Azure Go SDKs, but then the author of said tool has to implement lots of non-trivial logic to manage state, retries and handle errors.

With the Pulumi automation API, almost all of this logic is taken care of for you. You can write your infrastructure provisioning code with the [Pulumi Azure SDK](https://www.pulumi.com/docs/reference/pkg/azure-nextgen/), and then use the automation API and a sprinkling of cobra to wrap it up into a CLI tool that can be distributed quickly and easily.

## Usage

You'll need a Pulumi account to store your state. 

*Future versions may add support for open source backends*

Once you've logged in, you can create a cluster like so:

```bash

aksctl create cluster --name aksctl -o jaxxstorm -p aksctl -s test
```

Here I'm specifying my org, the name of my cluster and a pulumi project and stack to use.

Destroy my cluster is just as easy:

```
aksctl delete cluster --name aksctl -o jaxxstorm -p aksctl -s test
```

## Configuration

When you create a cluster, you need to specify a Pulumi org to operate in. You can specify that using the `-o` flag on the command line, but that gets unwieldy, especially considering it's not changed very often

You can alternatively specify a global organization in a config file:

```bash
cat ~/.aksctl/config.yml
org: jaxxstorm
```





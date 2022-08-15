# Yatai Helm Chart

The Yatai Helm Chart is the official way to operate Yatai on Kubernetes. It contains all the required components to get started, and can configure with external services base on needs.

See the [Yatai administrator's guide](https://github.com/bentoml/Yatai/blob/main/docs/admin-guide.md) for how to install Yatai and other information on charts, and advanced configuration.

Advantage of using Yatai Helm chart:

* Easy to deploy, upgrade, and maintain Yatai service on Kubernetes cluster
* Easy to configure with external services
* Up to date with the latest Yatai release

This is the top level Yatai chart, which configures all the necessary components to run Yatai. With this chart, a user can make decisions like:

* Use default created service account or use existing service account.
* Use the default PostgreSQL or use an external database like AWS RDS.
* Use external object storage services like AWS S3 or use the default Minio.
* Use external Docker registry or use the default Docker registry.


## TL;DR:

```bash
helm repo add yatai https://bentoml.github.io/yatai-chart
helm repo update
helm install yatai yatai/yatai -n yatai-system --create-namespace
```

## Helm chart deployment overview

This chart will create the following resources on Kubernetes:
1. Yatai service under the `yatai-system` namespace.
2. Default PostgreSQL under the `yatai-system` namespace (if not configured).
3. service account (if not configured).

After Yatai is running, it will create the following resources on Kuberenetes:
1. `yatai-operators`, `yatai-components`, `yatai-builders` and `yatai` namespaces.
2. `yatai-deployment-comp-operator` under the `yatai-operators` namespace.
3. Nginx Ingress Controller under the `yatai-components` namespace using `yatai-deployment-comp-operator`.

# Community

- To report a bug or suggest a feature request, use [GitHub Issues](https://github.com/bentoml/yatai-chart/issues/new/choose).
- For other discussions, use [Github Discussions](https://github.com/bentoml/BentoML/discussions) under the [BentoML repo](https://github.com/bentoml/BentoML/)
- To receive release announcements and get support, join us on [Slack](https://join.slack.com/t/bentoml/shared_invite/enQtNjcyMTY3MjE4NTgzLTU3ZDc1MWM5MzQxMWQxMzJiNTc1MTJmMzYzMTYwMjQ0OGEwNDFmZDkzYWQxNzgxYWNhNjAxZjk4MzI4OGY1Yjg).


# Contributing

There are many ways to contribute to the project:

- If you have any feedback on the project, share it with the community in [Github Discussions](https://github.com/bentoml/BentoML/discussions) under the [BentoML repo](https://github.com/bentoml/BentoML/).
- Report issues you're facing and "Thumbs up" on issues and feature requests that are relevant to you.
- Investigate bugs and reviewing other developer's pull requests.
- Contributing code or documentation to the project by submitting a Github pull request.

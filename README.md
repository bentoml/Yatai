# 🦄️ Yatai: Model serving at scale on Kubernetes

[![actions_status](https://github.com/bentoml/yatai/workflows/CICD/badge.svg)](https://github.com/bentoml/yatai/actions)
[![join_slack](https://badgen.net/badge/Join/BentoML%20Slack/cyan?icon=slack)](https://join.slack.com/t/bentoml/shared_invite/enQtNjcyMTY3MjE4NTgzLTU3ZDc1MWM5MzQxMWQxMzJiNTc1MTJmMzYzMTYwMjQ0OGEwNDFmZDkzYWQxNzgxYWNhNjAxZjk4MzI4OGY1Yjg)

Yatai makes it easy to deploy and operate machine learning serving workload at scale on Kubernetes.
Yatai is built upon [BentoML](https://github.com/bentoml), the unified model serving framework.

Core features:

* **Bento Registry** - manage all your team's Bentos and Models, backed by cloud blob storage(S3, MinIO)
* **Deployment Automation** - deploy Bentos as auto-scaling API endpoints on Kubernetes and easily rollout new versions
* **Observability** - monitoring dashboard helping users to identify model performance issues
* **CI/CD** - flexible APIs for integrating with your training and CI pipelines


## Why Yatai

* Yatai is built upon [BentoML](https://github.com/bentoml/BentoML), the unified model serving framework that is high-performing and feature-rich
* Yatai focus on the model serving and deployment part of your MLOps stack, works well with any ML training/monitoring platforms, such as AWS SageMaker or MLFlow
* Yatai is Kubernetes native, integrates well with other cloud native tools in the K8s eco-system
* Yatai is human-centric, provides easy-to-use Web UI and APIs for ML scientists, MLOps engineers, and project managers


## Getting Started

1. Create an ML Service with BentoML following the [Quickstart Guide](https://docs.bentoml.org/en/latest/quickstart.html) or sample projects in the [BentoML Gallery](https://github.com/bentoml/gallery).

2. Install Yatai locally with Minikube.
   * Prerequisites:
     * Install latest minikube: https://minikube.sigs.k8s.io/docs/start/
     * Install latest Helm: https://helm.sh/docs/intro/install/
   * Start a minikube Kubernetes cluster: `minikube start --cpus 4 --memory 4096`
   * Install Yatai Helm Chart:
     ```bash
     helm repo add yatai https://bentoml.github.io/yatai-chart
     helm repo update
     helm install yatai yatai/yatai -n yatai-system --create-namespace
     ```
   * Wait for installation to complete: `helm status yatai -n yatai-system`
   * Start minikube tunnel for accessing Yatai UI: `minikube tunnel`
   * Open browser at http://yatai.127.0.0.1.sslip.io and login with default account `admin:admin`

3. See [Administrator's Guide](https://github.com/bentoml/yatai/blob/main/docs/admin-guide.md) for a comprehensive overview for deploying and configuring Yatai for production use.


## Community

- To report a bug or suggest a feature request, use [GitHub Issues](https://github.com/bentoml/yatai/issues/new/choose).
- For other discussions, use [Github Discussions](https://github.com/bentoml/BentoML/discussions) under the [BentoML repo](https://github.com/bentoml/BentoML/)
- To receive release announcements and get support, join us on [Slack](https://join.slack.com/t/bentoml/shared_invite/enQtNjcyMTY3MjE4NTgzLTU3ZDc1MWM5MzQxMWQxMzJiNTc1MTJmMzYzMTYwMjQ0OGEwNDFmZDkzYWQxNzgxYWNhNjAxZjk4MzI4OGY1Yjg).


## Contributing

There are many ways to contribute to the project:

- If you have any feedback on the project, share it with the community in [Github Discussions](https://github.com/bentoml/BentoML/discussions) under the [BentoML repo](https://github.com/bentoml/BentoML/).
- Report issues you're facing and "Thumbs up" on issues and feature requests that are relevant to you.
- Investigate bugs and reviewing other developer's pull requests.
- Contributing code or documentation to the project by submitting a Github pull request. See the [development guide](https://github.com/bentoml/yatai/blob/main/DEVELOPMENT.md).


## Licence

[Elastic License 2.0 (ELv2)](https://github.com/bentoml/yatai/blob/main/LICENSE.md)

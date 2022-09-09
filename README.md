# 🦄️ Yatai: Production-first ML platform on Kubernetes

[![actions_status](https://github.com/bentoml/yatai/workflows/CICD/badge.svg)](https://github.com/bentoml/yatai/actions)
[![docs](https://readthedocs.org/projects/yatai/badge/?version=latest&style=flat-square)](https://docs.bentoml.org/projects/yatai)

[![join_slack](https://badgen.net/badge/Join/BentoML%20Slack/cyan?icon=slack&style=flat-square)](https://join.slack.bentoml.org)

Yatai is a production-first platform for your machine learning needs. It brings collaborative [BentoML](https://github.com/bentoml) workflows to Kubernetes, helps ML teams to run model serving at scale, while simplifying model management and deployment across teams.

👉 [Pop into our Slack community!](https://l.linklyhq.com/l/ktPW) We're happy to help with any issue you face or even just to meet you and hear what you're working on :)

## Why Yatai?

-   Yatai accelerates the process of taking ML models from training stage to production and reduces the operational overhead of running a reliable model serving system.

-   Yatai simplifies collaboration between Data Science and Engineering teams. It is designed to leverage the BentoML standard and streamline production ML workflows.

-   Yatai is a cloud native platform with a wide range of integrations to best fit your infrastructure needs, and it is easily customizable for your CI/CD needs.

## Core features:

-   **Bento Registry** - manage all your team's ML models via simple Web UI and API, and store ML assets on cloud blob storage
-   **Deployment Automation** - deploy Bentos as auto-scaling API endpoints on Kubernetes and easily rollout new versions
-   **Observability** - monitoring dashboard and logging integration helping users to identify model performance issues
-   **CI/CD** - flexible APIs for integrating with your training and CI pipelines

<img width="785" alt="yatai-overview-page" src="https://user-images.githubusercontent.com/489344/151455964-4fe30eb7-f000-43cc-8a5f-807ee450b8b6.png">

<details>

  <summary>See more product screenshots</summary>

  <img width="785" alt="yatai-deployment-creation" src="https://user-images.githubusercontent.com/489344/151456002-d4e9f84d-8a71-4bf9-bde7-f94a74abbf3f.png">
  <img width="785" alt="yatai-bento-repos" src="https://user-images.githubusercontent.com/489344/151456379-da255519-274d-41de-a1b9-a347be279230.png">
  <img width="785" alt="yatai-model-detail" src="https://user-images.githubusercontent.com/489344/151456021-360a6d6e-acb8-494b-9f6b-868ef9d13bce.png">
  <img width="785" alt="yatai-cluster-components" src="https://user-images.githubusercontent.com/489344/151456017-abf0c77a-ba8a-43e5-8949-901ef4a8074a.png">
  <img width="785" alt="yatai-deployment-details" src="https://user-images.githubusercontent.com/489344/151456024-151c275d-b33e-480e-be34-dadab5b01915.png">
  <img width="785" alt="yatai-activities" src="https://user-images.githubusercontent.com/489344/151456011-69c283bc-7382-4b30-bfbf-2686e2abdc0f.png">

</details>

## Getting Started

-   [Documentation](https://docs.bentoml.org/projects/yatai/) - Overview of the Yatai docs and related resources
-   [Installation](https://docs.bentoml.org/projects/yatai/en/latest/installation/index.html) - Hands-on instruction on how to install Yatai for production use

## Quick Tour

Here's a quick tour for trying out Yatai locally. 

#### Prerequisites:
  * Install latest minikube: https://minikube.sigs.k8s.io/docs/start/
  * Install latest Helm: https://helm.sh/docs/intro/install/
  * Start a minikube Kubernetes cluster: `minikube start --cpus 4 --memory 4096`
  * Check that minikube cluster status is "running": `minikube status`
  * Make sure your `kubectl` is configured with `minikube` context: `kubectl config current-context`
  * Enable ingress controller: `minikube addons enable ingress`

#### Install Yatai

Install Yatai with the following script:

```bash
DEVEL=true bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/main/scripts/quick-install-yatai.sh")
```

This script will install Yatai along with its dependencies (PostgreSQL and MinIO) on
your minikube cluster. 

Note that this installation script is made for development and testing use only.
For production deployment, check out the [Installation Guide](https://docs.bentoml.org/projects/yatai/en/latest/installation/index.html).

To access Yatai web UI, run the following command and keep the terminal open:

```bash
kubectl --namespace yatai-system port-forward svc/yatai 8080:80
```

In a separate terminal, run:

```bash
YATAI_INITIALIZATION_TOKEN=$(kubectl get secret env --namespace yatai-system -o jsonpath="{.data.YATAI_INITIALIZATION_TOKEN}" | base64 --decode)
echo "Open in browser: http://127.0.0.1:8080/setup?token=$YATAI_INITIALIZATION_TOKEN"
``` 

Open the URL printed above from your browser to finish admin account setup.


#### Push Bento to Yatai

First, get an API token and login BentoML CLI:

* Keep the `kubectl port-forward` command in step above running
* Go to Yatai's API tokens page: http://127.0.0.1:8080/api_tokens
* Create a new API token form the UI, make sure to assign "API" access under "Scopes"
* Copy login command upon token creation and run as shell command, e.g.:

    ```bash
    bentoml yatai login --api-token {YOUR_TOKEN} --endpoint http://127.0.0.1:8080
    ```

If you don't already have a Bento built, run the following commands from the [BentoML Quickstart Project](https://github.com/bentoml/BentoML/tree/main/examples/quickstart) to build a sample Bento:

```bash
git clone https://github.com/bentoml/bentoml.git && cd ./examples/quickstart
pip install -r ./requirements.txt
python train.py
bentoml build
```

Push your newly built Bento to Yatai:

```bash
bentoml push iris_classifier:latest
```

#### Install yatai-deployment componet

Yatai's Deployment feature comes as a separate component, you can install it via the following
script:

```bash
DEVEL=true bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai-deployment/main/scripts/quick-install-yatai-deployment.sh")
```

Similiarly, this script is made for development and testing purpose only.

#### Deploy Bento!

Once the `yatai-deployment` component was installed, you can new deploy Bentos to your Kubernetes
cluster via Yatai. A Bento Deployment can be created via Web UI or via kubectl command.

* Deploy via Web UI

  * Go to deployments page: http://127.0.0.1:8080/deployments
  * Click `Create` button and follow instructions on UI

* Deploy directly via `kubectl` command:

Define your Bento deployment in a `my_deployment.yaml` file:
```yaml
apiVersion: serving.yatai.ai/v1alpha2
kind: BentoDeployment
metadata:
    name: my-bento-deployment
    namespace: yatai
spec:
    bento_tag: iris_classifier:3oevmqfvnkvwvuqj
    ingress:
        enabled: true
    resources:
        limits:
            cpu: "500m"
            memory: "512m"
        requests:
            cpu: "250m"
            memory: "128m"
    autoscaling:
        max_replicas: 10
        min_replicas: 2
    runners:
        - name: iris_clf
          resources:
              limits:
                  cpu: "1000m"
                  memory: "1Gi"
              requests:
                  cpu: "500m"
                  memory: "512m"
              autoscaling:
                  max_replicas: 4
                  min_replicas: 1
```

Apply the deployment to your minikube cluster:
```bash
kubectl apply -f my_deployment.yaml
```

Now you can see the deployment process from the Yatai Web UI and find the endpoint URL for accessing
the deployed Bento.



## Community

-   To report a bug or suggest a feature request, use [GitHub Issues](https://github.com/bentoml/yatai/issues/new/choose).
-   For other discussions, use [GitHub Discussions](https://github.com/bentoml/BentoML/discussions) under the [BentoML repo](https://github.com/bentoml/BentoML/)
-   To receive release announcements and get support, join us on [Slack](https://join.slack.bentoml.org).

## Contributing

There are many ways to contribute to the project:

-   If you have any feedback on the project, share it with the community in [GitHub Discussions](https://github.com/bentoml/BentoML/discussions) under the [BentoML repo](https://github.com/bentoml/BentoML/).
-   Report issues you're facing and "Thumbs up" on issues and feature requests that are relevant to you.
-   Investigate bugs and reviewing other developer's pull requests.
-   Contributing code or documentation to the project by submitting a GitHub pull request. See the [development guide](https://github.com/bentoml/yatai/blob/main/DEVELOPMENT.md).

## Licence

[Elastic License 2.0 (ELv2)](https://github.com/bentoml/yatai/blob/main/LICENSE.md)

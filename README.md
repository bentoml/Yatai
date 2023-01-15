# 🦄️ Yatai: Model Deployment at Scale on Kubernetes

[![actions_status](https://github.com/bentoml/yatai/workflows/CICD/badge.svg)](https://github.com/bentoml/yatai/actions)
[![docs](https://readthedocs.org/projects/yatai/badge/?version=latest&style=flat-square)](https://docs.bentoml.org/projects/yatai)
[![join_slack](https://badgen.net/badge/Join/Community%20Slack/cyan?icon=slack&style=flat-square)](https://join.slack.bentoml.org)

Yatai (屋台, food cart) lets you deploy, operate and scale Machine Learning services on Kubernetes. 

It supports deploying any ML models via [BentoML: the unified model serving framework](https://github.com/bentoml/bentoml).

<img width="785" alt="yatai-overview-page" src="https://user-images.githubusercontent.com/489344/151455964-4fe30eb7-f000-43cc-8a5f-807ee450b8b6.png">


👉 [Join our Slack community today!](https://l.bentoml.com/join-slack)

✨ Looking for the fastest way to give Yatai a try? Check out [BentoML Cloud](https://www.bentoml.com/bentoml-cloud/) to get started today.


---

## Why Yatai?

🍱 Made for BentoML, deploy at scale

- Scale [BentoML](https://github.com/bentoml) to its full potential on a distributed system, optimized for cost saving and performance.
- Manage deployment lifecycle to deploy, update, or rollback via API or Web UI.
- Centralized registry providing the **foundation for CI/CD** via artifact management APIs, labeling, and WebHooks for custom integration.

🚅 Cloud native & DevOps friendly

- Kubernetes-native workflow via [BentoDeployment CRD](https://docs.bentoml.org/projects/yatai/en/latest/concepts/bentodeployment_crd.html) (Custom Resource Definition), which can easily fit into an existing GitOps workflow.
- Native [integration with Grafana](https://docs.bentoml.org/projects/yatai/en/latest/observability/metrics.html) stack for observability.
- Support for traffic control with Istio.
- Compatible with all major cloud platforms (AWS, Azure, and GCP).


## Getting Started

- 📖 [Documentation](https://docs.bentoml.org/projects/yatai/) - Overview of the Yatai docs and related resources
- ⚙️ [Installation](https://docs.bentoml.org/projects/yatai/en/latest/installation/index.html) - Hands-on instruction on how to install Yatai for production use
- 👉 [Join Community Slack](https://l.linklyhq.com/l/ktPW) - Get help from our community and maintainers


## Quick Tour

Let's try out Yatai locally in a minikube cluster!

### ⚙️ Prerequisites:
  * Install latest minikube: https://minikube.sigs.k8s.io/docs/start/
  * Install latest Helm: https://helm.sh/docs/intro/install/
  * Start a minikube Kubernetes cluster: `minikube start --cpus 4 --memory 4096`, if you are using macOS, you should use [hyperkit](https://minikube.sigs.k8s.io/docs/drivers/hyperkit/) driver to prevent the macOS docker desktop [network limitation](https://docs.docker.com/desktop/networking/#i-cannot-ping-my-containers)
  * Check that minikube cluster status is "running": `minikube status`
  * Make sure your `kubectl` is configured with `minikube` context: `kubectl config current-context`
  * Enable ingress controller: `minikube addons enable ingress`

### 🚧 Install Yatai

Install Yatai with the following script:

```bash
bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/main/scripts/quick-install-yatai.sh")
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
YATAI_INITIALIZATION_TOKEN=$(kubectl get secret yatai-env --namespace yatai-system -o jsonpath="{.data.YATAI_INITIALIZATION_TOKEN}" | base64 --decode)
echo "Open in browser: http://127.0.0.1:8080/setup?token=$YATAI_INITIALIZATION_TOKEN"
``` 

Open the URL printed above from your browser to finish admin account setup.


### 🍱 Push Bento to Yatai

First, get an API token and login to the BentoML CLI:

* Keep the `kubectl port-forward` command in the step above running
* Go to Yatai's API tokens page: http://127.0.0.1:8080/api_tokens
* Create a new API token from the UI, making sure to assign "API" access under "Scopes"
* Copy the login command upon token creation and run as a shell command, e.g.:

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

Now you can view and manage models and bentos from the web UI:

<img width="785" alt="yatai-bento-repos" src="https://user-images.githubusercontent.com/489344/151456379-da255519-274d-41de-a1b9-a347be279230.png">
<img width="785" alt="yatai-model-detail" src="https://user-images.githubusercontent.com/489344/151456021-360a6d6e-acb8-494b-9f6b-868ef9d13bce.png">

### 🔧 Install yatai-image-builder component

Yatai's image builder feature comes as a separate component, you can install it via the following
script:

```bash
bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai-image-builder/main/scripts/quick-install-yatai-image-builder.sh")
```

This will install the `BentoRequest` CRD([Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)) and `Bento` CRD
in your cluster. Similarly, this script is made for development and testing purposes only.

### 🔧 Install yatai-deployment component

Yatai's Deployment feature comes as a separate component, you can install it via the following
script:

```bash
bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai-deployment/main/scripts/quick-install-yatai-deployment.sh")
```

This will install the `BentoDeployment` CRD([Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/))
in your cluster and enable the deployment UI on Yatai. Similarly, this script is made for development and testing purposes only.

### 🚢 Deploy Bento!

Once the `yatai-deployment` component was installed, Bentos pushed to Yatai can be deployed to your 
Kubernetes cluster and exposed via a Service endpoint. 

A Bento Deployment can be created either via Web UI or via a Kubernetes CRD config:

#### Option 1. Simple Deployment via Web UI

* Go to the deployments page: http://127.0.0.1:8080/deployments
* Click `Create` button and follow the instructions on the UI

<img width="785" alt="yatai-deployment-creation" src="https://user-images.githubusercontent.com/489344/151456002-d4e9f84d-8a71-4bf9-bde7-f94a74abbf3f.png">

#### Option 2. Deploy with kubectl & CRD

Define your Bento deployment in a `my_deployment.yaml` file:

```yaml
apiVersion: resources.yatai.ai/v1alpha1
kind: BentoRequest
metadata:
    name: iris-classifier
    namespace: yatai
spec:
    bentoTag: iris_classifier:3oevmqfvnkvwvuqj
---
apiVersion: serving.yatai.ai/v2alpha1
kind: BentoDeployment
metadata:
    name: my-bento-deployment
    namespace: yatai
spec:
    bento: iris-classifier
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
        maxReplicas: 10
        minReplicas: 2
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
                  maxReplicas: 4
                  minReplicas: 1
```

Apply the deployment to your minikube cluster:
```bash
kubectl apply -f my_deployment.yaml
```

Now you can see the deployment process from the Yatai Web UI and find the endpoint URL for accessing
the deployed Bento.

<img width="785" alt="yatai-deployment-details" src="https://user-images.githubusercontent.com/489344/151456024-151c275d-b33e-480e-be34-dadab5b01915.png">




## Community

-   To report a bug or suggest a feature request, use [GitHub Issues](https://github.com/bentoml/yatai/issues/new/choose).
-   For other discussions, use [GitHub Discussions](https://github.com/bentoml/BentoML/discussions) under the [BentoML repo](https://github.com/bentoml/BentoML/)
-   To receive release announcements and get support, join us on [Slack](https://join.slack.bentoml.org).

## Contributing

There are many ways to contribute to the project:

-   If you have any feedback on the project, share it with the community in [GitHub Discussions](https://github.com/bentoml/BentoML/discussions) under the [BentoML repo](https://github.com/bentoml/BentoML/).
-   Report issues you're facing and "Thumbs up" on issues and feature requests that are relevant to you.
-   Investigate bugs and review other developers' pull requests.
-   Contributing code or documentation to the project by submitting a GitHub pull request. See the [development guide](https://github.com/bentoml/yatai/blob/main/DEVELOPMENT.md).

### Usage Reporting

Yatai collects usage data that helps our team to improve the product.
Only Yatai's internal API calls are being reported. We strip out as much potentially
sensitive information as possible, and we will never collect user code, model data, model names, or stack traces.
Here's the [code](./api-server/services/tracking/) for usage tracking.
You can opt-out of usage by configuring the helm chart option `doNotTrack` to
`true`.

```yaml
doNotTrack: false
```

Or by setting the `YATAI_DONOT_TRACK` env var in yatai deployment.
```yaml
spec:
  template:
    spec:
      containers:
        env:
        - name: YATAI_DONOT_TRACK
          value: "true"
```


## Licence

[Elastic License 2.0 (ELv2)](https://github.com/bentoml/yatai/blob/main/LICENSE.md)

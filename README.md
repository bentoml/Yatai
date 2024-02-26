# 🦄️ Yatai: Model Deployment at Scale on Kubernetes

[![actions_status](https://github.com/bentoml/yatai/workflows/Release/badge.svg)](https://github.com/bentoml/yatai/actions)
[![join_slack](https://badgen.net/badge/Join/Community%20Slack/cyan?icon=slack&style=flat-square)](https://join.slack.bentoml.org)

⚠️ Yatai for [BentoML 1.2](https://github.com/bentoml/BentoML/releases/tag/v1.2.0) is currently under construction. See [Yatai 2.0 Proposal](https://github.com/bentoml/Yatai/issues/504) for more details. 

---

Yatai (屋台, food cart) is the Kubernetes deployment operator for [BentoML](https://github.com/bentoml/bentoml).

It let DevOps teams to seamlessly integrate BentoML into their GitOps workflow, for deploying and scaling Machine Learning services on any Kubernetes cluster.

👉 [Join our Slack community today!](https://l.bentoml.com/join-slack)

---

## Why Yatai?

Yatai empowers developers to deploy [BentoML](https://github.com/bentoml) on Kubernetes, optimized for CI/CD and DevOps workflow.

Yatai is Cloud native and DevOps friendly. Via its Kubernetes-native workflow, specifically the [BentoDeployment CRD](https://docs.yatai.io/en/latest/concepts/bentodeployment_crd.html) (Custom Resource Definition), DevOps teams can easily fit BentoML powered services into their existing workflow.


## Getting Started

- 📖 [Documentation](https://docs.yatai.io/) - Overview of the Yatai docs and related resources
- ⚙️ [Installation](https://docs.yatai.io/en/latest/installation/index.html) - Hands-on instruction on how to install Yatai for production use
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
For production deployment, check out the [Installation Guide](https://docs.yatai.io/en/latest/installation/index.html).

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

A Bento Deployment can be created via applying a BentoDeployment resource:

Define your Bento deployment in a `my_deployment.yaml` file:

```yaml
apiVersion: resources.yatai.ai/v1alpha1
kind: BentoRequest
metadata:
    name: iris-classifier
    namespace: yatai
spec:
    bentoTag: iris_classifier:3oevmqfvnkvwvuqj  # check the tag by `bentoml list iris_classifier`
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
            memory: "512Mi"
        requests:
            cpu: "250m"
            memory: "128Mi"
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
                  memory: "512Mi"
          autoscaling:
              maxReplicas: 4
              minReplicas: 1
```

Apply the deployment to your minikube cluster:
```bash
kubectl apply -f my_deployment.yaml
```

Now you can check the deployment status via `kubectl get BentoDeployment -n my-bento-deployment`



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




## Licence

[Elastic License 2.0 (ELv2)](https://github.com/bentoml/yatai/blob/main/LICENSE.md)

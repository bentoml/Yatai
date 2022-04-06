# 🦄️ Yatai: Model Deployment at scale on Kubernetes

[![actions_status](https://github.com/bentoml/yatai/workflows/CICD/badge.svg)](https://github.com/bentoml/yatai/actions)
[![join_slack](https://badgen.net/badge/Join/BentoML%20Slack/cyan?icon=slack)](https://join.slack.bentoml.org)

Yatai helps ML teams to deploy large scale model serving workloads on Kubernetes. It standarlizes [BentoML](https://github.com/bentoml) deployment on Kubernetes, provides UI for managing all your ML models and deployments in one place, and enables advanced GitOps and CI/CD workflow.

👉 [Pop into our Slack community!](https://l.linklyhq.com/l/ktPW) We're happy to help with any issue you face or even just to meet you and hear what you're working on :)


Core features:

* **Deployment Automation** - deploy Bentos as auto-scaling API endpoints on Kubernetes and easily rollout new versions
* **Bento Registry** - manage all your team's Bentos and Models, backed by cloud blob storage (S3, MinIO)
* **Observability** - monitoring dashboard helping users to identify model performance issues
* **CI/CD** - flexible APIs for integrating with your training and CI pipelines


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



## Why Yatai

* Yatai is built upon [BentoML](https://github.com/bentoml/BentoML), the unified model serving framework that is high-performing and feature-rich
* Yatai focus on the model serving and deployment part of your MLOps stack, works well with any ML training/monitoring platforms, such as AWS SageMaker or MLFlow
* Yatai is Kubernetes native, integrates well with other cloud native tools in the K8s eco-system
* Yatai is human-centric, provides easy-to-use Web UI and APIs for ML scientists, MLOps engineers, and project managers


## Getting Started

<details>
  <summary>1. Install Yatai locally with Minikube</summary>
    
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
   * Wait for installation to complete, this may take a few minutes to complete: `helm status yatai -n yatai-system`
   * Start minikube tunnel for accessing Yatai UI: `sudo minikube tunnel`
   * Get initialization link for creating your admin account:
      ```bash 
      export YATAI_INITIALIZATION_TOKEN=$(kubectl get secret yatai --namespace yatai-system -o jsonpath="{.data.initialization_token}" | base64 --decode)
      echo "Visit: http://yatai.127.0.0.1.sslip.io/setup?token=$YATAI_INITIALIZATION_TOKEN"
      ```
</details>

    
<details>
  <summary>2. Get an API token and login BentoML CLI</summary>
    
  * Create a new API token in Yatai web UI: http://yatai.127.0.0.1.sslip.io/api_tokens
  * Copy login command upon token creation and run as shell command, e.g.: 
    ```bash
    bentoml yatai login --api-token {YOUR_TOKEN_GOES_HERE} --endpoint http://yatai.127.0.0.1.sslip.io
    ```
</details>

<details>
  <summary>3. Pushing Bento to Yatai</summary>
    
  * Train a sample ML model and build a Bento using code from the [BentoML Quickstart Project](https://github.com/bentoml/gallery/tree/main/quickstart):
    ```bash
    git clone https://github.com/bentoml/gallery.git && cd ./gallery/quickstart
    pip install -r ./requirements.txt
    python train.py
    bentoml build
    ```
  * Push your newly built Bento to Yatai:
    ```bash
    bentoml push iris_classifier:latest
    ```
</details>

    
<details>
  <summary>4. Create your first deployment!</summary>
    
  * A Bento Deployment can be created via Web UI or via kubectl command:

    * Deploy via Web UI
        * Go to deployments page: http://yatai.127.0.0.1.sslip.io/deployments
        * Click `Create` button and follow instructions on UI

    * Deploy directly via `kubectl` command:
        * Define your Bento deployment in a YAML file:
          ```yaml
          # my_deployment.yaml
          apiVersion: serving.yatai.ai/v1alpha1
          kind: BentoDeployment
          metadata:
            name: demo
          spec:
            bento_tag: iris_classifier:3oevmqfvnkvwvuqj
            resources:
              limits:
                cpu: 1000m
              requests:
                cpu: 500m
          ```
        * Apply the deployment to your minikube cluster
          ```bash
          kubeclt apply -f my_deployment.yaml
          ```

  * Monitor deployment process on Web UI and test out endpoint when deployment created
    ```bash
    curl \                                                                                                                                                      
        -X POST \
        -H "content-type: application/json" \
        --data "[5, 4, 3, 2]" \
        https://demo-default-yatai-127-0-0-1.apps.yatai.dev/classify
    ```
</details>
    
<details>
  <summary>5. Moving to production</summary>
    
  * See [Administrator's Guide](https://github.com/bentoml/yatai/blob/main/docs/admin-guide.md) for a comprehensive overview for deploying and configuring Yatai for production use.
</details>


## Community

- To report a bug or suggest a feature request, use [GitHub Issues](https://github.com/bentoml/yatai/issues/new/choose).
- For other discussions, use [GitHub Discussions](https://github.com/bentoml/BentoML/discussions) under the [BentoML repo](https://github.com/bentoml/BentoML/)
- To receive release announcements and get support, join us on [Slack](https://join.slack.bentoml.org).


## Contributing

There are many ways to contribute to the project:

- If you have any feedback on the project, share it with the community in [GitHub Discussions](https://github.com/bentoml/BentoML/discussions) under the [BentoML repo](https://github.com/bentoml/BentoML/).
- Report issues you're facing and "Thumbs up" on issues and feature requests that are relevant to you.
- Investigate bugs and reviewing other developer's pull requests.
- Contributing code or documentation to the project by submitting a GitHub pull request. See the [development guide](https://github.com/bentoml/yatai/blob/main/DEVELOPMENT.md).


## Licence

[Elastic License 2.0 (ELv2)](https://github.com/bentoml/yatai/blob/main/LICENSE.md)

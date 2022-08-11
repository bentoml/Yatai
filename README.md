# 🦄️ Yatai: Production-first ML platform on Kubernetes

[![actions_status](https://github.com/bentoml/yatai/workflows/CICD/badge.svg)](https://github.com/bentoml/yatai/actions)
[![join_slack](https://badgen.net/badge/Join/BentoML%20Slack/cyan?icon=slack)](https://join.slack.bentoml.org)

Yatai is a production-first platform for your machine learning needs. It brings collaborative [BentoML](https://github.com/bentoml) workflows to Kubernetes, helps ML teams to run model serving at scale, while simplifying model management and deployment across teams. 

👉 [Pop into our Slack community!](https://l.linklyhq.com/l/ktPW) We're happy to help with any issue you face or even just to meet you and hear what you're working on :)

## Why Yatai?

* Yatai accelerates the process of taking ML models from training stage to production and reduces the operational overhead of running a reliable model serving system.

* Yatai simplifies collaboration between Data Science and Engineering teams. It is designed to leverage the BentoML standard and streamline production ML workflows.

* Yatai is a cloud native platform with a wide range of integrations to best fit your infrastructure needs, and it is easily customizable for your CI/CD needs.


## Core features:

* **Bento Registry** - manage all your team's ML models via simple Web UI and API, and store ML assets on cloud blob storage
* **Deployment Automation** - deploy Bentos as auto-scaling API endpoints on Kubernetes and easily rollout new versions
* **Observability** - monitoring dashboard and logging integration helping users to identify model performance issues
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


## Getting Started

<details>
  <summary>1. Install Yatai locally with Minikube</summary>
    
   * Prerequisites:
     * Install latest minikube: https://minikube.sigs.k8s.io/docs/start/
     * Install latest Helm: https://helm.sh/docs/intro/install/
   * Start a minikube Kubernetes cluster: `minikube start --cpus 4 --memory 4096`
   * Enable ingress controller: `minikube addons enable ingress`
   * Install Yatai Helm Chart:
     ```bash
     helm repo add yatai https://bentoml.github.io/yatai-chart
     helm repo update
     helm install yatai yatai/yatai -n yatai-system --create-namespace
     ```
   * [Verify installation](./docs/admin-guide.md#verify-installation)
   * You can access the Yatai Web UI: http://{Yatai URL}/setup?token=<token>. You can find the **Yatai URL** link and the **token** again using `helm get notes yatai -n yatai-system` command.
</details>

    
<details>
  <summary>2. Get an API token and login BentoML CLI</summary>
    
  * Create a new API token in Yatai web UI: http://${Yatai URL}/api_tokens
  * Copy login command upon token creation and run as shell command, e.g.: 
    ```bash
    bentoml yatai login --api-token {YOUR_TOKEN_GOES_HERE} --endpoint http://{Yatai URL}
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
        * Go to deployments page: http://{Yatai URL}/deployments
        * Click `Create` button and follow instructions on UI

    * Deploy directly via `kubectl` command:
        * Define your Bento deployment in a `my_deployment.yaml` file:
          ```yaml
            apiVersion: serving.yatai.ai/v1alpha2
            kind: BentoDeployment
            metadata:
              name: my-bento-deployment
              namespace: my-namespace
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
        * Apply the deployment to your minikube cluster
          ```bash
          kubectl apply -f my_deployment.yaml
          ```

  * Monitor deployment process on Web UI and test out endpoint when deployment created
    ```bash
    curl \                                                                                                                                                      
        -X POST \
        -H "content-type: application/json" \
        --data "[[5, 4, 3, 2]]" \
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

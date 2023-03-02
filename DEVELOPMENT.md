# Developer Guide

I'm glad you can see this document and I'm looking forward to your contributions to the Yatai.

Yatai does not rely on cloud-native, but it is accessible to cloud-native based yatai-deployment as a RESTful api-server, so how to bridge the network between the Kubernetes cluster and the local development environment is a problem that needs to be solved

As you know, Kubernetes has a complex network environment, so developing cloud-native related products locally can be a challenge. But don't worry, this document will show you how to develop Yatai locally easily, quickly and comfortably.

## Prerequisites

- [Git LFS](https://git-lfs.github.com/)
  
  Yatai Git project use LFS to storage large image files in the documentation.
  

- A Yatai installed in the **development environment** for development and debugging

    > NOTE: Since you are developing, **you must not use the production environment**, so we recommend using the quick install script to install Yatai in the local minikube

    Using a pre-installed Yatai in the development environment, the aim is to quickly provide a range of out-of-the-box infrastructure dependencies

    You can start by reading this [installation document](https://docs.bentoml.org/projects/yatai/en/latest/installation/yatai.html) to install Yatai. It is highly recommended to use the [quick install script](https://docs.bentoml.org/projects/yatai/en/latest/installation/yatai.html#quick-install) to install Yatai

    Remember, **never use infrastructure from the production environment**, only use newly installed infrastructure in the cluster, such as SQL databases, blob storage, docker registry, etc. The [quick install script](https://docs.bentoml.org/projects/yatai/en/latest/installation/yatai.html#quick-install) mentioned above will prevent you from using the infrastructure in the production environment, this script will help you to install all the infrastructure from scratch, you can use it without any worries.

    If you have already installed it, please verify that your kubectl context is correct with the following command:

    ```bash
    kubectl config current-context
    ```

- [jq](https://stedolan.github.io/jq/)

    Used to parse json from the command line

- [Go language compiler](https://go.dev/)

    Yatai api-server is implemented by Go Programming Language

- [Node.js](https://nodejs.org/en/)

    Yatai Web UI is implemented by TypeScript + React

    * We recommend installing NodeJS using `nvm` which allows developers to quickly install and use different versions of node:

    ```bash
    # Install NVM
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.38.0/install.sh | bash

    nvm install 14.17.1
    nvm alias default 14.17.1
    ```

- [Yarn Package Manager](https://yarnpkg.com/)

    Yatai Web UI uses `yarn` to manage dependencies.

    ```bash
    npm install -g yarn
    ```

- [Telepresence](https://www.telepresence.io/)

    The most critical dependency in this document for bridging the local network and the Kubernetes cluster network

## Start Developing

<details>

1. Fork the Yatai project on [GitHub](https://github.com/bentoml/Yatai)

2. Clone the source code from your fork of Yatai's GitHub repository:

    ```bash
    git clone git@github.com:${your github username}/Yatai.git && cd Yatai
    ```

3. Add the Yatai upstream remote to your local Yatai clone:

    ```bash
    git remote add upstream git@github.com:bentoml/Yatai.git
    ```

4. Installing Go dependencies

    ```bash
    go mod download
    ```
</details>

## Making Changes

<details>
1. Make sure you're on the main branch.

   ```bash
   git checkout main
   ```

2. Use the git pull command to retrieve content from the BentoML GitHub repository.

   ```bash
   git pull upstream main -r
   ```

3. Create a new branch and switch to it.

   ```bash
   git checkout -b your-new-branch-name
   ```

4. Make your changes!

5. Use the git add command to save the state of files you have changed.

   ```bash
   git add <names of the files you have changed>
   ```

6. Commit your changes.

   ```bash
   git commit -m 'your commit message'
   ```

7. Synchronize upstream changes

    ```bash
    git pull upstream main -r
    ```

8. Push all changes to your forked repo on GitHub.

   ```bash
   git push origin your-new-branch-name
   ```
</details>

## Run Yatai api-server

1. Connect to the Kubernetes cluster network

    ```bash
    telepresence connect
    ```

2. Run Yatai api-server

    > NOTE: The following command uses the infrastructure of the Kubernetes environment in the current kubectl context and replaces the behavior of Yatai in the current Kubernetes environment, so please proceed with caution

    ```bash
    make start-dev
    ```

3. ✨ Enjoy it!

## Run Yatai Web UI

1. Install dependencies

    ```bash
    cd dashboard
    yarn
    cd -
    ```

2. Run frontend proxy server

    ```bash
    cd dashboard
    yarn start
    ```

3. ✨ Enjoy it!

<div align="center">
    <h1 align="center">Yatai</h1>
    <p><strong>BentoML's model management server<br></strong></p>
    <p align="center">
        <a href="https://github.com/bentoml/yatai/actions/workflows/cicd.yml">
            <img alt="Build Status" src="https://github.com/bentoml/yatai/actions/workflows/cicd.yml/badge.svg">
        </a>
    </p>
</div>


# Running Yatai on minikube

### Prerequisites

  - Minikube. Visit [minikube.io](https://minikube.io/) to install and manage minikube.
  - Helm. Visit [helm.sh](https://helm.sh/) to install and manage helm.


**Start minkube cluster**

```bash
minikube start --memory 8192 --cpus 8
```


**1. Install and update helm repo**

```bash
helm repo add yatai https://bentoml.github.io/yatai-chart
```

```bash
helm repo update
```

**2. Install Yatai**

```bash
helm install yatai yatai/yatai -n yatai-system --create-namespace
```

**3. Access Yatai**

Create minikube tunnel:
```bash
sudo minikube tunnel
```
Visit [Yatai](http://yatai.127.0.0.1.sslip.io/) on the browser. To login, use the following credentials:
username: `admin`, password: `admin`


**Delete Yatai**

```bash
helm uninstall yatai
kubectl delete ns yatai-components
kubectl delete ns yatai-operators
```



# Development guide

__TLDR__: Run `make yatai-dev` to quickly spin up yatai in development mode. Make sure to create `yatai-config.dev.yaml` that follows [./yatai-config.sample.yaml](./yatai-config.sample.yaml) templates.

## Backend

### Install go and postgresql, create yatai db and create an default user if none exists.

```bash
brew install go
brew install postgresql

createdb yatai
```

### Install dependencies
```bash
make be-deps
```

### Setup configs and run server all at once

```bash
make be-run
```

Visit backend swagger endpoints via [`localhost:7777/swagger`](http://localhost:7777/swagger)

## Frontend

__NOTES__: Make sure to create [GitHub OAuth](https://docs.github.com/en/developers/apps/building-oauth-apps/creating-an-oauth-app) and edit `yatai-config.{dev,test,production}.yaml`

### Install nvm and nodejs and yarn

```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.38.0/install.sh | bash

# install nodejs with nvm
nvm install 14.16.1
nvm alias default 14.16.1

# for Mac M1 chips(arm) only
nvm install 14.17.1
nvm alias default 14.17.1


# then install yarn
npm install -g yarn
```

### Install the dependencies

```bash
make fe-deps
```

### Run front-end development server

```bash
make fe-run
```

Visit React App via [`localhost:3000`](http://localhost:3000). You can also accessed swagger via [`localhost:3000/swagger`](http://localhost:3000/swagger)

## Docker

```bash
# Build docker images
make yatai-d

# Run docker images
make yatai-d-r
```

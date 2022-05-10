# Yatai development guide

Yatai uses Golang for its backend and react/typescript for the frontend web UI. Download the source code:

```bash
git clone https://github.com/bentoml/yatai.git
```

# Prerequisites

You can do this the [hard way](#conventional-way) or the [easy way](#nix)

## Conventional way

### Yatai Web UI

1. NodeJS version 14.16.1 or above

    > For Apple computer with M1 chip, please install nodejs version `>=14.17.1`
    >
    - We recommend installing NodeJS using `nvm` which allows developers to quickly install and use different versions of node:

        ```bash
        # Install NVM
        curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.38.0/install.sh | bash

        nvm install 14.17.1
        nvm alias default 14.17.1
        ```

2. Yarn package manager (optional)

    Yatai uses `yarn` package manager to run varies of Web UI related ops in the Makefile.

    ```bash
    npm install -g yarn
    ```


### Yatai Server

1. PostgreSQL

    A local PostgreSQL database is required to set up a local development environment. Follow the official installation guide for your system here: [https://www.postgresql.org/download/](https://www.postgresql.org/download/)

    For mac users, install Postgres with homebrew:

    ```bash
    brew install postgresql
    ```

    After installation, create a database for Yatai

    ```bash
    createdb yatai
    ```

2. Golang

    Yatai uses golang for its backend. Install Golang for your system following the official installation guide here: [https://go.dev/doc/install](https://go.dev/doc/install)

    For Mac users, install Golang with homebrew:

    ```bash
    brew install go
    ```


### Install dependencies

#### Yatai WebUI

Yatai uses yarn to manage its front-end dependencies.  Run the make command:

```bash
make fe-deps
```

Alternatively navigate to the `dashboard` directory and run `yarn` command:

```bash
cd dasboard
yarn
```

#### Yatai server

Yatai uses go command to download the dependency packages.  Run the make command:

```bash
make be-deps
```

Alternatively to run the download command directly:

```bash
go mod download
```

## Nix

Using [nix](https://nixos.org/download.html) will enable a reproducable workflow
and faster development setup.

We are using [niv](https://github.com/nmattia/niv) to manage your dependencies.

Install `nix`:
```shell
sh <(curl -L https://nixos.org/nix/install) --daemon
```

If you are on MacOS, then do:
```bash
sh <(curl -L https://nixos.org/nix/install) --darwin-use-unencrypted-nix-store-volume --daemon
```
and follow the instruction from the installer. Then reboot your system.

Enter the development environment without any package leakage via `nix-shell`.
```bash
nix-shell --pure
```
`nix-shell` enables users to start an interactive shell based on a Nix
expression [`shell.nix`](./shell.nix) without any local
environment variables, more [here](https://nixos.org/manual/nix/stable/command-ref/nix-shell.html).

In order for `minikube` to recognize the `PostgreSQL` database managed via `nix`, open another terminal windows and then run:
```bash
nix-shell
```
This will run `nix-shell` in `impure` mode. This means that `nix-shell` won't
clear your environment variable before starting the interactive shell. Then
proceed to step 3 and 4 in [Run development server](#run-development-server)

## Run development server

1. Generate Yatai config file

    Create `yatai-config.dev.yaml` file that bases on the `yatai-config.sample.yaml` template and update the `postgrsql` section in the configuration file.

    An example for a local `yatai-config.dev.yaml` (nix-shell compatible):

    ```yaml
        in_cluster: false

        server:
        enable_https: false
        port: 7777
        session_secret_key:
        migration_dir: ./api-server/db/migrations

        postgresql:
        host: localhost
        port: 5432
        user: postgres
        password: ''
        database: yatai

        initialization_token: 12345
    ```

2. Spin up `minikube`:
```bash
minikube start --cpus 4 --memory 4096
```

3. Run `sudo minikube tunnel` to enable ingress controller.

4. Run make command that start the development server for both Yatai UI and Yatai server.

    ```bash
    make yatai-dev
    ```

    Visit http://localhost:7777 to view the Yatai Web UI

    Visit http://localhost:7777/swagger to view Yatai server’s API definitions.

    Visit http://localhost:3000/setup?token=12345 to initially setup a dev
    credentials.

5. In order to push bentos locally to the development Yatai:

    1. Clone [`yatai-deployment-operator`](https://github.com/bentoml/yatai-deployment-operator):

    ```bash
    git clone https://github.com/bentoml/yatai-deployment-operator && cd
    yatai-deployment-operator
    ```

    2. `yatai-deployment-operator` is automatically setup when you run
       `yatai-api-server`, thus we need to scale the replica to 0 from
       `kubectl`:

    ```bash
    kubectl -n yatai-components scale deployment/yatai-yatai-deployment-operator --replicas=0
    ```

    3. One can retrieve `YATAI_API_TOKEN` from the yatai dashboard, it should be
       the same `API_TOKEN` that is used for `bentoml yatai login`

    4. Run the following command:
    ```bash
    YATAI_ENDPOINT=http://localhost:3000 YATAI_API_TOKEN=<API_TOKEN_HERE> YATAI_CLUSTER_NAME=default ENABLE_WEBHOOKS=false make run
    ```

    5. You should then be able to push bentos with `bentoml push`

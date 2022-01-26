# Yatai development guide

Yatai uses Golang for its backend and react/typescript for the frontend web UI. Download the source code:

```bash
git clone https://github.com/bentoml/yatai.git
```

## Prerequisites

### Yatai Web UI

1. NodeJS version 14.16.1 or above

    > For Apple computer with M1 chip, please install nodejs version `>=14.17.1`
    >
    - We recommend installing NodeJS using `nvm`which allows developers to quickly install and use different versions of node:

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


## Install dependencies

### Yatai WebUI

Yatai uses yarn to manage its front-end dependencies.  Run the make command:

```bash
make fe-deps
```

Alternatively navigate to the `dashboard` directory and run `yarn` command:

```bash
cd dasboard
yarn
```

### Yatai server

Yatai uses go command to download the dependency packages.  Run the make command:

```bash
make be-deps
```

Alternatively to run the download command directly:

```bash
go mod download
```

## Run development server

1. Generate Yatai config file

    Create `yatai-config.dev.yaml` file that bases on the `yatai-config.sample.yaml` template and update the `postgrsql` section in the configuration file.


1. Run make command that start the development server for both Yatai UI and Yatai server.

    ```bash
    make yatai-dev
    ```

    Visit http://localhost:7777 to view the Yatai Web UI

    Visit http://localhost:7777/swagger to view Yatai server’s API definitions.


To start Yatai UI separately, run make command:

```bash
make fe-run
```

To start Yatai server separately, run make command:

```bash
make be-run
```

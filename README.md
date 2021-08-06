yatai
-------

# Development Guide

## Back-end Part

### 1. Install go and postgresql and create yatai database on you OS

```bash
brew install go
brew install postgresql

createdb yatai
```

### 2. Go to the project root directory

### 3. Copy the sample config file and modify it

```bash
cp ./yatai-config.sample.yaml ./yatai-config.dev.yaml
```

### 4. Install dependencies

```bash
go mod download
```

### 5. Run your server

```bash
go run ./api-server/main.go serve -d -c ./yatai-config.dev.yaml
```

Now you can visit the swagger on http://localhost:7777/swagger

## Front-end part

### 1. Install nvm and nodejs and yarn

#### 1.1 Install nvm

```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.38.0/install.sh | bash
```

#### 1.2 Install nodejs

```bash
nvm install 14.16.1
nvm alias default 14.16.1
```

#### 1.3 Install yarn

```bash
npm install -g yarn
```

### 1. Go to the ui directory

```bash
cd ui
```

### 2. Install the dependencies

```bash
yarn
```

### 3. Run front-end development server

```bash
yarn start
```

Now you also can visit the swagger on http://localhost:3000/swagger

### 4. Login with GitHub OAuth

visit http://localhost:3000/oauth/github
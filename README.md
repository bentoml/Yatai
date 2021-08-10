yatai
-------

# Development Guide

## Back-end Part

### 1. Install go and postgresql, create yatai db and create an default user if none exists.

```bash
brew install go
brew install postgresql

createdb yatai
```

### 2. Install dependencies
```bash
make be-deps
```

### 3. Setup configs and run server all at once

```bash
make be-run
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

### 1. Install the dependencies

```bash
make fe-deps
```

### 2. Run front-end development server

```bash
make fe-run
```

Now you also can visit the swagger on http://localhost:3000/swagger

### 4. Login with GitHub OAuth

visit http://localhost:3000/oauth/github

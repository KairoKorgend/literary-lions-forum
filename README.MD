# Literary-lions forum

This forum is a web application that allows users to communicate, associate categories with posts, like/dislike posts & comments, and filter posts.

![](frontend/static/assets/forum_image.png)

## Installation

```bash
git clone https://github.com/KairoKorgend/literary-lions-forum.git
```

## Installing packages

Depending on your package manager (npm, yarn) run

```bash
npm install
```

or

```bash
yarn install
```

in projects root directory.

## Running the Application

Make sure you have the Docker installed on your system.
You have two options to run the application:

### Using Docker Compose

Use Docker Compose to build and run the application. In the root directory of the project, run the following command:

```bash
docker compose up --build
```

This command will build a Docker image and run a container using that image.

### Using the Docker Build Script

Use Docker Build Script to build and run the application. In the root directory of the project, run the following command:

```bash
./docker-build-run.sh
```

This script will build a Docker image and run a container using that image.

## Usage

Visit `http://localhost:3000`

## Running Tests

To run tests, use the following command in the root directory:

```bash
docker build -t forum-tests -f Dockerfile.test . -q > /dev/null && docker run --rm --name forum-tests forum-tests
```

This command will build a Docker image for testing and run the tests in a Docker container.

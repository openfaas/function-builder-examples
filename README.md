# OpenFaaS Pro Function Builder API examples
This repo contains some code examples that show you how to use the [OpenFaaSFunction Builder API](https://docs.openfaas.com/openfaas-pro/builder/) from different languages.

> The Function Builder API provides a simple REST API to create your functions from source code.
> See [Function Builder API docs](https://docs.openfaas.com/openfaas-pro/builder/) for installation instructions

Before attempting to run the examples make sure the pro-builder is port-forwarded to port 8081 on the local host.

```bash
kubectl port-forward \
    -n openfaas \
    svc/pro-builder 8081:8080
```

Save the HMAC signing secret created during the installation to a file `payload.txt` at the root of this repo.
```bash
kubectl get secret \
    -n openfaas payload-secret -o jsonpath='{.data.payload-secret}' \
    | base64 --decode \
    > payload.txt
```

The directory [hello-world](./hello-world/) can be passed as the handler directory to the examples. It contains a javascript handler for a function. The hello-world directory was created by running:

```bash
faas-cli new hello-world --lang node20
```

You can use the `faas-cli` to create any other handler to try these example scripts with.

## Use Python to call the pro-builder

The [python-request](./python-request/) directory has an example on how to invoke the Function Builder API from python. Run the `build.py` script with the required flags to turn a function handler into a container image.

The python script uses the [Requests](https://requests.readthedocs.io/en/latest/) package so you will have to install that to run the example.

```
sudo python3 -m pip install requests
```

Run the script
```bash
python3 python-request/build.py \
    --image ttl.sh/hello-world-python:1h \
    --handler ./hello-world \
    --lang node20
```

## Use NodeJS to call the pro-builder
The [nodejs-request](./nodejs-request/) directory has an example on how to invoke the Function Builder API from NodeJS. Run the `index.js` script with the required arguments to turn a function handler into a container image.

Install the required packages
```bash
cd nodejs-request
npm install
```

The script takes three arguments:
- The docker image name to build
- A directory with the handler for the function
- The language or template to use

```bash
node nodejs-request/index.js \
    'ttl.sh/hello-world-node:1h' \
    ./hello-world \
    node20
```

## Use php to call the pro-builder
The [php-request](./php-request/) directory has an example on how to invoke the Function Builder API from php. Run the `build.php` script with the required arguments to turn a function handler into a container image.

Install the required packages
```bash
cd php-request
php composer.phar install
```

The script takes three arguments:
- The docker image name to build
- A directory with the handler for the function
- The language or template to use

```bash
php php-request/build.php \
    --image=ttl.sh/hello-world-php:1h \
    --handler=./hello-world \
    --lang=node20
```

## Use go and the OpenFaaS go-sdk to call the pro-builder
The [go-request](./go-request/) directory has an example on how to invoke the Function Builder API from go using the official OpenFaaS go-sdk. Run the `main.go` script with the required flags to turn a function handler into a container image.

The example script does not fetch templates. How this is done is up to your implementation. Templates can be pulled from a git repository, copied from an S3 bucket, downloaded with an http call.

To try out this example you can fetch them using the faas-cli before running the script:
```sh
faas-cli template store pull node20
```

Run the script
```bash
go run go-request/main.go \
    -image=ttl.sh/hello-world-go:1h \
    -handler=./hello-world \
    -lang=node20 \
    -name="hello-world" \
    -platforms="linux/amd64"
```
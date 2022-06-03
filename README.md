# OpenFaaS Pro Function Builder API examples
This repo contains some code examples that show you how to use the [OpenFaaSFunction Builder API](https://docs.openfaas.com/openfaas-pro/builder/) from different languages.

> The Function Builder API provides a simple REST API to create your functions from source code.
> See [Function Builder API docs](https://docs.openfaas.com/openfaas-pro/builder/) for installation instructions

## Building functions form python
The [./python-request] folder has an example on how to invoke the Function Builder API from python. You can run the `build.py` file with the required arguments to turn a snippet of javascript into a container image.

The python script uses the [Requests](https://requests.readthedocs.io/en/latest/) package so you will have to install that to run the example.
```
python -m pip install requests
```

Run the script
```shell
cd python-request
python3 build.py docker.io/welteki/test-image-hello:0.1.3 ./hello-world
```
import argparse
import hmac
import subprocess
import json
import os
import tarfile
import hmac
import requests

parser = argparse.ArgumentParser(description='Build a NodeJs function.')
parser.add_argument('image', type=str)
parser.add_argument('handler', type=str)

args = parser.parse_args()

handler = os.path.abspath(args.handler)

def shrinkwrap(handler):
    cmd = [
        "faas-cli",
        "build",
        "--lang",
        "node17",
        "--handler",
        handler,
        "--name",
        "context",
        "--image",
        args.image,
        "--shrinkwrap"
    ]

    completed = subprocess.run(cmd)

    if completed.returncode != 0:
        raise Exception('Failed to shrinkwrap handler')

def makeTar(buildConfig, path, tarFile):
    configFile = os.path.join(path, 'com.openfaas.docker.config')
    with open(configFile, 'w') as f:
        json.dump(buildConfig, f)

    with tarfile.open(tarFile, 'w') as tar:
        tar.add(configFile, arcname='com.openfaas.docker.config')
        tar.add(os.path.join(path, "context"), arcname="context")

buildConfig = { 'image': args.image }
shrinkwrap(handler)
makeTar(buildConfig, 'build', 'req.tar')

with open('req.tar', 'rb') as t, open('payload.txt', 'rb') as s:
    secret = s.read()
    data = t.read()
    digest = hmac.new(secret, data, 'sha256').hexdigest()
    res = requests.post("http://127.0.0.1:8081/build", headers={'X-Build-Signature': 'sha256=%s'%digest, 'Content-Type': 'application/octet-stream'}, data=data)


if res.status_code != 200:
    print('Building image %s failed'%args.image)
else:
    content = res.json()
    print('Success building image %s'%content['image'])

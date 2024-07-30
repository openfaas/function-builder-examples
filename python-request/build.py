import argparse
import hmac
import subprocess
import json
import os
import tarfile
import hmac
import requests
import tempfile

def shrinkwrap(image, handler, lang):
    cmd = [
        "faas-cli",
        "build",
        "--lang",
        lang,
        "--handler",
        handler,
        "--name",
        "context",
        "--image",
        image,
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

def callBuilder(tarFile):
    with open(tarFile, 'rb') as t, open('payload.txt', 'r') as s:
        secret = s.read().strip()
        data = t.read()
        digest = hmac.new(bytes(secret, 'utf-8'), data, 'sha256').hexdigest()
        headers = {
            'X-Build-Signature': 'sha256={}'.format(digest),
            'Content-Type': 'application/octet-stream'
        }
        return requests.post("http://127.0.0.1:8081/build", headers=headers, data=data)

parser = argparse.ArgumentParser(
    description='Build a function with the OpenFaaS Pro Builder')

parser.add_argument('--image', type=str,
                    help="Docker image name to build", required=True)
parser.add_argument('--handler', type=str,
                    help="Directory with handler for function, e.g. handler.js", required=True)
parser.add_argument('--lang', type=str,
                    help="Language or template to use, e.g. node20", required=True)

args = parser.parse_args()

handler = os.path.abspath(args.handler)
buildConfig = {'image': args.image, 'buildArgs': {}}

with tempfile.TemporaryDirectory() as tmpdir:
    tarFile = os.path.join(tmpdir, 'req.tar')

    shrinkwrap(args.image, handler, args.lang)
    makeTar(buildConfig, 'build', tarFile)

    res = callBuilder(tarFile) 

content = res.json()
if res.status_code != 200:
    print('Building image {} failed'.format(args.image))
    print(content['status'])
else:
    print('Success building image %s' % content['image'])

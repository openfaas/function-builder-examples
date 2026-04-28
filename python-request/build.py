import argparse
import os
import tempfile

from openfaas_sdk.builder import BuildConfig, FunctionBuilder, create_build_context, make_tar

parser = argparse.ArgumentParser(
    description='Build a function with the OpenFaaS Pro Builder')

parser.add_argument('--image', type=str,
                    help="Docker image name to build", required=True)
parser.add_argument('--handler', type=str,
                    help="Directory with handler for function, e.g. handler.js", required=True)
parser.add_argument('--lang', type=str,
                    help="Language or template to use, e.g. node20", required=True)
parser.add_argument('--name', type=str,
                    help="Name of the function", required=True)
parser.add_argument('--platforms', type=str, default='linux/amd64',
                    help="Comma separated list of target platforms for multi-arch image builds.")
parser.add_argument('--build-args', type=str, default='',
                    help="Additional build arguments for the docker build in the form of key1=value1,key2=value2")
parser.add_argument('--builder-url', type=str, default='http://127.0.0.1:8081',
                    help="URL for the function builder (default: http://127.0.0.1:8081)")

args = parser.parse_args()

platforms = args.platforms.split(',')

build_args = {}
if args.build_args:
    for pair in args.build_args.split(','):
        kv = pair.split('=', 1)
        if len(kv) == 2:
            build_args[kv[0]] = kv[1]

# Get the HMAC secret used for payload authentication with the builder API.
with open('payload.txt', 'r') as f:
    payload_secret = f.read().strip()

# Initialize a new builder client.
builder = FunctionBuilder(args.builder_url, hmac_secret=payload_secret)

# Create the function build context using the provided function handler and language template.
build_context = create_build_context(args.name, os.path.abspath(args.handler), args.lang)

with tempfile.NamedTemporaryFile(suffix='.tar', delete=False) as tmp:
    tar_path = tmp.name

try:
    # Configuration for the build.
    # Set the image name plus optional build arguments and target platforms for multi-arch images.
    build_config = BuildConfig(
        image=args.image,
        platforms=platforms,
        build_args=build_args,
    )

    # Prepare a tar archive that contains the build config and build context.
    make_tar(tar_path, build_context, build_config)

    # Invoke the function builder with the tar archive containing the build config and context
    # to build and push the function image. Stream the build logs as they arrive.
    result = None
    for result in builder.build_stream(tar_path):
        for line in result.log:
            print(line)
finally:
    os.remove(tar_path)

if result:
    print('Image: {} built.'.format(result.image))

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/openfaas/go-sdk/builder"
)

var (
	image        string
	handler      string
	lang         string
	functionName string
	platformsStr string
	buildArgsStr string
)

func main() {
	flag.StringVar(&image, "image", "", "Docker image name to build")
	flag.StringVar(&handler, "handler", "", "Directory with handler for function, e.g. handler.js")
	flag.StringVar(&lang, "lang", "", "Language or template to use, e.g. node20")
	flag.StringVar(&functionName, "name", "", "Name of the function")
	flag.StringVar(&platformsStr, "platforms", "linux/amd64", "Comma separated list of target platforms for multi-arch image builds.")
	flag.StringVar(&buildArgsStr, "build-args", "", "Additional build arguments for the docker build in the form of key1=value1,key2=value2")
	flag.Parse()

	platforms := strings.Split(platformsStr, ",")
	buildArgs := parseBuildArgs(buildArgsStr)

	// Get the HMAC secret used for payload authentication with the builder API.
	payloadSecret, err := os.ReadFile("payload.txt")
	if err != nil {
		log.Fatal(err)
	}
	payloadSecret = bytes.TrimSpace(payloadSecret)

	// Initialize a new builder client.
	builderURL, _ := url.Parse("http://127.0.0.1:8081")
	b := builder.NewFunctionBuilder(builderURL, http.DefaultClient, builder.WithHmacAuth(string(payloadSecret)))

	// Create the function build context using the provided function handler and language template.
	buildContext, err := builder.CreateBuildContext(functionName, handler, lang, []string{})
	if err != nil {
		log.Fatalf("failed to create build context: %s", err)
	}

	// Create a temporary file for the build tar.
	tarFile, err := os.CreateTemp(os.TempDir(), "build-context-*.tar")
	if err != nil {
		log.Fatalf("failed to temporary file: %s", err)
	}
	tarFile.Close()

	tarPath := tarFile.Name()
	defer os.Remove(tarPath)

	// Configuration for the build.
	// Set the image name plus optional build arguments and target platforms for multi-arch images.
	buildConfig := builder.BuildConfig{
		Image:     image,
		Platforms: platforms,
		BuildArgs: buildArgs,
	}

	// Prepare a tar archive that contains the build config and build context.
	if err := builder.MakeTar(tarPath, buildContext, &buildConfig); err != nil {
		log.Fatal(err)
	}

	// Invoke the function builder with the tar archive containing the build config and context
	// to build and push the function image.
	result, err := b.Build(tarPath)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Image: %s built.", result.Image)

	// Print build logs
	log.Println("Build logs:")
	for _, logMsg := range result.Log {
		fmt.Printf("%s\n", logMsg)
	}
}

func parseBuildArgs(str string) map[string]string {
	buildArgs := map[string]string{}

	if str != "" {
		pairs := strings.Split(str, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				buildArgs[kv[0]] = kv[1]
			}
		}
	}

	return buildArgs
}

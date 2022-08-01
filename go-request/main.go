package main

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	hmac "github.com/alexellis/hmac/v2"
)

type buildConfig struct {
	Image     string            `json:"image"`
	BuildArgs map[string]string `json:"buildArgs,omitempty"`
}

type buildResult struct {
	Log    []string `json:"log"`
	Image  string   `json:"image"`
	Status string   `json:"status"`
}

const ConfigFileName = "com.openfaas.docker.config"

var (
	image   string
	handler string
	lang    string
)

func main() {
	flag.StringVar(&image, "image", "", "Docker image name to build")
	flag.StringVar(&handler, "handler", "", "Directory with handler for function, e.g. handler.js")
	flag.StringVar(&lang, "lang", "", "Language or template to use, e.g. node17")
	flag.Parse()

	tempDir, err := os.MkdirTemp(os.TempDir(), "builder-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tarPath := path.Join(tempDir, "req.tar")
	fmt.Println(tarPath)

	if err := shrinkwrap(image, handler, lang); err != nil {
		log.Fatal(err)
	}

	if err := makeTar(buildConfig{Image: image}, path.Join("build", "context"), tarPath); err != nil {
		log.Fatalf("Failed to create tar file: %s", err)
	}

	res, err := callBuilder(tarPath)
	if err != nil {
		log.Fatalf("Failed to call builder API: %s", err)
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)

	result := buildResult{}
	if err := json.Unmarshal(data, &result); err != nil {
		log.Fatalf("Failed to unmarshal build result: %s", err)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		log.Fatalf("Unable to build image %s: %s", image, result.Status)
	}

	log.Printf("Success building image: %s", result.Image)
}

func shrinkwrap(image, handler, lang string) error {
	buildCmd := exec.Command(
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
		"--shrinkwrap",
	)

	err := buildCmd.Start()
	if err != nil {
		return fmt.Errorf("cannot start faas-cli build: %t", err)
	}

	err = buildCmd.Wait()
	if err != nil {
		return fmt.Errorf("failed to shrinkwrap handler")
	}

	return nil
}

func makeTar(buildConfig buildConfig, base, tarPath string) error {
	configBytes, _ := json.Marshal(buildConfig)
	if err := ioutil.WriteFile(path.Join(base, ConfigFileName), configBytes, 0664); err != nil {
		return err
	}

	tarFile, err := os.Create(tarPath)
	if err != nil {
		return err
	}

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	err = filepath.Walk(base, func(path string, f os.FileInfo, pathErr error) error {
		if pathErr != nil {
			return pathErr
		}

		targetFile, err := os.Open(path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(f, f.Name())
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, base)
		if header.Name != fmt.Sprintf("/%s", ConfigFileName) {
			header.Name = filepath.Join("context", header.Name)
		}

		header.Name = strings.TrimPrefix(header.Name, "/")

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if f.Mode().IsDir() {
			return nil
		}

		_, err = io.Copy(tarWriter, targetFile)
		return err
	})

	return err
}

func callBuilder(tarPath string) (*http.Response, error) {
	payloadSecret, err := os.ReadFile("payload.txt")
	if err != nil {
		return nil, err
	}

	tarFile, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}
	defer tarFile.Close()

	tarFileBytes, err := ioutil.ReadAll(tarFile)
	if err != nil {
		return nil, err
	}

	digest := hmac.Sign(tarFileBytes, bytes.TrimSpace(payloadSecret), sha256.New)
	fmt.Println(hex.EncodeToString(digest))

	r, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8081/build", bytes.NewReader(tarFileBytes))
	if err != nil {
		return nil, err
	}

	r.Header.Set("X-Build-Signature", "sha256="+hex.EncodeToString(digest))
	r.Header.Set("Content-Type", "application/octet-stream")

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	return res, nil
}

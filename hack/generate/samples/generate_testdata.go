// Copyright 2021 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	gohelmsamples "github.com/operator-framework/helm-operator-plugins/hack/generate/samples/internal/hybrid/go-helm-samples"
	gosamples "github.com/operator-framework/helm-operator-plugins/hack/generate/samples/internal/hybrid/go-samples"
	helmsamples "github.com/operator-framework/helm-operator-plugins/hack/generate/samples/internal/hybrid/helm-samples"
	"github.com/operator-framework/helm-operator-plugins/hack/generate/samples/internal/samples"
)

func main() {
	// binaryPath allow inform the binary that should be used.
	// By default it is helm-operator-plugins
	var binaryPath string

	flag.StringVar(&binaryPath, "binaryPath", "bin/helm-operator-plugins", "Binary path that should be used")
	flag.Parse()

	// Make the binary path absolute if pathed, for reproducibility and debugging purposes.
	if dir, _ := filepath.Split(binaryPath); dir != "" {
		tmp, err := filepath.Abs(binaryPath)
		if err != nil {
			log.Fatalf("Failed to make binary path %q absolute: %v", binaryPath, err)
		}
		binaryPath = tmp
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// samplesPath is the path where all samples should be generated
	samplesPath := filepath.Join(wd, "testdata")
	log.Infof("writing sample directories under %s", samplesPath)

	var samples []samples.Sample

	samples = append(samples, gosamples.NewGoMemcached(gosamples.WithBinaryPath(binaryPath), gosamples.WithSamplesPath(filepath.Join(samplesPath, "hybrid", "go"))))
	samples = append(samples, helmsamples.NewHelmSample(helmsamples.WithBinaryPath(binaryPath), helmsamples.WithSamplesPath(filepath.Join(samplesPath, "hybrid", "helm"))))
	samples = append(samples, gohelmsamples.NewGoHelmMemcached(gohelmsamples.WithBinaryPath(binaryPath), gohelmsamples.WithSamplesPath(filepath.Join(samplesPath, "hybrid", "go-helm"))))

	for _, sample := range samples {
		log.Infof("creating sample at path: %s", sample.Path())
		sample.Generate()
	}
}

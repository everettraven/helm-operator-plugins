package gohelmsamples

import (
	"os"
	"path/filepath"

	"github.com/operator-framework/helm-operator-plugins/hack/generate/samples/internal/pkg"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GoHelmMemcached defines the Memcached Sample in GoHelm
type GoHelmMemcached struct {
	ctx         *pkg.SampleContext
	binaryPath  string
	samplesPath string
	sampleName  string
	helmGvk     schema.GroupVersionKind
	goGvk       schema.GroupVersionKind
	repo        string
}

type GoHelmMemcachedOption func(ghm *GoHelmMemcached)

func WithBinaryPath(binaryPath string) GoHelmMemcachedOption {
	return func(ghm *GoHelmMemcached) {
		ghm.binaryPath = binaryPath
	}
}

func WithSamplesPath(samplesPath string) GoHelmMemcachedOption {
	return func(ghm *GoHelmMemcached) {
		ghm.samplesPath = samplesPath
	}
}

func WithSampleName(name string) GoHelmMemcachedOption {
	return func(ghm *GoHelmMemcached) {
		ghm.sampleName = name
	}
}

func WithHelmGVK(gvk schema.GroupVersionKind) GoHelmMemcachedOption {
	return func(ghm *GoHelmMemcached) {
		ghm.helmGvk = gvk
	}
}

func WithGoGVK(gvk schema.GroupVersionKind) GoHelmMemcachedOption {
	return func(ghm *GoHelmMemcached) {
		ghm.goGvk = gvk
	}
}

func NewGoHelmMemcached(opts ...GoHelmMemcachedOption) *GoHelmMemcached {
	ghm := &GoHelmMemcached{
		ctx:         nil,
		binaryPath:  "bin/helm-operator-plugins",
		samplesPath: "testdata/hybrid/helm",
		sampleName:  "memcached-operator",
		helmGvk: schema.GroupVersionKind{
			Group:   "cache",
			Version: "v1alpha1",
			Kind:    "Memcached",
		},
		goGvk: schema.GroupVersionKind{
			Group:   "cache",
			Version: "v1",
			Kind:    "MemcachedBackup",
		},
		repo: "github.com/example/memcached-operator",
	}

	for _, opt := range opts {
		opt(ghm)
	}

	ctx, err := pkg.NewSampleContext(ghm.binaryPath, filepath.Join(ghm.samplesPath, ghm.sampleName), "GO111MODULE=on")
	pkg.CheckError("generating GoHelm memcached context", err)

	ghm.ctx = &ctx

	return ghm
}

func (ghm *GoHelmMemcached) Path() string {
	return filepath.Join(ghm.samplesPath, ghm.sampleName)
}

// GenerateGoHelmMemcachedSample will call all actions to create the directory and generate the sample
// The Context to run the samples are not the same in the e2e test. In this way, note that it should NOT
// be called in the e2e tests since it will call the Prepare() to set the sample context and generate the files
// in the testdata directory. The e2e tests only ought to use the Run() method with the TestContext.
func (ghm *GoHelmMemcached) Generate() {
	ghm.Prepare()
	ghm.Run()
}

func (ghm *GoHelmMemcached) Prepare() {
	log.Infof("destroying directory for memcached GoHelm samples")
	ghm.ctx.Destroy()

	log.Infof("creating directory")
	err := ghm.ctx.Prepare()
	pkg.CheckError("creating directory", err)
}

func (ghm *GoHelmMemcached) Run() {
	// When we scaffold GoHelm based projects, it tries to use the discovery API of a Kubernetes
	// cluster to intelligently build the RBAC rules that the operator will require based on the
	// content of the helm chart.
	//
	// Here, we intentionally set KUBECONFIG to a broken value to ensure that operator-sdk will be
	// unable to reach a real cluster, and thus will generate a default RBAC rule set. This is
	// required to make GoHelm project generation idempotent because contributors and CI environments
	// can all have slightly different environments that can affect the content of the generated
	// role and cause sanity testing to fail.
	os.Setenv("KUBECONFIG", "broken_so_we_generate_static_default_rules")
	log.Infof("using init command and scaffolding the project")

	err := ghm.ctx.Init(
		"--plugins", "hybrid/v1-alpha",
		"--repo", ghm.repo,
	)

	pkg.CheckError("creating the project", err)

	err = ghm.ctx.CreateAPI(
		"--plugins", "base.helm.sdk.operatorframework.io/v1",
		"--group", ghm.helmGvk.Group,
		"--version", ghm.helmGvk.Version,
		"--kind", ghm.helmGvk.Kind,
	)

	pkg.CheckError("creating helm api", err)

	err = ghm.ctx.CreateAPI(
		"--plugins", "base.go.kubebuilder.io/v3",
		"--group", ghm.goGvk.Group,
		"--version", ghm.goGvk.Version,
		"--kind", ghm.goGvk.Kind,
		"--resource", "--controller",
	)

	pkg.CheckError("creating go api", err)
}

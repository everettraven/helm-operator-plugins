package gosamples

import (
	"os"
	"path/filepath"

	"github.com/operator-framework/helm-operator-plugins/hack/generate/samples/internal/pkg"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GoMemcached defines the Memcached Sample in Go
type GoMemcached struct {
	ctx         *pkg.SampleContext
	binaryPath  string
	samplesPath string
	sampleName  string
	gvk         schema.GroupVersionKind
	repo        string
}

type GoMemcachedOption func(gm *GoMemcached)

func WithBinaryPath(binaryPath string) GoMemcachedOption {
	return func(gm *GoMemcached) {
		gm.binaryPath = binaryPath
	}
}

func WithSamplesPath(samplesPath string) GoMemcachedOption {
	return func(gm *GoMemcached) {
		gm.samplesPath = samplesPath
	}
}

func WithSampleName(name string) GoMemcachedOption {
	return func(gm *GoMemcached) {
		gm.sampleName = name
	}
}

func WithGVK(gvk schema.GroupVersionKind) GoMemcachedOption {
	return func(gm *GoMemcached) {
		gm.gvk = gvk
	}
}

func NewGoMemcached(opts ...GoMemcachedOption) *GoMemcached {
	gm := &GoMemcached{
		ctx:         nil,
		binaryPath:  "bin/helm-operator-plugins",
		samplesPath: "testdata/hybrid/go",
		sampleName:  "memcached-operator",
		gvk: schema.GroupVersionKind{
			Group:   "cache",
			Version: "v1alpha1",
			Kind:    "Memcached",
		},
		repo: "github.com/example/memcached-operator",
	}

	for _, opt := range opts {
		opt(gm)
	}

	ctx, err := pkg.NewSampleContext(gm.binaryPath, filepath.Join(gm.samplesPath, gm.sampleName), "GO111MODULE=on")
	pkg.CheckError("generating Go memcached context", err)

	gm.ctx = &ctx

	return gm
}

func (gm *GoMemcached) Path() string {
	return filepath.Join(gm.samplesPath, gm.sampleName)
}

// GenerateGoMemcachedSample will call all actions to create the directory and generate the sample
// The Context to run the samples are not the same in the e2e test. In this way, note that it should NOT
// be called in the e2e tests since it will call the Prepare() to set the sample context and generate the files
// in the testdata directory. The e2e tests only ought to use the Run() method with the TestContext.
func (gm *GoMemcached) Generate() {
	gm.Prepare()
	gm.Run()
}

func (gm *GoMemcached) Prepare() {
	log.Infof("destroying directory for memcached Go samples")
	gm.ctx.Destroy()

	log.Infof("creating directory")
	err := gm.ctx.Prepare()
	pkg.CheckError("creating directory", err)
}

func (gm *GoMemcached) Run() {
	// When we scaffold Helm based projects, it tries to use the discovery API of a Kubernetes
	// cluster to intelligently build the RBAC rules that the operator will require based on the
	// content of the helm chart.
	//
	// Here, we intentionally set KUBECONFIG to a broken value to ensure that operator-sdk will be
	// unable to reach a real cluster, and thus will generate a default RBAC rule set. This is
	// required to make Helm project generation idempotent because contributors and CI environments
	// can all have slightly different environments that can affect the content of the generated
	// role and cause sanity testing to fail.
	os.Setenv("KUBECONFIG", "broken_so_we_generate_static_default_rules")
	log.Infof("using init command and scaffolding the project")

	err := gm.ctx.Init(
		"--plugins", "hybrid/v1-alpha",
		"--repo", gm.repo,
	)

	pkg.CheckError("creating the project", err)

	err = gm.ctx.CreateAPI(
		"--plugins", "base.go.kubebuilder.io/v3",
		"--group", gm.gvk.Group,
		"--version", gm.gvk.Version,
		"--kind", gm.gvk.Kind,
		"--resource", "--controller",
	)

	pkg.CheckError("creating go api", err)
}

package helmsamples

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/operator-framework/helm-operator-plugins/hack/generate/samples/internal/pkg"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kbutil "sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
)

// HelmSample defines a Sample project scaffolded with Helm
type HelmSample struct {
	ctx         *pkg.SampleContext
	binaryPath  string
	samplesPath string
	sampleName  string
	gvk         schema.GroupVersionKind
	repo        string
	domain      string
}

type HelmSampleOption func(hs *HelmSample)

func WithBinaryPath(binaryPath string) HelmSampleOption {
	return func(hs *HelmSample) {
		hs.binaryPath = binaryPath
	}
}

func WithSamplesPath(samplesPath string) HelmSampleOption {
	return func(hs *HelmSample) {
		hs.samplesPath = samplesPath
	}
}

func WithSampleName(name string) HelmSampleOption {
	return func(hs *HelmSample) {
		hs.sampleName = name
	}
}

func WithGVK(gvk schema.GroupVersionKind) HelmSampleOption {
	return func(hs *HelmSample) {
		hs.gvk = gvk
	}
}

func WithDomain(domain string) HelmSampleOption {
	return func(hs *HelmSample) {
		hs.domain = domain
	}
}

func NewHelmSample(opts ...HelmSampleOption) *HelmSample {
	hs := &HelmSample{
		ctx:         nil,
		binaryPath:  "bin/helm-operator-plugins",
		samplesPath: "testdata/hybrid/helm",
		sampleName:  "memcached-operator",
		gvk: schema.GroupVersionKind{
			Group:   "cache",
			Version: "v1alpha1",
			Kind:    "Memcached",
		},
		repo:   "github.com/example/memcached-operator",
		domain: "example.com",
	}

	for _, opt := range opts {
		opt(hs)
	}

	ctx, err := pkg.NewSampleContext(hs.binaryPath, filepath.Join(hs.samplesPath, hs.sampleName), "GO111MODULE=on")
	pkg.CheckError("generating Helm memcached context", err)

	hs.ctx = &ctx

	return hs
}

func (hs *HelmSample) Path() string {
	return filepath.Join(hs.samplesPath, hs.sampleName)
}

// GenerateHelmSampleSample will call all actions to create the directory and generate the sample
// The Context to run the samples are not the same in the e2e test. In this way, note that it should NOT
// be called in the e2e tests since it will call the Prepare() to set the sample context and generate the files
// in the testdata directory. The e2e tests only ought to use the Run() method with the TestContext.
func (hs *HelmSample) Generate() {
	hs.Prepare()
	hs.Run()
}

func (hs *HelmSample) Prepare() {
	log.Infof("destroying directory for memcached Helm samples")
	hs.ctx.Destroy()

	log.Infof("creating directory")
	err := hs.ctx.Prepare()
	pkg.CheckError("creating directory", err)
}

func (hs *HelmSample) Run() {
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

	err := hs.ctx.Init(
		"--plugins", "hybrid/v1-alpha",
		"--repo", hs.repo,
		"--domain", hs.domain,
	)

	pkg.CheckError("creating the project", err)

	err = hs.ctx.CreateAPI(
		"--plugins", "base.helm.sdk.operatorframework.io/v1",
		"--group", hs.gvk.Group,
		"--version", hs.gvk.Version,
		"--kind", hs.gvk.Kind,
	)

	pkg.CheckError("creating helm api", err)

	log.Infof("enabling prometheus metrics")
	err = kbutil.UncommentCode(
		filepath.Join(hs.ctx.Dir, "config", "default", "kustomization.yaml"),
		"#- ../prometheus", "#")
	pkg.CheckError("enabling prometheus metrics", err)

	log.Infof("adding customized roles")
	err = kbutil.ReplaceInFile(filepath.Join(hs.ctx.Dir, "config", "rbac", "role.yaml"),
		"#+kubebuilder:scaffold:rules", fmt.Sprintf(policyRolesFragment, hs.gvk.Group, hs.domain, hs.gvk.Version, hs.gvk.Kind))
	pkg.CheckError("adding customized roles", err)

}

const policyRolesFragment = `
##
## Rules customized for %s.%s/%s, Kind: %s
##
- apiGroups:
  - policy
  resources:
  - events
  - poddisruptionbudgets
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  - services
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
#+kubebuilder:scaffold:rules
`

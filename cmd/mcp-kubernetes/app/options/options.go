package options

import (
	mcpkubernetes "github.com/fleezesd/mcp-kubernetes/internal/mcp-kubernetes"
	"github.com/fleezesd/mcp-kubernetes/pkg/app"
	"github.com/fleezesd/mcp-kubernetes/pkg/log"
	"github.com/spf13/viper"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
)

var _ app.CliOptions = (*Options)(nil)

type Options struct {
	SSEPort    int          `json:"sse-port" mapstructure:"sse-port"`
	SSEBaseURL string       `json:"sse-base-url" mapstructure:"sse-base-url"`
	KubeConfig string       `json:"kubeconfig" mapstructure:"kubeconfig"`
	Log        *log.Options `json:"log" mapstructure:"log"`
}

func NewOptions() *Options {
	o := &Options{
		Log: log.NewOptions(),
	}
	return o
}

func (o *Options) Flags() (fss cliflag.NamedFlagSets) {
	o.Log.AddFlags(fss.FlagSet("logs"))
	fs := fss.FlagSet("mcp-kubernetes-server")
	fs.IntVar(&o.SSEPort, "sse-port", 0, "Start a SSE server on the specified port")
	fs.StringVar(&o.SSEBaseURL, "sse-base-url", "", "SSE public base URL to use when sending the endpoint message (e.g. https://example.com)")
	fs.StringVar(&o.KubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	return fss
}

func (o *Options) Complete() error {
	if err := viper.Unmarshal(&o); err != nil {
		return err
	}
	return nil
}

func (o *Options) Validate() error {
	errs := []error{}

	errs = append(errs, o.Log.Validate()...)
	return utilerrors.NewAggregate(errs)
}

func (o *Options) ApplyTo(c *mcpkubernetes.Config) error {
	c.SSEPort = o.SSEPort
	c.SSEBaseURL = o.SSEBaseURL
	c.KubeConfig = o.KubeConfig
	return nil
}

// Config return xnightwatch config object.
func (o *Options) Config() (*mcpkubernetes.Config, error) {
	c := &mcpkubernetes.Config{}

	if err := o.ApplyTo(c); err != nil {
		return nil, err
	}
	return c, nil
}

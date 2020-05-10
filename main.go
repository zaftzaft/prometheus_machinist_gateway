package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	resty "github.com/go-resty/resty/v2"
	prometheusAPI "github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"
)

var (
	configFile = kingpin.Flag("config", "config file").Required().Short('c').String()
	dryRun     = kingpin.Flag("dry-run", "dry run").Bool()
)

type Config struct {
	PrometheusURL  string         `yaml:"prometheus_url"`
	MachinistToken string         `yaml:"machinist_token"`
	Interval       time.Duration  `yaml:"interval,omitempty"`
	AgentConfigs   []*AgentConfig `yaml:"agent_configs,omitempty"`
}

type AgentConfig struct {
	AgentName    string            `yaml:"agent_name"`
	Query        string            `yaml:"query"`
	Namespace    string            `yaml:"namespace,omitempty"`
	TagIncludes  []string          `yaml:"tag_includes,omitempty"`
	MetaIncludes []string          `yaml:"meta_includes,omitempty"`
	Tag          map[string]string `yaml:"tag,omitempty"`
	Meta         map[string]string `yaml:"meta,omitempty"`
}

type MachinistMetric struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
	DataPoint struct {
		Timestamp int               `json:"timestamp,omitempty"`
		Value     float64           `json:"value"`
		Meta      map[string]string `json:"meta,omitempty"`
	} `json:"data_point"`
}

type MachinistRequest struct {
	Agent   string             `json:"agent"`
	Metrics []*MachinistMetric `json:"metrics"`
}

func PullAndPush(cfg *Config) error {
	for _, agent := range cfg.AgentConfigs {

		client, err := prometheusAPI.NewClient(prometheusAPI.Config{
			Address: cfg.PrometheusURL,
		})
		if err != nil {
			return err
		}

		v1api := v1.NewAPI(client)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result, warnings, err := v1api.Query(ctx, agent.Query, time.Time{})
		if err != nil {
			return err
		}

		for _, warn := range warnings {
			log.Warn(warn)
		}
		//fmt.Printf("%v\n", warnings)
		//fmt.Printf("%v\n", result.Type())
		//fmt.Printf("%v\n", result.(model.Vector))

		req := &MachinistRequest{}
		req.Agent = agent.AgentName

		// TODO support non Vector
		for _, sample := range result.(model.Vector) {
			m := &MachinistMetric{}
			m.Name = string(sample.Metric[model.MetricNameLabel])
			m.DataPoint.Value = float64(sample.Value)
			m.DataPoint.Meta = map[string]string{}
			m.Tags = map[string]string{}

			if len(agent.Namespace) > 0 {
				m.Namespace = agent.Namespace
			}

			//fmt.Println(sample.Metric[model.MetricNameLabel])
			for k, v := range sample.Metric {
				if k == model.MetricNameLabel {
					continue
				}

				for _, tag := range agent.TagIncludes {
					if string(k) == tag {
						m.Tags[string(k)] = string(v)
					}
				}

				for _, meta := range agent.MetaIncludes {
					if string(k) == meta {
						m.DataPoint.Meta[string(k)] = string(v)
					}
				}
			}

			for k, v := range agent.Tag {
				m.Tags[k] = v
			}

			for k, v := range agent.Meta {
				m.DataPoint.Meta[k] = v
			}

			req.Metrics = append(req.Metrics, m)
		}

		data, err := json.Marshal(req)
		if err != nil {
			return err
		}

		if *dryRun {
			var out bytes.Buffer
			json.Indent(&out, data, "", "  ")
			fmt.Println(out.String())
		} else {
			client := resty.New()
			resp, err := client.R().
				SetBody(data).
				SetAuthToken(cfg.MachinistToken).
				Post("https://gw.machinist.iij.jp/endpoint")
			if err != nil {
				return err
			}

			log.Info(resp)
		}

	}

	return nil
}

func Run() int {
	cfg := &Config{}

	content, err := ioutil.ReadFile(*configFile)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	if err := yaml.UnmarshalStrict(content, cfg); err != nil {
		fmt.Println(err)
		return 1
	}

	if cfg.Interval == 0 {
		cfg.Interval = time.Second * 60
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		if err := PullAndPush(cfg); err != nil {
			log.Error(err)
		}

		if *dryRun {
			break
		}

	}

	return 0
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	kingpin.Version("0.0.1")
	kingpin.Parse()
	os.Exit(Run())
}

package provider

import (
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELKALIASES_URL", nil),
				Description: "The URL for the Elasticsearch instance.",
			},
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ELKALIASES_TOKEN", nil),
				Description: "The token for API authentication.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"elkaliases_index": resourceelkAliasesIndex(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	url := d.Get("url").(string)
	token := d.Get("token").(string)

	cfg := elasticsearch.Config{
		Addresses: []string{url},
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Header: http.Header{
			"Authorization": []string{fmt.Sprintf("ApiKey %s", token)},
		},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating Elasticsearch client: %s", err)
	}

	return es, nil
}

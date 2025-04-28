---
layout: "elkaliases"
page_title: "Provider: Elkaliases"
sidebar_current: "docs-elkaliases-index"
description: |-
  A provider for ElkAliases Server.
---

# ElkAliases Provider

The ElkAliases provider gives the ability to add aliases on indexes, created by an [index templates](https://registry.terraform.io/providers/elastic/elasticstack/latest/docs/resources/elasticsearch_index_template).

With the current provider from elasticstack, there is no solution to [add aliases to an existing index](https://github.com/elastic/terraform-provider-elasticstack/issues/285).
To solve this problem, we have created this provider, which enables you through the `elkaliases_index_aliases` resource to add aliases to existing index or data stream.
If you use this provider, you should not declare `alias` with `elasticstack_elasticsearch_index_template` resource.

This provider is a workaround, waiting for a long term solution in [the official provider](https://registry.terraform.io/providers/elastic/elasticstack/latest).

Use the navigation to the left to read about the available resources.

## Usage

```hcl
provider "elkaliases" {
  url   = "http://localhost:9200"
  token = "token"
}
```

### Environment Variables
Provider settings can be specified via environment variables as follows:

```shell
export ELASTICSEARCH_ENDPOINT=http://localhost:9200
export ELASTICSEARCH_API_KEY=token
```

## Argument Reference

The following arguments are supported:

* `url` - (Required) Url to the ElasticSsearch API
* `token` - (Required) Authentication token to the ElasticSearch API
* `insecure` - (Optional) Skip server certification verification

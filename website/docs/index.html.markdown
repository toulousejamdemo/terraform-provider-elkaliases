---
layout: "elkaliases"
page_title: "Provider: Elkaliases"
sidebar_current: "docs-elkaliases-index"
description: |-
  A provider for ElkAliases Server.
---

# ElkAliases Provider

The ElkAliases provider gives the ability to deploy index template to a ElasticSearch server

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

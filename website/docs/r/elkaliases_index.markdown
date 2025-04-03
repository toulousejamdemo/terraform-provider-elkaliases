---
layout: "elkaliases"
page_title: "ElkAliases: elkaliases-index"
sidebar_current: "docs-elkaliases-resource-elkaliases_index"
description: |-
  Create an index template
---

# alkaliases\_index

The ``elkaliases_index`` resource creates an index template

See [ElasticSearch documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index-templates.html)

## Usage

```hcl
resource "elkaliases_index" "name" {
  name = "name"

  index_patterns = ["pattern"]

  template {
    mappings = jsonencode({})
    settings = jsonencode({})
    alias {
      name = "alias_name"
      filter = jsonencode({})
    }
  }

  composed_of = []
}
```

## Argument Reference

* `name` - (Required) Name of the index template to create.
* `index_patterns` - (Required) Array of wildcard (*) expressions used to match the names of data streams and indices during creation.
* `template` - (Required) Template to be applied. It may optionally include an aliases, mappings, lifecycle, or settings configuration.
  * `mappings` - (Required) Mapping for fields in the index. Should be specified as a JSON object of field mappings. See the documentation (https://www.elastic.co/guide/en/elasticsearch/reference/current/explicit-mapping.html) for more details
  * `settings` - (Required) Configuration options for the index. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#index-modules-settings
  * `alias` - (Required) Alias to add.
    * `name` - (Required) The alias name.
    * `filter` - (Required) Query used to limit documents the alias can access.
* `composed_of` - (Required) An ordered list of component template names.
* `data_stream` - (Optional) If this object is included, the template is used to create data streams and their backing indices. Supports an empty object.
  * `allow_custom_routing` - (Optional) If true, the data stream supports custom routing. Defaults to false.
  * `hidden` - (Optional) If true, the data stream is hidden. Defaults to false.

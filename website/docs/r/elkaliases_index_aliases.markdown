---
layout: "elkaliases"
page_title: "ElkAliases: elkaliases-index-aliases"
sidebar_current: "docs-elkaliases-resource-elkaliases_index_aliases"
description: |-
  Create aliases for an index
---

# alkaliases\_index

The ``elkaliases_index_aliases`` resource creates aliases in an existing index.


~> **Note**: The index needs to already exist at deployment time. This means that you cannot create an index template and apply aliases to it directly. You need to create an index or data stream first.

## Usage

```hcl
resource "elkaliases_index_aliases" "name" {
  index = "name"

  alias {
    name = "test"
    filter = jsonencode({
      "match" = {
        "event.dataset" = "test"
      }
    })
  }

  alias {
    name = "name"
  }
}
```

## Argument Reference

* `index` - (Required) Index or data stream in which create the aliases.
* `alias` - (Required) Alias to create in the index or data stream.
  * `name` - (Required) Name of the alias.
  * `filter` - (Optional) Filter to apply to the alias.

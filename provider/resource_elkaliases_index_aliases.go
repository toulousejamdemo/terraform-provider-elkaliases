package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceElkaliasesIndexAliases() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceElkaliasesAliasCreate,
		ReadContext:   resourceElkaliasesAliasRead,
		UpdateContext: resourceElkaliasesAliasUpdate,
		DeleteContext: resourceElkaliasesAliasDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceElkaliasesAliasImport,
		},

		Schema: map[string]*schema.Schema{
			"index": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Data stream or index for the actions.",
			},
			"alias": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "List of aliases to create",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Alias for the action.",
						},
						"filter": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     nil,
							Description: "Query used to limit documents the alias can access.",
						},
					},
				},
			},
		},
	}
}

func resourceElkaliasesAliasCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*elasticsearch.Client)
	aliases := d.Get("alias").([]any)
	indexName := d.Get("index").(string)

	actions := make([]any, 0)
	for _, alias := range aliases {
		alias := alias.(map[string]any)
		aliasName := alias["name"].(string)
		action := map[string]any{
			"index": indexName,
			"alias": aliasName,
		}
		if filter, ok := alias["filter"].(string); ok && filter != "" {
			action["filter"] = json.RawMessage(filter)
		}
		actions = append(actions, map[string]any{
			"add": action,
		})
	}

	if _, err := updateAliases(ctx, client, actions); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(indexName)

	return resourceElkaliasesAliasRead(ctx, d, m)
}

func resourceElkaliasesAliasRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*elasticsearch.Client)
	indexName := d.Id()

	res, err := client.Indices.GetAlias(client.Indices.GetAlias.WithIndex(indexName), client.Indices.GetAlias.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return diag.Errorf("Error response from elasticsearch: %s", res.String())
	}

	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return diag.FromErr(err)
	}

	indexBody, ok := body[indexName].(map[string]any)
	if !ok {
		d.SetId("")
	}

	aliases := make([]any, 0)
	aliasesState := d.Get("alias").([]any)
	aliasesBody := indexBody["aliases"].(map[string]any)

	for _, aliasState := range aliasesState {
		aliasState := aliasState.(map[string]any)
		if aliasBody, exist := aliasesBody[aliasState["name"].(string)]; exist {
			aliasBody := aliasBody.(map[string]any)
			alias := map[string]any{
				"name": aliasState["name"].(string),
			}
			if filter, ok := aliasBody["filter"]; ok {
				filterJson, err := json.Marshal(filter)
				if err != nil {
					return diag.FromErr(err)
				}
				alias["filter"] = string(filterJson)
			}
			aliases = append(aliases, alias)
		}
	}

	for aliasName, aliasContent := range aliasesBody {
		aliasContent := aliasContent.(map[string]any)
		if !isMapKeyInArray(aliases, "name", aliasName) {
			alias := map[string]any{
				"name": aliasName,
			}
			if filter, ok := aliasContent["filter"]; ok {
				filterJson, err := json.Marshal(filter)
				if err != nil {
					return diag.FromErr(err)
				}
				alias["filter"] = string(filterJson)
			}
			aliases = append(aliases, alias)
		}
	}

	d.Set("alias", aliases)

	return nil
}

func resourceElkaliasesAliasUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*elasticsearch.Client)
	indexName := d.Id()

	if d.HasChange("alias") {
		old, new := d.GetChange("alias")
		oldArray := old.([]any)
		newArray := new.([]any)

		actions := make([]any, 0)
		for _, elem := range oldArray {
			elem := elem.(map[string]any)
			if !isMapKeyInArray(newArray, "name", elem["name"]) {
				aliasName := elem["name"].(string)
				actions = append(actions, map[string]any{
					"remove": map[string]any{
						"index": indexName,
						"alias": aliasName,
					},
				})
			}
		}

		if len(actions) != 0 {
			if _, err := updateAliases(ctx, client, actions); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("index") || d.HasChange("alias") {
		if err := resourceElkaliasesAliasDelete(ctx, d, m); err != nil {
			return err
		}
		if err := resourceElkaliasesAliasCreate(ctx, d, m); err != nil {
			return err
		}
	}

	return resourceElkaliasesAliasRead(ctx, d, m)
}

func resourceElkaliasesAliasDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*elasticsearch.Client)
	indexName := d.Id()
	aliases := d.Get("alias").([]any)

	actions := make([]any, 0)
	for _, alias := range aliases {
		alias := alias.(map[string]any)
		aliasName := alias["name"].(string)
		actions = append(actions, map[string]any{
			"remove": map[string]any{
				"index": indexName,
				"alias": aliasName,
			},
		})
	}

	if code, err := updateAliases(ctx, client, actions); err != nil && code != 404 {
		return diag.FromErr(err)
	}
	return nil
}

func resourceElkaliasesAliasImport(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
	err := resourceElkaliasesAliasRead(ctx, d, m)
	if err != nil {
		return nil, fmt.Errorf("error while import: %v", err)
	}
	return []*schema.ResourceData{d}, nil
}

func updateAliases(ctx context.Context, client *elasticsearch.Client, content []any) (int, error) {
	body, err := json.Marshal(map[string]any{
		"actions": content,
	})
	if err != nil {
		return 0, err
	}
	res, err := client.Indices.UpdateAliases(bytes.NewReader(body), client.Indices.UpdateAliases.WithContext(ctx))
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return res.StatusCode, fmt.Errorf("error response from elasticsearch: %s", res.String())
	}
	return res.StatusCode, nil
}

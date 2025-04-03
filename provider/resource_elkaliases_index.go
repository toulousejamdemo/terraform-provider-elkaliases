package provider

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceelkAliasesIndex() *schema.Resource {
	return &schema.Resource{
		Create: resourceelkAliasesIndexCreate,
		Read:   resourceelkAliasesIndexRead,
		Update: resourceelkAliasesIndexUpdate,
		Delete: resourceelkAliasesIndexDelete,
		Importer: &schema.ResourceImporter{
			State: resourceelkAliasesIndexImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"index_patterns": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
			"data_stream": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_custom_routing": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"hidden": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"template": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mappings": {
							Type:     schema.TypeString,
							Required: true,
						},
						"settings": {
							Type:     schema.TypeString,
							Required: true,
						},
						"alias": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"filter": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"composed_of": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
		},
	}
}

func resourceelkAliasesIndexCreate(d *schema.ResourceData, m interface{}) error {
	es := m.(*elasticsearch.Client)

	name := d.Get("name").(string)
	indexPatterns := d.Get("index_patterns").([]interface{})
	mappings := d.Get("template.0.mappings").(string)
	settings := d.Get("template.0.settings").(string)
	aliasesData := d.Get("template.0.alias").([]interface{})
	composedOf := d.Get("composed_of").([]interface{})

	// Convert index patterns to a slice of strings
	var patterns []string
	for _, pattern := range indexPatterns {
		patterns = append(patterns, pattern.(string))
	}

	// Convert aliases to the correct structure
	aliases := make(map[string]interface{})
	for _, alias := range aliasesData {
		aliasMap := alias.(map[string]interface{})
		aliasName := aliasMap["name"].(string)
		aliasFilter := aliasMap["filter"].(string)
		aliases[aliasName] = map[string]interface{}{
			"filter": json.RawMessage(aliasFilter),
		}
	}

	// Create the index template body
	templateBody := map[string]interface{}{
		"index_patterns": patterns,
		"template": map[string]interface{}{
			"mappings": json.RawMessage(mappings),
			"settings": json.RawMessage(settings),
			"aliases":  aliases,
		},
		"composed_of": composedOf,
	}

	// Convert data_stream
	if value, ok := d.GetOk("data_stream"); ok {
		content := value.([]any)[0].(map[string]any)
		templateBody["data_stream"] = map[string]any{
			"allow_custom_routing": content["allow_custom_routing"],
			"hidden":               content["hidden"],
		}
	}

	// Convert templateBody to JSON
	body, err := json.Marshal(templateBody)
	if err != nil {
		return fmt.Errorf("error marshaling template body: %s", err)
	}

	// Send the request to create the index template
	res, err := es.Indices.PutIndexTemplate(name, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error creating Elasticsearch index template: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	// Set the ID for Terraform state
	d.SetId(name)
	return resourceelkAliasesIndexRead(d, m)
}

func resourceelkAliasesIndexRead(d *schema.ResourceData, m interface{}) error {
	es := m.(*elasticsearch.Client)

	name := d.Id()

	// Get the index template
	res, err := es.Indices.GetIndexTemplate(es.Indices.GetIndexTemplate.WithName(name))
	if err != nil {
		return fmt.Errorf("error getting Elasticsearch index template: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		// Template not found, remove from state
		d.SetId("")
		return nil
	}

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return fmt.Errorf("error decoding Elasticsearch response: %s", err)
	}

	// Extract the template data
	templates, ok := response["index_templates"].([]interface{})
	if !ok || len(templates) == 0 {
		d.SetId("")
		return nil
	}

	template := templates[0].(map[string]interface{})
	templateDetails := template["index_template"].(map[string]interface{})

	// Prepare the template map to set in Terraform state
	templateVar := map[string]interface{}{}

	// Handle index_patterns
	if patterns, ok := templateDetails["index_patterns"].([]interface{}); ok {
		d.Set("index_patterns", patterns)
	}

	// Handle composed_of
	if composedOf, ok := templateDetails["composed_of"].([]interface{}); ok {
		d.Set("composed_of", composedOf)
	}

	// Handle data_stream
	if dataStream, ok := templateDetails["data_stream"].(map[string]any); ok {
		dataSteamTemplate := make(map[string]any)
		if routing, ok := dataStream["allow_custom_routing"]; ok {
			dataSteamTemplate["allow_custom_routing"] = routing
		}

		if hidden, ok := dataStream["hidden"]; ok {
			dataSteamTemplate["hidden"] = hidden
		}

		d.Set("data_stream", []any{dataSteamTemplate})
	} else {
		d.Set("data_stream", nil)
	}

	// Extract and handle template components
	if templateContent, ok := templateDetails["template"].(map[string]interface{}); ok {
		// Handle mappings
		if mappings, ok := templateContent["mappings"]; ok {
			mappingsJSON, err := json.Marshal(mappings)
			if err != nil {
				return fmt.Errorf("error marshaling mappings: %s", err)
			}
			templateVar["mappings"] = string(mappingsJSON)
		}

		// Handle settings
		if settings, ok := templateContent["settings"]; ok {
			settingsJSON, err := json.Marshal(settings)
			if err != nil {
				return fmt.Errorf("error marshaling settings: %s", err)
			}
			templateVar["settings"] = string(settingsJSON)
		}

		// Handle aliases
		if aliases, ok := templateContent["aliases"].(map[string]any); ok {
			stateAliases := d.Get("template.0.alias").([]any)
			var aliasList []any

			for _, alias := range stateAliases {
				alias := alias.(map[string]any)
				if config, exist := aliases[alias["name"].(string)]; exist {
					filterJson, err := json.Marshal(config.(map[string]any)["filter"])
					if err != nil {
						return fmt.Errorf("error marshaling filter: %s", err)
					}
					aliasList = append(aliasList, map[string]any{
						"name":   alias["name"],
						"filter": string(filterJson),
					})
				}
			}

			for name, config := range aliases {
				if !isInMap(stateAliases, "name", name) {
					filterJson, err := json.Marshal(config.(map[string]any)["filter"])
					if err != nil {
						return fmt.Errorf("error marshaling filter: %s", err)
					}
					aliasList = append(aliasList, map[string]any{
						"name":   name,
						"filter": string(filterJson),
					})
				}
			}

			templateVar["alias"] = aliasList
		}
	}

	// Set the entire template in the Terraform state
	if err := d.Set("template", []interface{}{templateVar}); err != nil {
		return fmt.Errorf("error setting template in state: %s", err)
	}

	return nil
}

func isInMap(list []any, key string, value any) bool {
	for _, element := range list {
		element := element.(map[string]any)
		if element[key] == value {
			return true
		}
	}
	return false
}

func resourceelkAliasesIndexUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceelkAliasesIndexCreate(d, m)
}

func resourceelkAliasesIndexDelete(d *schema.ResourceData, m interface{}) error {
	es := m.(*elasticsearch.Client)

	name := d.Id()

	// Perform the delete operation
	res, err := es.Indices.DeleteIndexTemplate(name)
	if err != nil {
		return fmt.Errorf("error deleting Elasticsearch index template: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch when deleting index template: %s", res.String())
	}

	// If delete is successful, remove it from the state
	d.SetId("")

	return nil
}

func resourceelkAliasesIndexImport(d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
	err := resourceelkAliasesIndexRead(d, m)
	if err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

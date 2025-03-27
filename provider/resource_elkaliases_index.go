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
							Type: schema.TypeList,
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
							Optional: true,
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
		if aliases, ok := templateContent["aliases"].(map[string]interface{}); ok {
			templateAliases := d.Get("template.0.alias").([]any)

			var aliasList []any
			for _, alias := range templateAliases {
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
			templateVar["alias"] = aliasList
		}
	}

	// Set the entire template in the Terraform state
	if err := d.Set("template", []interface{}{templateVar}); err != nil {
		return fmt.Errorf("error setting template in state: %s", err)
	}

	return nil
}

func resourceelkAliasesIndexUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceelkAliasesIndexCreate(d, m)
}

func resourceelkAliasesIndexDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

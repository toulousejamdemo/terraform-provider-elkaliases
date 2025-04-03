package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type dataStream struct {
	AllowCustomRouting bool
	Hidden             bool
}

func TestAccElkaliasesIndex_basic(t *testing.T) {
	resourceName := "elkaliases_index.test"
	config := `
	resource "elkaliases_index" "test" {
		name = "test_index"
		index_patterns = ["test_index"]
		template {
			mappings = jsonencode({
				"_source" = {
					mode = "synthetic"
				}
			})
			settings = jsonencode({
				index = {
					codec = "default"
				}
			})
		}
		composed_of = []
	}`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckElkaliasesIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElkaliasesIndexExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test_index"),
					resource.TestCheckResourceAttr(resourceName, "index_patterns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "index_patterns.0", "test_index"),
				),
			},
		},
	})
}

func TestAccElkaliasesIndex_aliases(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		resourceName := "elkaliases_index.test_alias"
		config := `
		resource "elkaliases_index" "test_alias" {
			name = "test_index_alias"
			index_patterns = ["test_index_alias"]
			template {
				mappings = jsonencode({
					"_source" = {
						mode = "synthetic"
					}
				})
				settings = jsonencode({
					index = {
						codec = "default"
					}
				})
				alias {
					name = "test_alias1"
					filter = jsonencode({
						"match" = {
							"event.dataset" = "test1"
						}
					})
				}
			}
			composed_of = []
		}`

		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckElkaliasesIndexDestroy,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckElkaliasesIndexExists(resourceName),
						testAccCheckElkaliasesIndexAliasExists(resourceName, "test_alias1"),
						resource.TestCheckResourceAttr(resourceName, "template.0.alias.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "template.0.alias.0.name", "test_alias1"),
					),
				},
			},
		})
	})

	t.Run("alias redifinition", func(t *testing.T) {
		resourceName := "elkaliases_index.test_alias2"
		config := `
		resource "elkaliases_index" "test_alias2" {
			name = "test_index_alias2"
			index_patterns = ["test_index_alias2"]
			template {
				mappings = jsonencode({
					"_source" = {
						mode = "synthetic"
					}
				})
				settings = jsonencode({
					index = {
						codec = "default"
					}
				})
				alias {
					name = "test_alias1"
					filter = jsonencode({
						"match" = {
							"event.dataset" = "test1"
						}
					})
				}
			}
			composed_of = []
		}`

		config2 := `
		resource "elkaliases_index" "test_alias2" {
			name = "test_index_alias2"
			index_patterns = ["test_index_alias2"]
			template {
				mappings = jsonencode({
					"_source" = {
						mode = "synthetic"
					}
				})
				settings = jsonencode({
					index = {
						codec = "default"
					}
				})
				alias {
					name = "test_alias1"
					filter = jsonencode({
						"match" = {
							"event.dataset" = "test1"
						}
					})
				}
				alias {
					name = "test_alias2"
					filter = jsonencode({
						"match" = {
							"event.dataset" = "test1"
						}
					})
				}
			}
			composed_of = []
		}`

		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckElkaliasesIndexDestroy,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckElkaliasesIndexExists(resourceName),
						testAccCheckElkaliasesIndexAliasExists(resourceName, "test_alias1"),
						resource.TestCheckResourceAttr(resourceName, "template.0.alias.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "template.0.alias.0.name", "test_alias1"),
					),
				},
				{
					Config: config2,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckElkaliasesIndexExists(resourceName),
						testAccCheckElkaliasesIndexAliasExists(resourceName, "test_alias1"),
						testAccCheckElkaliasesIndexAliasExists(resourceName, "test_alias2"),
						resource.TestCheckResourceAttr(resourceName, "template.0.alias.#", "2"),
					),
				},
			},
		})
	})
}

func TestAccElkaliasesIndex_datastream(t *testing.T) {
	t.Run("apply data_stream", func(t *testing.T) {
		resourceName := "elkaliases_index.test_ds"
		config := `
		resource "elkaliases_index" "test_ds" {
			name = "test_index_ds"
			index_patterns = ["test_index_ds"]
			template {
				mappings = jsonencode({
					"_source" = {
						mode = "synthetic"
					}
				})
				settings = jsonencode({
					index = {
						codec = "default"
					}
				})
				alias {
					name = "test_alias1"
					filter = jsonencode({
						"match" = {
							"event.dataset" = "test1"
						}
					})
				}
			}
			data_stream {
				allow_custom_routing = true
				hidden = true
			}
			composed_of = []
		}`
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckElkaliasesIndexDestroy,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckElkaliasesIndexExists(resourceName),
						testAccCheckElkaliasesIndexDataStreamExists(resourceName, dataStream{
							AllowCustomRouting: true,
							Hidden:             true,
						}),
						resource.TestCheckResourceAttr(resourceName, "data_stream.0.hidden", "true"),
						resource.TestCheckResourceAttr(resourceName, "data_stream.0.allow_custom_routing", "true"),
					),
				},
			},
		})
	})
	t.Run("add data_stream", func(t *testing.T) {
		resourceName := "elkaliases_index.test_ds"
		configWithoutDataStream := `
		resource "elkaliases_index" "test_ds" {
			name = "test_index_ds"
			index_patterns = ["test_index_ds"]
			template {
				mappings = jsonencode({
					"_source" = {
						mode = "synthetic"
					}
				})
				settings = jsonencode({
					index = {
						codec = "default"
					}
				})
				alias {
					name = "test_alias1"
					filter = jsonencode({
						"match" = {
							"event.dataset" = "test1"
						}
					})
				}
			}
			composed_of = []
		}`
		config := `
		resource "elkaliases_index" "test_ds" {
			name = "test_index_ds"
			index_patterns = ["test_index_ds"]
			template {
				mappings = jsonencode({
					"_source" = {
						mode = "synthetic"
					}
				})
				settings = jsonencode({
					index = {
						codec = "default"
					}
				})
				alias {
					name = "test_alias1"
					filter = jsonencode({
						"match" = {
							"event.dataset" = "test1"
						}
					})
				}
			}
			data_stream {
				allow_custom_routing = true
				hidden = true
			}
			composed_of = []
		}`
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckElkaliasesIndexDestroy,
			Steps: []resource.TestStep{
				{
					Config: configWithoutDataStream,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckElkaliasesIndexExists(resourceName),
						testAccCheckElkaliasesIndexDataStreamNotExists(resourceName),
						resource.TestCheckNoResourceAttr(resourceName, "data_stream.0"),
					),
				},
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckElkaliasesIndexExists(resourceName),
						testAccCheckElkaliasesIndexDataStreamExists(resourceName, dataStream{
							AllowCustomRouting: true,
							Hidden:             true,
						}),
						resource.TestCheckResourceAttr(resourceName, "data_stream.0.hidden", "true"),
						resource.TestCheckResourceAttr(resourceName, "data_stream.0.allow_custom_routing", "true"),
					),
				},
			},
		})
	})
	t.Run("remove data_stream", func(t *testing.T) {
		resourceName := "elkaliases_index.test_ds"
		configWithoutDataStream := `
		resource "elkaliases_index" "test_ds" {
			name = "test_index_ds"
			index_patterns = ["test_index_ds"]
			template {
				mappings = jsonencode({
					"_source" = {
						mode = "synthetic"
					}
				})
				settings = jsonencode({
					index = {
						codec = "default"
					}
				})
				alias {
					name = "test_alias1"
					filter = jsonencode({
						"match" = {
							"event.dataset" = "test1"
						}
					})
				}
			}
			composed_of = []
		}`
		config := `
		resource "elkaliases_index" "test_ds" {
			name = "test_index_ds"
			index_patterns = ["test_index_ds"]
			template {
				mappings = jsonencode({
					"_source" = {
						mode = "synthetic"
					}
				})
				settings = jsonencode({
					index = {
						codec = "default"
					}
				})
				alias {
					name = "test_alias1"
					filter = jsonencode({
						"match" = {
							"event.dataset" = "test1"
						}
					})
				}
			}
			data_stream {
				allow_custom_routing = true
				hidden = true
			}
			composed_of = []
		}`
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckElkaliasesIndexDestroy,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckElkaliasesIndexExists(resourceName),
						testAccCheckElkaliasesIndexDataStreamExists(resourceName, dataStream{
							AllowCustomRouting: true,
							Hidden:             true,
						}),
						resource.TestCheckResourceAttr(resourceName, "data_stream.0.hidden", "true"),
						resource.TestCheckResourceAttr(resourceName, "data_stream.0.allow_custom_routing", "true"),
					),
				},
				{
					Config: configWithoutDataStream,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckElkaliasesIndexExists(resourceName),
						testAccCheckElkaliasesIndexDataStreamNotExists(resourceName),
						resource.TestCheckNoResourceAttr(resourceName, "data_stream.0"),
					),
				},
			},
		})
	})
}

func TestAccElkaliasesIndex_invalid(t *testing.T) {
	t.Run("multiple template", func(t *testing.T) {
		config := `
		resource "elkaliases_index" "invalid" {
			name = "test_index_alias"
			index_patterns = ["test_index_alias"]
			template {
				mappings = jsonencode({
					"_source" = {
						mode = "synthetic"
					}
				})
				settings = jsonencode({
					index = {
						codec = "default"
					}
				})
				alias {
					name = "test_alias1"
					filter = jsonencode({
						"match" = {
							"event.dataset" = "test1"
						}
					})
				}
			}
			template {
				mappings = jsonencode({
					"_source" = {
						mode = "synthetic"
					}
				})
				settings = jsonencode({
					index = {
						codec = "default"
					}
				})
				alias {
					name = "test_alias1"
					filter = jsonencode({
						"match" = {
							"event.dataset" = "test1"
						}
					})
				}
			}
			composed_of = []
		}`

		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckElkaliasesIndexDestroy,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile("Too many template"),
				},
			},
		})
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ELASTICSEARCH_ENDPOINT"); v == "" {
		t.Fatal("ELKALIASES_URL must be set for acceptance tests")
	}
	if v := os.Getenv("ELASTICSEARCH_API_KEY"); v == "" {
		t.Fatal("ELKALIASES_TOKEN must be set for acceptance tests")
	}
}

func getIndexTemplate(id string) (map[string]any, error) {
	client := testAccProvider.Meta().(*elasticsearch.Client)
	res, err := client.Indices.GetIndexTemplate(client.Indices.GetIndexTemplate.WithName(id))
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 404 {
		return nil, fmt.Errorf("Index template not found")
	}

	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding Elasticsearch response: %s", err)
	}

	templates, ok := response["index_templates"].([]interface{})
	if !ok || len(templates) == 0 {
		return nil, fmt.Errorf("No templates found in response")
	}

	template := templates[0].(map[string]interface{})
	templateDetails := template["index_template"].(map[string]interface{})
	return templateDetails, nil
}

func testAccCheckElkaliasesIndexExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		_, err := getIndexTemplate(rs.Primary.ID)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckElkaliasesIndexAliasExists(resourceName string, aliasName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		response, err := getIndexTemplate(rs.Primary.ID)
		if err != nil {
			return err
		}

		if templateContent, ok := response["template"].(map[string]interface{}); ok {
			if aliases, ok := templateContent["aliases"].(map[string]interface{}); ok {
				if _, exists := aliases[aliasName]; !exists {
					return fmt.Errorf("Alias %s not found in index template", aliasName)
				}
			} else {
				return fmt.Errorf("No aliases found in index template")
			}
		} else {
			return fmt.Errorf("No template content found in index template")
		}

		return nil
	}
}

func testAccCheckElkaliasesIndexDataStreamNotExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		response, err := getIndexTemplate(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, ok = response["data_stream"].(map[string]any)
		if ok {
			return fmt.Errorf("data_stream found in %s", resourceName)
		}

		return nil
	}
}

func testAccCheckElkaliasesIndexDataStreamExists(resourceName string, ds dataStream) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		response, err := getIndexTemplate(rs.Primary.ID)
		if err != nil {
			return err
		}

		dataStream, ok := response["data_stream"].(map[string]any)
		if !ok {
			return fmt.Errorf("No data_stream found in %s", resourceName)
		}

		if ds.AllowCustomRouting != dataStream["allow_custom_routing"].(bool) {
			return fmt.Errorf("`allow_custom_routing` doesn't have the proper value in %s", resourceName)
		}

		if ds.Hidden != dataStream["hidden"].(bool) {
			return fmt.Errorf("`hidden` doesn't have the proper value in %s", resourceName)
		}

		return nil
	}
}

func testAccCheckElkaliasesIndexDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*elasticsearch.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elkaliases_index" {
			continue
		}

		// Check if the index still exists
		res, err := client.Indices.GetIndexTemplate(client.Indices.GetIndexTemplate.WithName(rs.Primary.ID))
		if err == nil {
			if res.StatusCode != 404 {
				return fmt.Errorf("Index template %s still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccElkaliasesIndexAliases_basic(t *testing.T) {
	t.Run("simple alias", func(t *testing.T) {
		resourceName := "elkaliases_index_aliases.name"
		config := `
resource "elkaliases_index_aliases" "name" {
	index = "index"
	alias {
		name = "al"
	}
}`
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAliasNotExist([]string{"al"}),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "index", "index"),
						resource.TestCheckResourceAttr(resourceName, "alias.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "alias.0.name", "al"),
						testAccCheckAliasExist("al"),
					),
				},
			},
		})
	})

	t.Run("multiple aliases", func(t *testing.T) {
		resourceName := "elkaliases_index_aliases.name"
		config := `
resource "elkaliases_index_aliases" "name" {
	index = "index"
	alias {
		name = "al"
	}
	alias {
		name = "la"
	}
}`
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAliasNotExist([]string{"al", "la"}),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "index", "index"),
						resource.TestCheckResourceAttr(resourceName, "alias.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "alias.0.name", "al"),
						resource.TestCheckResourceAttr(resourceName, "alias.1.name", "la"),
						testAccCheckAliasExist("al"),
						testAccCheckAliasExist("la"),
					),
				},
			},
		})
	})
}

func TestAccElkaliasesIndexAliases_invalid(t *testing.T) {
	t.Run("no alias", func(t *testing.T) {
		config := `resource "elkaliases_index_aliases" "name" {}`

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile("Insufficient alias blocks"),
				},
			},
		})
	})
	t.Run("empty alias", func(t *testing.T) {
		config := `
resource "elkaliases_index_aliases" "name" {
	alias {}
}`
		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile("Missing required argument"),
				},
			},
		})
	})
	t.Run("missing name", func(t *testing.T) {
		config := `
resource "elkaliases_index_aliases" "name" {
	index = "index"
	alias {
	}
}`
		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile("Missing required argument"),
				},
			},
		})
	})
	t.Run("missing index", func(t *testing.T) {
		config := `
resource "elkaliases_index_aliases" "name" {
	alias {
		name = "na"
	}
}`
		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile("Missing required argument"),
				},
			},
		})
	})
}

func TestAccElkaliasesIndexAliases_add(t *testing.T) {
	config := `
resource "elkaliases_index_aliases" "name" {
	index = "index"
	alias {
		name = "na"
	}
}`
	configAdd := `
resource "elkaliases_index_aliases" "name" {
	index = "index"
	alias {
		name = "na"
	}
	alias {
		name = "an"
	}
}`
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExist("na"),
				),
			},
			{
				Config: configAdd,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExist("na"),
					testAccCheckAliasExist("an"),
				),
			},
		},
	})
}

func TestAccElkaliasesIndexAliases_remove(t *testing.T) {
	config := `
resource "elkaliases_index_aliases" "name" {
	index = "index"
	alias {
		name = "an"
	}
	alias {
		name = "na"
	}
}`
	configAdd := `
resource "elkaliases_index_aliases" "name" {
	index = "index"
	alias {
		name = "an"
	}
}`
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExist("an"),
					testAccCheckAliasExist("na"),
				),
			},
			{
				Config: configAdd,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExist("an"),
				),
			},
		},
	})
}

func testAccCheckAliasExist(aliasName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*elasticsearch.Client)
		res, err := client.Indices.GetAlias(client.Indices.GetAlias.WithName(aliasName))
		if err != nil {
			return err
		}

		if res.StatusCode == 404 {
			return fmt.Errorf("Alias %s does not exist", aliasName)
		}

		return nil
	}
}

func testAccCheckAliasNotExist(aliasesName []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*elasticsearch.Client)
		for _, aliasName := range aliasesName {
			res, err := client.Indices.GetAlias(client.Indices.GetAlias.WithName(aliasName))
			if err != nil {
				return err
			}

			if res.StatusCode != 404 {
				return fmt.Errorf("Alias %s exists", aliasName)
			}
		}
		return nil
	}
}

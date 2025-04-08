package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"elkaliases": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ELASTICSEARCH_ENDPOINT"); v == "" {
		t.Fatal("ELKALIASES_URL must be set for acceptance tests")
	}
	if v := os.Getenv("ELASTICSEARCH_API_KEY"); v == "" {
		t.Fatal("ELKALIASES_TOKEN must be set for acceptance tests")
	}
}

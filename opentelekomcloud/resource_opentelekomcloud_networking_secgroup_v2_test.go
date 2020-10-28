package opentelekomcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/extensions/security/groups"
)

func TestAccNetworkingV2SecGroup_basic(t *testing.T) {
	var securityGroup groups.SecGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkingV2SecGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingV2SecGroup_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingV2SecGroupExists("opentelekomcloud_networking_secgroup_v2.secgroup_1", &securityGroup),
					testAccCheckNetworkingV2SecGroupRuleCount(&securityGroup, 2),
				),
			},
			{
				Config: testAccNetworkingV2SecGroup_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr("opentelekomcloud_networking_secgroup_v2.secgroup_1", "id", &securityGroup.ID),
					resource.TestCheckResourceAttr("opentelekomcloud_networking_secgroup_v2.secgroup_1", "name", "security_group_2"),
				),
			},
		},
	})
}

func TestAccNetworkingV2SecGroup_noDefaultRules(t *testing.T) {
	var securityGroup groups.SecGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkingV2SecGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingV2SecGroup_noDefaultRules,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingV2SecGroupExists("opentelekomcloud_networking_secgroup_v2.secgroup_1", &securityGroup),
					testAccCheckNetworkingV2SecGroupRuleCount(&securityGroup, 0),
				),
			},
		},
	})
}

func TestAccNetworkingV2SecGroup_timeout(t *testing.T) {
	var securityGroup groups.SecGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkingV2SecGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingV2SecGroup_timeout,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingV2SecGroupExists("opentelekomcloud_networking_secgroup_v2.secgroup_1", &securityGroup),
				),
			},
		},
	})
}

func testAccCheckNetworkingV2SecGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	networkingClient, err := config.networkingV2Client(OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud Networkingv2 client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_networking_secgroup_v2" {
			continue
		}

		_, err := groups.Get(networkingClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("security group still exists")
		}
	}

	return nil
}

func testAccCheckNetworkingV2SecGroupExists(n string, securityGroup *groups.SecGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		networkingClient, err := config.networkingV2Client(OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("error creating OpenTelekomCloud networking client: %s", err)
		}

		found, err := groups.Get(networkingClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("security group not found")
		}

		*securityGroup = *found

		return nil
	}
}

func testAccCheckNetworkingV2SecGroupRuleCount(
	sg *groups.SecGroup, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(sg.Rules) == count {
			return nil
		}

		return fmt.Errorf("unexpected number of rules in group %s. Expected %d, got %d",
			sg.ID, count, len(sg.Rules))
	}
}

const testAccNetworkingV2SecGroup_basic = `
resource "opentelekomcloud_networking_secgroup_v2" "secgroup_1" {
  name        = "security_group"
  description = "terraform security group acceptance test"
}
`

const testAccNetworkingV2SecGroup_update = `
resource "opentelekomcloud_networking_secgroup_v2" "secgroup_1" {
  name        = "security_group_2"
  description = "terraform security group acceptance test"
}
`

const testAccNetworkingV2SecGroup_noDefaultRules = `
resource "opentelekomcloud_networking_secgroup_v2" "secgroup_1" {
  name                 = "security_group_1"
  description          = "terraform security group acceptance test"
  delete_default_rules = true
}
`

const testAccNetworkingV2SecGroup_timeout = `
resource "opentelekomcloud_networking_secgroup_v2" "secgroup_1" {
  name        = "security_group"
  description = "terraform security group acceptance test"

  timeouts {
    delete = "5m"
  }
}
`

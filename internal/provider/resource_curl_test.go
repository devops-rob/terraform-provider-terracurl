package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccresourceCurl(t *testing.T) {

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurl,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"curl_request.hashicorp_products", "response",
						`["atlas-upload-cli","boundary","boundary-desktop","consul","consul-api-gateway","consul-aws","consul-ecs","consul-esm","consul-k8s","consul-k8s-control-plane","consul-lambda-registrator","consul-replicate","consul-template","consul-terraform-sync","crt-core-helloworld","docker-base","docker-basetool","envconsul","hcdiag","hcs","levant","nomad","nomad-autoscaler","nomad-device-nvidia","nomad-driver-ecs","nomad-driver-lxc","nomad-driver-podman","nomad-pack","otto","packer","sentinel","serf","terraform","terraform-ls","terraform-provider-aci","terraform-provider-acme","terraform-provider-ad","terraform-provider-akamai","terraform-provider-alicloud","terraform-provider-archive","terraform-provider-arukas","terraform-provider-atlas","terraform-provider-auth0","terraform-provider-avi","terraform-provider-aviatrix","terraform-provider-aws","terraform-provider-awscc","terraform-provider-azure","terraform-provider-azuread","terraform-provider-azuredevops","terraform-provider-azurerm","terraform-provider-azurestack","terraform-provider-baiducloud","terraform-provider-bigip","terraform-provider-bitbucket","terraform-provider-boundary","terraform-provider-brightbox","terraform-provider-checkpoint","terraform-provider-chef","terraform-provider-cherryservers","terraform-provider-circonus","terraform-provider-ciscoasa","terraform-provider-clc","terraform-provider-cloudamqp","terraform-provider-cloudflare","terraform-provider-cloudinit","terraform-provider-cloudscale","terraform-provider-cloudstack","terraform-provider-cobbler","terraform-provider-cohesity","terraform-provider-constellix","terraform-provider-consul","terraform-provider-datadog","terraform-provider-digitalocean","terraform-provider-dme","terraform-provider-dns","terraform-provider-dnsimple","terraform-provider-docker","terraform-provider-dome9","terraform-provider-dyn","terraform-provider-ecl","terraform-provider-equinix","terraform-provider-exoscale","terraform-provider-external","terraform-provider-fakewebservices","terraform-provider-fastly","terraform-provider-flexibleengine","terraform-provider-fortios","terraform-provider-genymotion","terraform-provider-github","terraform-provider-gitlab","terraform-provider-google","terraform-provider-google-beta","terraform-provider-googleworkspace","terraform-provider-grafana","terraform-provider-gridscale","terraform-provider-hcloud","terraform-provider-hcp","terraform-provider-hcs","terraform-provider-hedvig","terraform-provider-helm","terraform-provider-heroku","terraform-provider-http","terraform-provider-huaweicloud","terraform-provider-huaweicloudstack","terraform-provider-icinga2","terraform-provider-ignition","terraform-provider-incapsula","terraform-provider-influxdb","terraform-provider-infoblox","terraform-provider-jdcloud","terraform-provider-ksyun","terraform-provider-kubernetes","terraform-provider-kubernetes-alpha","terraform-provider-lacework","terraform-provider-launchdarkly","terraform-provider-librato","terraform-provider-linode","terraform-provider-local","terraform-provider-logentries","terraform-provider-logicmonitor","terraform-provider-mailgun","terraform-provider-metalcloud","terraform-provider-mongodbatlas","terraform-provider-mso","terraform-provider-mysql","terraform-provider-ncloud","terraform-provider-netlify","terraform-provider-newrelic","terraform-provider-nomad","terraform-provider-ns1","terraform-provider-nsxt","terraform-provider-null","terraform-provider-nutanix","terraform-provider-oci","terraform-provider-okta","terraform-provider-oktaasa","terraform-provider-oneandone","terraform-provider-onelogin","terraform-provider-opc","terraform-provider-opennebula","terraform-provider-openstack","terraform-provider-opentelekomcloud","terraform-provider-opsgenie","terraform-provider-oraclepaas","terraform-provider-ovh","terraform-provider-packet","terraform-provider-pagerduty","terraform-provider-panos","terraform-provider-postgresql","terraform-provider-powerdns","terraform-provider-prismacloud","terraform-provider-profitbricks","terraform-provider-pureport","terraform-provider-rabbitmq","terraform-provider-rancher","terraform-provider-rancher2","terraform-provider-random","terraform-provider-rightscale","terraform-provider-rubrik","terraform-provider-rundeck","terraform-provider-runscope","terraform-provider-salesforce","terraform-provider-scaleway","terraform-provider-sdm","terraform-provider-selectel","terraform-provider-selvpc","terraform-provider-signalfx","terraform-provider-skytap","terraform-provider-softlayer","terraform-provider-spotinst","terraform-provider-stackpath","terraform-provider-statuscake","terraform-provider-sumologic","terraform-provider-telefonicaopencloud","terraform-provider-template","terraform-provider-tencentcloud","terraform-provider-terraform","terraform-provider-tfe","terraform-provider-thunder","terraform-provider-time","terraform-provider-tls","terraform-provider-triton","terraform-provider-turbot","terraform-provider-ucloud","terraform-provider-ultradns","terraform-provider-vault","terraform-provider-vcd","terraform-provider-venafi","terraform-provider-vmc","terraform-provider-vra","terraform-provider-vra7","terraform-provider-vsphere","terraform-provider-vthunder","terraform-provider-vultr","terraform-provider-wavefront","terraform-provider-yandex","tfc-agent","vagrant","vagrant-vmware-utility","vault","vault-auditor","vault-csi-provider","vault-k8s","vault-lambda-extension","vault-mssql-ekm-provider","vault-plugin-database-oracle","vault-servicenow-credential-resolver","vault-ssh-helper","waypoint","waypoint-entrypoint"]`),
				),
			},
		},
	})
}

const testAccresourceCurl = `
resource "curl_request" "hashicorp_products" {
  name           = "releases"
  url            = "https://api.releases.hashicorp.com/v1/products"
  method         = "GET"
  destroy_url    = "https://api.releases.hashicorp.com/v1/products"
  destroy_method = ""
}
`

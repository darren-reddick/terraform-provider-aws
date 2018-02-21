package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"log"
)

func resourceAwsOrganization() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsOrganizationCreate,
		Read:   resourceAwsOrganizationRead,
		Delete: resourceAwsOrganizationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"feature_set": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      organizations.OrganizationFeatureSetAll,
				ValidateFunc: validation.StringInSlice([]string{organizations.OrganizationFeatureSetAll, organizations.OrganizationFeatureSetConsolidatedBilling}, true),
			},
		},
	}
}

func resourceAwsOrganizationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	createOpts := &organizations.CreateOrganizationInput{
		FeatureSet: aws.String(d.Get("feature_set").(string)),
	}
	log.Printf("[DEBUG] Creating Organization: %#v", createOpts)

	resp, err := conn.CreateOrganization(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating organization: %s", err)
	}

	org := resp.Organization
	d.SetId(*org.Id)
	log.Printf("[INFO] Organization ID: %s", d.Id())

	return resourceAwsOrganizationRead(d, meta)
}

func resourceAwsOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	log.Printf("[INFO] Reading Organization: %s", d.Id())
	org, err := conn.DescribeOrganization(&organizations.DescribeOrganizationInput{})
	if err != nil {
		if orgerr, ok := err.(awserr.Error); ok && orgerr.Code() == "AWSOrganizationsNotInUseException" {
			log.Printf("[WARN] Organization does not exist, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("arn", org.Organization.Arn)
	d.Set("feature_set", org.Organization.FeatureSet)
	d.Set("master_account_arn", org.Organization.MasterAccountArn)
	d.Set("master_account_email", org.Organization.MasterAccountEmail)
	d.Set("master_account_id", org.Organization.MasterAccountId)
	return nil
}

func resourceAwsOrganizationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	log.Printf("[INFO] Deleting Organization: %s", d.Id())

	_, err := conn.DeleteOrganization(&organizations.DeleteOrganizationInput{})
	if err != nil {
		return fmt.Errorf("Error deleting Organization: %s", err)
	}

	d.SetId("")

	return nil
}

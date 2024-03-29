package dnsimple

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDNSimpleRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceDNSimpleRecordCreate,
		Read:   resourceDNSimpleRecordRead,
		Update: resourceDNSimpleRecordUpdate,
		Delete: resourceDNSimpleRecordDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDNSimpleRecordImport,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"domain_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"value": {
				Type:     schema.TypeString,
				Required: true,
			},

			"regions": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},

			"ttl": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "3600",
			},

			"priority": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func resourceDNSimpleRecordCreate(d *schema.ResourceData, meta interface{}) error {
	provider := meta.(*Client)
	name := d.Get("name").(string)
	domain := d.Get("domain").(string)

	regionsRaw := d.Get("regions").(*schema.Set).List()
	regions := make([]string, len(regionsRaw))
	for i, raw := range regionsRaw {
		regions[i] = raw.(string)
	}
	log.Printf("DNSimple: Region %#v", regions)

	// Create the new record
	newRecord := dnsimple.ZoneRecordAttributes{
		ZoneID:  domain,
		Type:    d.Get("type").(string),
		Name:    &name,
		Content: d.Get("value").(string),
		Regions: regions,
	}

	if attr, ok := d.GetOk("ttl"); ok {
		newRecord.TTL, _ = strconv.Atoi(attr.(string))
	}

	if attr, ok := d.GetOk("priority"); ok {
		newRecord.Priority, _ = strconv.Atoi(attr.(string))
	}

	log.Printf("[DEBUG] DNSimple Record create configuration: %#v", newRecord)

	resp, err := provider.client.Zones.CreateRecord(context.Background(), provider.config.Account, domain, newRecord)
	if err != nil {
		return fmt.Errorf("Failed to create DNSimple Record: %s", err)
	}

	d.SetId(strconv.FormatInt(resp.Data.ID, 10))
	log.Printf("[INFO] DNSimple Record ID: %s", d.Id())
	return resourceDNSimpleRecordRead(d, meta)
}

func resourceDNSimpleRecordRead(d *schema.ResourceData, meta interface{}) error {
	provider := meta.(*Client)

	recordID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error converting Record ID: %s", err)
	}
	log.Printf("[INFO] DNSimple Record ID: %s", d.Id())

	resp, err := provider.client.Zones.GetRecord(context.Background(), provider.config.Account, d.Get("domain").(string), int64(recordID))
	if err != nil {
		if err != nil && strings.Contains(err.Error(), "404") {
			log.Printf("DNSimple Record Not Found - Refreshing from State")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Couldn't find DNSimple Record: %s", err)
	}
	log.Printf("[INFO] DNSimple raw responce: %s", resp)

	record := resp.Data
	d.Set("domain_id", record.ZoneID)
	d.Set("name", record.Name)
	d.Set("type", record.Type)
	d.Set("value", record.Content)
	d.Set("regions", record.Regions)
	d.Set("ttl", strconv.Itoa(record.TTL))
	d.Set("priority", strconv.Itoa(record.Priority))

	if record.Name == "" {
		d.Set("hostname", d.Get("domain").(string))
	} else {
		d.Set("hostname", fmt.Sprintf("%s.%s", record.Name, d.Get("domain").(string)))
	}
	return nil
}

func resourceDNSimpleRecordUpdate(d *schema.ResourceData, meta interface{}) error {
	provider := meta.(*Client)

	recordID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error converting Record ID: %s", err)
	}

	regionsRaw := d.Get("regions").(*schema.Set).List()
	regions := make([]string, len(regionsRaw))
	for i, raw := range regionsRaw {
		regions[i] = raw.(string)
	}
	log.Printf("DNSimple: Region %#v", regions)

	updateRecord := dnsimple.ZoneRecordAttributes{}

	if attr, ok := d.GetOk("name"); ok {
		updateRecord.ZoneID = attr.(string)
	}
	if attr, ok := d.GetOk("type"); ok {
		updateRecord.Type = attr.(string)
	}
	if attr, ok := d.GetOk("value"); ok {
		updateRecord.Content = attr.(string)
	}

	updateRecord.Regions = regions

	if attr, ok := d.GetOk("ttl"); ok {
		updateRecord.TTL, _ = strconv.Atoi(attr.(string))
	}

	if attr, ok := d.GetOk("priority"); ok {
		updateRecord.Priority, _ = strconv.Atoi(attr.(string))
	}

	log.Printf("[DEBUG] DNSimple Record update configuration: %#v", updateRecord)

	_, err = provider.client.Zones.UpdateRecord(context.Background(), provider.config.Account, d.Get("domain").(string), int64(recordID), updateRecord)
	if err != nil {
		return fmt.Errorf("Failed to update DNSimple Record: %s", err)
	}

	return resourceDNSimpleRecordRead(d, meta)
}

func resourceDNSimpleRecordDelete(d *schema.ResourceData, meta interface{}) error {
	provider := meta.(*Client)

	log.Printf("[INFO] Deleting DNSimple Record: %s, %s", d.Get("domain").(string), d.Id())

	recordID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error converting Record ID: %s", err)
	}

	_, err = provider.client.Zones.DeleteRecord(context.Background(), provider.config.Account, d.Get("domain").(string), int64(recordID))
	if err != nil {
		return fmt.Errorf("Error deleting DNSimple Record: %s", err)
	}
	return nil
}

func resourceDNSimpleRecordImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "_")

	if len(parts) != 2 {
		return nil, fmt.Errorf("Error Importing dnsimple_record. Please make sure the record ID is in the form DOMAIN_RECORDID (i.e. example.com_1234")
	}

	d.SetId(parts[1])
	d.Set("domain", parts[0])

	if err := resourceDNSimpleRecordRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

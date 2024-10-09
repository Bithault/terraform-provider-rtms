package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
	})
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"auth_token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("RTMS_AUTH_TOKEN", nil),
				Description: "The X-AUTH-TOKEN for API authentication",
			},
			"cloud_temple_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("RTMS_CLOUD_TEMPLE_ID", nil),
				Description: "The cloudTempleId for API calls",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"rtms_appliance":   dataSourceRtmsAppliance(),
			"rtms_plugin":      dataSourceRtmsPlugin(),
			"rtms_template":    dataSourceRtmsTemplate(),
			"rtms_typology":    dataSourceRtmsTypology(),
			"rtms_team":        dataSourceRtmsTeam(),
			"rtms_checkperiod": dataSourceRtmsCheckPeriod(),
			"rtms_timeperiod":  dataSourceRtmsTimePeriod(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"rtms_host":               resourceHost(),
			"rtms_monitoring_service": resourceMonitoringService(),
		},
		ConfigureFunc: providerConfigure,
	}
}

type apiClient struct {
	authToken     string
	cloudTempleId string
	httpClient    *http.Client
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	return &apiClient{
		authToken:     d.Get("auth_token").(string),
		cloudTempleId: d.Get("cloud_temple_id").(string),
		httpClient:    &http.Client{},
	}, nil
}

type ValidationError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Errors  struct {
		Children map[string]struct {
			Errors []string `json:"errors"`
		} `json:"children"`
	} `json:"errors"`
}

func formatAPIError(resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("API request error. Status Code: %d. Error reading body: %s", resp.StatusCode, err)
	}

	var validationError ValidationError
	if err := json.Unmarshal(body, &validationError); err == nil && validationError.Code == 400 {
		var errorMessages []string
		for field, errorData := range validationError.Errors.Children {
			for _, errorMsg := range errorData.Errors {
				errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, errorMsg))
			}
		}
		return fmt.Errorf("API Validation Error. Status Code: %d. Message: %s. Errors: %s",
			validationError.Code,
			validationError.Message,
			strings.Join(errorMessages, "; "))
	}

	// Si ce n'est pas une erreur de validation, retournons le corps brut
	return fmt.Errorf("API request error. Status Code: %d. Response: %s", resp.StatusCode, string(body))
}

func dataSourceRtmsAppliance() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRtmsApplianceRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Required: true,
			},
			"appliance": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceRtmsApplianceRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(strconv.Itoa(d.Get("id").(int)))
	return nil
}

func dataSourceRtmsPlugin() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRtmsPluginRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"isdeprecated": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func dataSourceRtmsPluginRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(strconv.Itoa(d.Get("id").(int)))
	return nil
}

func dataSourceRtmsTemplate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRtmsTemplateRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func dataSourceRtmsTemplateRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(strconv.Itoa(d.Get("id").(int)))
	return nil
}

func dataSourceRtmsTypology() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRtmsTypologyRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
	}
}

func dataSourceRtmsTypologyRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Get("id").([]interface{})
	d.SetId(fmt.Sprintf("%v", id))
	return nil
}

func dataSourceRtmsTeam() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRtmsTeamRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func dataSourceRtmsTeamRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(strconv.Itoa(d.Get("id").(int)))
	return nil
}

func dataSourceRtmsCheckPeriod() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRtmsCheckPeriodRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func dataSourceRtmsCheckPeriodRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(strconv.Itoa(d.Get("id").(int)))
	return nil
}

func dataSourceRtmsTimePeriod() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRtmsTimePeriodRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func dataSourceRtmsTimePeriodRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(strconv.Itoa(d.Get("id").(int)))
	return nil
}

func resourceHost() *schema.Resource {
	return &schema.Resource{
		Create: resourceHostCreate,
		Read:   resourceHostRead,
		Update: resourceHostUpdate,
		Delete: resourceHostDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Required: true,
			},
			"address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"community": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"admin_login": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"admin_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"appliance": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceHostCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*apiClient)

	host := map[string]interface{}{
		"name":    d.Get("name"),
		"alias":   d.Get("alias"),
		"address": d.Get("address"),
	}

	if v, ok := d.GetOk("community"); ok {
		host["community"] = v
	}
	if v, ok := d.GetOk("admin_login"); ok {
		host["adminLogin"] = v
	}
	if v, ok := d.GetOk("admin_password"); ok {
		host["adminPassword"] = v
	}
	if v, ok := d.GetOk("type"); ok {
		host["type"] = v
	}
	if v, ok := d.GetOk("appliance"); ok {
		host["appliance"] = v
	}

	jsonBody, err := json.Marshal(host)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://rtms-api.cloud-temple.com/v1/hosts?cloudTempleId=%s", client.cloudTempleId)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AUTH-TOKEN", client.authToken)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return formatAPIError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	hostId, ok := result["hostId"].(float64)
	if !ok {
		return fmt.Errorf("Unexpected response format")
	}

	d.SetId(strconv.Itoa(int(hostId)))

	return resourceHostRead(d, m)
}

func resourceHostRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*apiClient)

	url := fmt.Sprintf("https://rtms-api.cloud-temple.com/v1/hosts/%s", d.Id())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AUTH-TOKEN", client.authToken)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		d.SetId("")
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return formatAPIError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Unexpected response format")
	}

	d.Set("name", data["name"])
	d.Set("alias", data["alias"])
	d.Set("address", data["address"])
	d.Set("community", data["community"])
	d.Set("admin_login", data["adminLogin"])
	d.Set("type", data["type"])
	if appliance, ok := data["appliance"].(map[string]interface{}); ok {
		d.Set("appliance", int(appliance["id"].(float64)))
	}

	return nil
}

func resourceHostUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*apiClient)

	host := map[string]interface{}{}

	if d.HasChange("name") {
		host["name"] = d.Get("name")
	}
	if d.HasChange("alias") {
		host["alias"] = d.Get("alias")
	}
	if d.HasChange("address") {
		host["address"] = d.Get("address")
	}
	if d.HasChange("community") {
		host["community"] = d.Get("community")
	}
	if d.HasChange("admin_login") {
		host["adminLogin"] = d.Get("admin_login")
	}
	if d.HasChange("admin_password") {
		host["adminPassword"] = d.Get("admin_password")
	}
	if d.HasChange("type") {
		host["type"] = d.Get("type")
	}
	if d.HasChange("appliance") {
		host["appliance"] = d.Get("appliance")
	}

	jsonBody, err := json.Marshal(host)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://rtms-api.cloud-temple.com/v1/hosts/%s", d.Id())
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AUTH-TOKEN", client.authToken)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return formatAPIError(resp)
	}

	return resourceHostRead(d, m)
}

func resourceHostDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*apiClient)

	url := fmt.Sprintf("https://rtms-api.cloud-temple.com/v1/hosts/%s", d.Id())
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AUTH-TOKEN", client.authToken)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return formatAPIError(resp)
	}

	d.SetId("")

	return nil
}

func resourceMonitoringService() *schema.Resource {
	return &schema.Resource{
		Create: resourceMonitoringServiceCreate,
		Read:   resourceMonitoringServiceRead,
		Update: resourceMonitoringServiceUpdate,
		Delete: resourceMonitoringServiceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"appliance": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"host": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"template": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"max_check_attempts": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"plugin": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"plugin_args": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_monitored": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"notifications_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"nice_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"keywords": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"help": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"severity": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"only_notify_if_critical": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"normal_check_interval": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"retry_check_interval": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"time_period": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"check_period": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ticket_catalogs_items": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"auto_processing": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"responsible_team": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceMonitoringServiceCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*apiClient)

	service := map[string]interface{}{
		"appliance": d.Get("appliance"),
		"host":      d.Get("host"),
		"name":      d.Get("name"),
		"template":  d.Get("template"),
	}

	if v, ok := d.GetOk("description"); ok {
		service["description"] = v
	}
	if v, ok := d.GetOk("max_check_attempts"); ok {
		service["maxCheckAttempts"] = v
	}
	if v, ok := d.GetOk("plugin"); ok {
		service["plugin"] = v
	}
	if v, ok := d.GetOk("plugin_args"); ok {
		service["pluginArgs"] = v
	}
	if v, ok := d.GetOk("is_monitored"); ok {
		service["isMonitored"] = v
	}
	if v, ok := d.GetOk("notifications_enabled"); ok {
		service["notificationsEnabled"] = v
	}
	if v, ok := d.GetOk("nice_name"); ok {
		service["niceName"] = v
	}
	if v, ok := d.GetOk("keywords"); ok {
		service["keywords"] = v
	}
	if v, ok := d.GetOk("help"); ok {
		service["help"] = v
	}
	if v, ok := d.GetOk("severity"); ok {
		service["severity"] = v
	}
	if v, ok := d.GetOk("only_notify_if_critical"); ok {
		service["onlyNotifyIfCritical"] = v
	}
	if v, ok := d.GetOk("normal_check_interval"); ok {
		service["normalCheckInterval"] = v
	}
	if v, ok := d.GetOk("retry_check_interval"); ok {
		service["retryCheckInterval"] = v
	}
	if v, ok := d.GetOk("time_period"); ok {
		service["timePeriod"] = v
	}
	if v, ok := d.GetOk("check_period"); ok {
		service["checkPeriod"] = v
	}
	if v, ok := d.GetOk("ticket_catalogs_items"); ok {
		service["ticketCatalogsItems"] = v.([]interface{})
	}
	if v, ok := d.GetOk("auto_processing"); ok {
		service["autoProcessing"] = v
	}
	if v, ok := d.GetOk("responsible_team"); ok {
		service["responsibleTeam"] = v
	}

	jsonBody, err := json.Marshal(service)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://rtms-api.cloud-temple.com/v1/monitoringServices?cloudTempleId=%s", client.cloudTempleId)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AUTH-TOKEN", client.authToken)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return formatAPIError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	serviceId, ok := result["id"].(float64)
	if !ok {
		return fmt.Errorf("Unexpected response format")
	}

	d.SetId(strconv.Itoa(int(serviceId)))

	return resourceMonitoringServiceRead(d, m)
}

func resourceMonitoringServiceRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*apiClient)

	url := fmt.Sprintf("https://rtms-api.cloud-temple.com/v1/monitoringServices/%s", d.Id())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AUTH-TOKEN", client.authToken)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		d.SetId("")
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return formatAPIError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Unexpected response format")
	}

	d.Set("name", data["name"])
	d.Set("description", data["description"])
	if host, ok := data["host"].(map[string]interface{}); ok {
		d.Set("host", int(host["id"].(float64)))
	}
	if template, ok := data["template"].(map[string]interface{}); ok {
		d.Set("template", int(template["id"].(float64)))
	}
	if plugin, ok := data["plugin"].(map[string]interface{}); ok {
		d.Set("plugin", int(plugin["id"].(float64)))
	}
	if appliance, ok := data["appliance"].(map[string]interface{}); ok {
		d.Set("appliance", int(appliance["id"].(float64)))
	}
	d.Set("is_monitored", data["isMonitored"])
	d.Set("notifications_enabled", data["notificationsEnabled"])
	d.Set("nice_name", data["niceName"])
	d.Set("keywords", data["keywords"])
	d.Set("help", data["help"])
	d.Set("severity", data["severity"])
	d.Set("only_notify_if_critical", data["onlyNotifyIfCritical"])
	d.Set("normal_check_interval", data["normalCheckInterval"])
	d.Set("retry_check_interval", data["retryCheckInterval"])
	d.Set("max_check_attempts", data["maxCheckAttempts"])
	if timePeriod, ok := data["timePeriod"].(map[string]interface{}); ok {
		d.Set("time_period", int(timePeriod["id"].(float64)))
	}
	if checkPeriod, ok := data["checkPeriod"].(map[string]interface{}); ok {
		d.Set("check_period", int(checkPeriod["id"].(float64)))
	}
	if ticketCatalogsItems, ok := data["ticketCatalogsItems"].([]interface{}); ok {
		var items []int
		for _, item := range ticketCatalogsItems {
			if itemMap, ok := item.(map[string]interface{}); ok {
				items = append(items, int(itemMap["id"].(float64)))
			}
		}
		d.Set("ticket_catalogs_items", items)
	}
	d.Set("auto_processing", data["autoProcessing"])
	if responsibleTeam, ok := data["responsibleTeam"].(map[string]interface{}); ok {
		d.Set("responsible_team", int(responsibleTeam["id"].(float64)))
	}

	return nil
}

func resourceMonitoringServiceUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*apiClient)

	service := map[string]interface{}{}

	if d.HasChange("appliance") {
		service["appliance"] = d.Get("appliance")
	}
	if d.HasChange("host") {
		service["host"] = d.Get("host")
	}
	if d.HasChange("name") {
		service["name"] = d.Get("name")
	}
	if d.HasChange("template") {
		service["template"] = d.Get("template")
	}
	if d.HasChange("description") {
		service["description"] = d.Get("description")
	}
	if d.HasChange("max_check_attempts") {
		service["maxCheckAttempts"] = d.Get("max_check_attempts")
	}
	if d.HasChange("plugin") {
		service["plugin"] = d.Get("plugin")
	}
	if d.HasChange("plugin_args") {
		service["pluginArgs"] = d.Get("plugin_args")
	}
	if d.HasChange("is_monitored") {
		service["isMonitored"] = d.Get("is_monitored")
	}
	if d.HasChange("notifications_enabled") {
		service["notificationsEnabled"] = d.Get("notifications_enabled")
	}
	if d.HasChange("nice_name") {
		service["niceName"] = d.Get("nice_name")
	}
	if d.HasChange("keywords") {
		service["keywords"] = d.Get("keywords")
	}
	if d.HasChange("help") {
		service["help"] = d.Get("help")
	}
	if d.HasChange("severity") {
		service["severity"] = d.Get("severity")
	}
	if d.HasChange("only_notify_if_critical") {
		service["onlyNotifyIfCritical"] = d.Get("only_notify_if_critical")
	}
	if d.HasChange("normal_check_interval") {
		service["normalCheckInterval"] = d.Get("normal_check_interval")
	}
	if d.HasChange("retry_check_interval") {
		service["retryCheckInterval"] = d.Get("retry_check_interval")
	}
	if d.HasChange("time_period") {
		service["timePeriod"] = d.Get("time_period")
	}
	if d.HasChange("check_period") {
		service["checkPeriod"] = d.Get("check_period")
	}
	if d.HasChange("ticket_catalogs_items") {
		service["ticketCatalogsItems"] = d.Get("ticket_catalogs_items")
	}
	if d.HasChange("auto_processing") {
		service["autoProcessing"] = d.Get("auto_processing")
	}
	if d.HasChange("responsible_team") {
		service["responsibleTeam"] = d.Get("responsible_team")
	}

	jsonBody, err := json.Marshal(service)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://rtms-api.cloud-temple.com/v1/monitoringServices/%s", d.Id())
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AUTH-TOKEN", client.authToken)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return formatAPIError(resp)
	}

	return resourceMonitoringServiceRead(d, m)
}

func resourceMonitoringServiceDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*apiClient)

	url := fmt.Sprintf("https://rtms-api.cloud-temple.com/v1/monitoringServices/%s", d.Id())
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AUTH-TOKEN", client.authToken)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return formatAPIError(resp)
	}

	d.SetId("")

	return nil
}
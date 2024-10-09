# Terraform Provider RTMS

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
- [Go](https://golang.org/doc/install) >= 1.18

## Using the provider

    provider "rtms" {
      # Can also be set as the RTMS_AUTH_TOKEN environment variable
      auth_token = "your-auth-token"

      # Can also be set as the RTMS_CLOUD_TEMPLE_ID environment variable
      cloud_temple_id = "your-cloud-temple-id"
    }

- `auth_token` (String, Sensitive) The X-AUTH-TOKEN for API authentication. Can also be specified with the environment variable `RTMS_AUTH_TOKEN`.
- `cloud_temple_id` (String) The cloudTempleId for identifying current Tenant.

## Examples

### Resources

#### rtms_host
```
    resource "rtms_host" "example-host" {
      name    = "example-host"
      alias   = "Example Host"
      address = "192.168.1.100"
      community = "public"
      admin_login = "admin"
      admin_password = "password"
      type = "server"
      appliance = data.rtms_appliance.example-appliance.id
    }
```
#### rtms_monitoring_service
```
    resource "rtms_monitoring_service" "example" {
      appliance = data.rtms_appliance.example-appliance.id
      host      = rtms_host.example-host.id
      name      = "example-service"
      template  = data.rtms_template.example-template.id
      description = "Example monitoring service"
      max_check_attempts = 3
      plugin = data.rtms_plugin.example-plugin.id
      plugin_args = "-w 80 -c 90"
      is_monitored = true
      notifications_enabled = true
      nice_name = "Example Service"
      keywords = "example,service"
      help = "This is an example service"
      severity = 3
      only_notify_if_critical = false
      normal_check_interval = 300
      retry_check_interval = 60
      time_period = data.rtms_timeperiod.example-timeperiod.id
      check_period = data.rtms_checkperiod.example-checkperiod.id
      ticket_catalogs_items = data.rtms_typology.example-typology.id
      auto_processing = true
      responsible_team = data.rtms_team.example-team.id
    }
```
### Data Sources

#### rtms_appliance
```
    data "rtms_appliance" "example-appliance" {
      name      = "example-appliance"
      id        = 1
      alias     = "Example Appliance"
      appliance = "192.168.1.1"
    }
```
#### rtms_plugin
```
    data "rtms_plugin" "example-plugin" {
      name         = "example-plugin"
      id           = 1
      isdeprecated = false
    }
```
#### rtms_template
```
    data "rtms_template" "example-template" {
      name = "example-template"
      id   = 1
    }
```
#### rtms_typology
```
    data "rtms_typology" "example-typology" {
      name        = "example-typology"
      description = "Example typology description"
      id          = [1, 2, 3]
    }
```
#### rtms_team
```
    data "rtms_team" "example-team" {
      name = "example-team"
      id   = 1
    }
```
#### rtms_checkperiod
```
    data "rtms_checkperiod" "example-checkperiod" {
      name = "24x7"
      id   = 1
    }
```
#### rtms_timeperiod
```
    data "rtms_timeperiod" "example-timeperiod" {
      name = "business-hours"
      id   = 1
    }
```

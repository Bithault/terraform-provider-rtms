##### SAISIE MANUELLE #####

$APIkey = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
$CloudTempleId = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

$TFMainFilePath = "c:\Temp\main.tf"
$TFDataFilePath = "c:\Temp\common.tf"
$TFImportCmdFilePath = "c:\Temp\imports.txt"

#################################################

$APIHeaders = @{"Content-Type" = "application/json"; "accept" = "application / json"; "X-AUTH-TOKEN" = $APIKey}

New-Item $TFMainFilePath -Force
New-Item $TFDataFilePath -Force
New-Item $TFImportCmdFilePath -Force

function Remove-Diacritics {
    param (
        [Parameter(Mandatory=$true)]
        [string]$Text
    )
    $normalizedString = $Text.Normalize([Text.NormalizationForm]::FormD)
    $stringBuilder = New-Object Text.StringBuilder
    foreach ($char in $normalizedString.ToCharArray()) {
        if ([Globalization.CharUnicodeInfo]::GetUnicodeCategory($char) -ne [Globalization.UnicodeCategory]::NonSpacingMark) {
            [void]$stringBuilder.Append($char)
        }
    }
    return $stringBuilder.ToString()
}

function Convert-ToTerraformDataSource {
    param (
        [Parameter(Mandatory=$true)]
        [string]$DataType,

        [Parameter(Mandatory=$true, ValueFromPipeline=$true)]
        [PSCustomObject]$InputObject
    )

    process {
        $resourceName = $InputObject.ResourceName
        $terraformConfig = @"
    data "$DataType" "$resourceName" {
"@

        foreach ($property in $InputObject.PSObject.Properties) {
            if ($property.Name -ne "ResourceName") {
                $propertyName = $property.Name.ToLower()
                $propertyValue = $property.Value

                # Traitement spécial pour différents types de données
                if ($propertyValue -is [bool]) {
                    $propertyValue = $propertyValue.ToString().ToLower()
                }
                elseif ($propertyValue -is [string]) {
                    $propertyValue = "`"$propertyValue`""
                }
                elseif ($propertyValue -is [array]) {
                    $propertyValue = "[" + ($propertyValue -join ", ") + "]"
                }

                $terraformConfig += @"

      $propertyName = $propertyValue
"@
            }
        }

        $terraformConfig += @"

    }
"@

        return $terraformConfig
    }
}

function Convert-ToTerraformResource {
    param (
        [Parameter(Mandatory=$true)]
        [string]$ResourceType,

        [Parameter(Mandatory=$true)]
        [PSCustomObject]$InputObject
    )

    $resourceName = $InputObject.resourcename
    $terraformConfig = @"
resource "$ResourceType" "$resourceName" {
"@

    foreach ($property in $InputObject.PSObject.Properties) {
        if ($property.Name -ne "resourcename") {
            $propertyName = $property.Name
            $propertyValue = $property.Value

            # Traitement spécial pour différents types de données
            if ($propertyValue -is [bool]) {
                $propertyValue = $propertyValue.ToString().ToLower()
            }
            elseif ($propertyValue -is [string]) {
                # Vérifie si la valeur est une référence à une autre ressource
                if ($propertyValue -match '^(data\.|rtms_)') {
                    # Ne pas mettre de guillemets autour des références
                } else {
                    $propertyValue = "`"$propertyValue`""
                }
            }
            elseif ($propertyValue -is [array]) {
                $propertyValue = "[" + ($propertyValue -join ", ") + "]"
            }

            $terraformConfig += @"

  $propertyName = $propertyValue
"@
        }
    }

    $terraformConfig += @"

}
"@

    return $terraformConfig
}

function Update-DuplicateItems {
    param (
        [Parameter(Mandatory=$true)]
        [Object[]]$List,
        
        [Parameter(Mandatory=$true)]
        [string]$PropertyName
    )

    $uniqueItems = @{}

    for ($i = 0; $i -lt $List.Count; $i++) {
        $item = $List[$i].$PropertyName
        if ($uniqueItems.ContainsKey($item)) {
            $count = $uniqueItems[$item] + 1
            $uniqueItems[$item] = $count
            $List[$i].$PropertyName = "${item}_$count"
        } else {
            $uniqueItems[$item] = 0
        }
    }
}

function Add-UnderscorePrefix {
    param (
        [Parameter(Mandatory=$true, ValueFromPipeline=$true)]
        [string]$InputString
    )
    
    process {
        return $InputString -replace '^(\d)', '_$1'
    }
}

# Liste des Hosts
$ListeHosts = @()
$CptPage = 1
$FinListe = $False
While (!($FinListe)){
    $ExtraitListeHosts = (Invoke-RestMethod -Uri "https://rtms-api.cloud-temple.com/v1/hosts?cloudTempleId=$CloudTempleId&order=DESC&orderBy=id&page=$($CptPage)&itemsPerPage=500" -Headers $APIHeaders).data
    if ($ExtraitListeHosts){
        $ListeHosts += $ExtraitListeHosts
        $CptPage ++
    }
    else{$FinListe = $True}
}
$ListeDetailedHosts = @()
foreach ($Id in $ListeHosts.id){
    $ListeDetailedHosts += (Invoke-RestMethod -Uri "https://rtms-api.cloud-temple.com/v1/hosts/$Id" -Headers $APIHeaders).data
}

# Liste des services
$ListeServices = @()
$CptPage = 1
$FinListe = $False
While (!($FinListe)){
    $ExtraitListeServices = (Invoke-RestMethod -Uri "https://rtms-api.cloud-temple.com/v1/monitoringServices?cloudTempleId=$CloudTempleId&order=DESC&orderBy=id&page=$($CptPage)&itemsPerPage=500" -Headers $APIHeaders).data
    if ($ExtraitListeServices){
        $ListeServices += $ExtraitListeServices
        $CptPage ++
    }
    else{$FinListe = $True}
}
$ListeDetailedServices = @()
foreach ($Id in $ListeServices.id){
    $DetailedService = (Invoke-RestMethod -Uri "https://rtms-api.cloud-temple.com/v1/monitoringServices/$Id" -Headers $APIHeaders).data
    $ListeDetailedServices += $DetailedService
}

############################# TERRAFORM CONFIG 

##############
# APPLIANCES #
##############
$ListeTFAppliances = @()
$ListeAppliancesId = $ListeDetailedServices.appliance.id | select -Unique
foreach ($Id in $ListeAppliancesId){
    $Appliance = ($ListeDetailedServices | ?{$_.appliance.id -eq $Id})[0].appliance
    $TFAppliance = [PSCustomObject]@{ResourceName=($Appliance.name -replace '[.()[\]"''\/\\ ]', '_');Name=$Appliance.name;Id=$Appliance.id;Alias=$Appliance.alias;Appliance=$Appliance.address}
    $ListeTFAppliances += $TFAppliance
    Add-Content -Path $TFDataFilePath -Value ($TFAppliance | Convert-ToTerraformDataSource -DataType "rtms_appliance") -Encoding UTF8
}

###########
# PLUGINS #
###########
$ListeTFPlugins = @()
$ListePluginsId = $ListeDetailedServices.plugin.id | select -Unique
foreach ($Id in $ListePluginsId){
    $Plugin = ($ListeDetailedServices | ?{$_.plugin.id -eq $Id})[0].plugin
    $TFPlugin = [PSCustomObject]@{ResourceName=($Plugin.name -replace '[.()[\]"''\/\\ ]', '_');Name=$Plugin.name;Id=$Plugin.id;IsDeprecated=$Plugin.isDeprecated}
    $ListeTFPlugins += $TFPlugin
    Add-Content -Path $TFDataFilePath -Value ($TFPlugin | Convert-ToTerraformDataSource -DataType "rtms_plugin") -Encoding UTF8
}

#############
# TEMPLATES #
#############
$ListeTFTemplates = @()
$ListeTemplatesId = $ListeDetailedServices.template.id | select -Unique
foreach ($Id in $ListeTemplatesId){
    $Template = ($ListeDetailedServices | ?{$_.template.id -eq $Id})[0].template
    $TFTemplate = [PSCustomObject]@{ResourceName=($Template.name -replace '[.()[\]"''\/\\ ]', '_');Name=$Template.name;Id=$Template.id}
    $ListeTFTemplates += $TFTemplate
    Add-Content -Path $TFDataFilePath -Value ($TFTemplate | Convert-ToTerraformDataSource -DataType "rtms_template") -Encoding UTF8
}

#################
# CATALOGSITEMS # (on se base sur les services existants pour générer une liste de possibilités)
#################
$ListeCatalogsCombo = @()
foreach ($DetailedService in $ListeDetailedServices){
    if ($DetailedService.ticketCatalogsItems){
        $ListeIds = ($DetailedService.ticketCatalogsItems.id | sort-object)
        $ListeCatalogsCombo += [PSCustomObject]@{IdService=$DetailedService.id; IdCatalogs=($ListeIds -join '-')}
    }
}
$ListeTFCatalogsItems = @()
foreach ($IdChain in ($ListeCatalogsCombo.IdCatalogs | select -unique)){
    $IdService = ($ListeCatalogsCombo | ? IdCatalogs -eq $IdChain)[0].IdService
    $Service = $ListeDetailedServices | ? id -eq $IdService
    $Description = Remove-Diacritics -Text (($Service.ticketCatalogsItems.name) -join '/')
    $ResourceName = $Description
    $ResourceName = $ResourceName.replace("Remontee d'alerte (RTMS)/","AlerteRTMS")
    $ResourceName = $ResourceName.replace("OSMOSe - Informatique de production/","InfoProd")
    $ResourceName = $ResourceName.replace(" - ","-").replace("'"," ")
    $ResourceName = (($ResourceName.split('/') -join "_").Split() | ForEach-Object { $_.Substring(0,1).ToUpper() + $_.Substring(1) }) -join ''
    $TFCatalogsItems = [PSCustomObject]@{ResourceName=($ResourceName -replace '[.()[\]"''\/\\ ]', '_');Name=$ResourceName;Description=$Description;Id=($Service.ticketCatalogsItems.id)}
    $ListeTFCatalogsItems += $TFCatalogsItems
    Add-Content -Path $TFDataFilePath -Value ($TFCatalogsItems | Convert-ToTerraformDataSource -DataType "rtms_typology") -Encoding UTF8
}

########
# TEAM #
########
$ListeTFTeams = @()
$ListeTeamsId = $ListeDetailedServices.responsibleTeam.id | select -Unique
foreach ($Id in $ListeTeamsId){
    $Team = ($ListeDetailedServices | ?{$_.responsibleTeam.id -eq $Id})[0].responsibleTeam
    $TFTeam = [PSCustomObject]@{ResourceName=($(Remove-Diacritics -Text $Team.name) -replace '[.()[\]"''\/\\ ]', '_');Name=$Team.name;Id=$Team.id}
    $ListeTFTeams += $TFTeam
    Add-Content -Path $TFDataFilePath -Value ($TFTeam | Convert-ToTerraformDataSource -DataType "rtms_team") -Encoding UTF8
}

###############
# CHECKPERIOD #
###############
$ListeTFCheckPeriods = @()
$ListeCheckPeriodsId = $ListeDetailedServices.checkperiod.id | select -Unique
foreach ($Id in $ListeCheckPeriodsId){
    $CheckPeriod = ($ListeDetailedServices | ?{$_.checkperiod.id -eq $Id})[0].checkperiod
    $TFCheckPeriod = [PSCustomObject]@{ResourceName=($(Remove-Diacritics -Text $CheckPeriod.name) -replace '[.()[\]"''\/\\ ]', '_' | Add-UnderscorePrefix);Name=$CheckPeriod.name;Id=$CheckPeriod.id}
    $ListeTFCheckPeriods += $TFCheckPeriod
    Add-Content -Path $TFDataFilePath -Value ($TFCheckPeriod | Convert-ToTerraformDataSource -DataType "rtms_checkperiod") -Encoding UTF8
}

##############
# TIMEPERIOD #
##############
$ListeTFTimePeriods = @()
$ListeTimePeriodsId = $ListeDetailedServices.timeperiod.id | select -Unique
foreach ($Id in $ListeTimePeriodsId){
    $TimePeriod = ($ListeDetailedServices | ?{$_.timeperiod.id -eq $Id})[0].timeperiod
    $TFTimePeriod = [PSCustomObject]@{ResourceName=($(Remove-Diacritics -Text $TimePeriod.name) -replace '[.()[\]"''\/\\ ]', '_' | Add-UnderscorePrefix);Name=$TimePeriod.name;Id=$TimePeriod.id}
    $ListeTFTimePeriods += $TFTimePeriod
    Add-Content -Path $TFDataFilePath -Value ($TFTimePeriod | Convert-ToTerraformDataSource -DataType "rtms_timeperiod") -Encoding UTF8
}

####################
# HOSTS & SERVICES #
####################

foreach ($RTMSHost in $ListeDetailedHosts){
    $TFHost = [PSCustomObject]@{resourcename=$RTMSHost.name
                                name=$RTMSHost.name
                                alias=$RTMSHost.alias
                                address=$RTMSHost.address
                                }
    if ($RTMSHost.community){ $TFHost | Add-Member -NotePropertyName 'community' -NotePropertyValue $RTMSHost.community }
    #if ($RTMSHost.isMonitored){ $TFHost | Add-Member -NotePropertyName 'is_monitored' -NotePropertyValue $RTMSHost.isMonitored }
    #if ($RTMSHost.notificationsEnabled){ $TFHost | Add-Member -NotePropertyName 'notifications_enabled' -NotePropertyValue $RTMSHost.notificationsEnabled }
    if ($RTMSHost.adminLogin){ $TFHost | Add-Member -NotePropertyName 'admin_login' -NotePropertyValue $RTMSHost.adminLogin }
    if ($RTMSHost.type){ $TFHost | Add-Member -NotePropertyName 'type' -NotePropertyValue $RTMSHost.type }
    if ($RTMSHost.appliance){$TFHost | Add-Member -NotePropertyName 'appliance' -NotePropertyValue "data.rtms_appliance.$(($ListeTFAppliances | Where-Object { $_.id -eq $RTMSHost.appliance.id }).name).id"}

    Add-Content -Path $TFMainFilePath -Value (Convert-ToTerraformResource $TFHost -ResourceType rtms_host)
    Add-Content -Path $TFImportCmdFilePath -Value ('terraform import "rtms_host.' + $TFHost.ResourceName + '" ' +$RTMSHost.id)

    $ListeHostServices = $($ListeDetailedServices | ?{$_.host.id -eq $RTMSHost.id}) | select @{N="resourceName";E={$_.name}},*
    if ($ListeHostServices){Update-DuplicateItems -List $ListeHostServices -PropertyName "resourceName"}

    foreach ($Service in $ListeHostServices){
        $TFService = [PSCustomObject]@{ resourcename=$TFHost.ResourceName + "_" + $($Service.resourceName -replace '[.()[\]"''\/\\ ]', '_')
                                        name=$Service.name
                                        description=$Service.description
                                        appliance="data.rtms_appliance.$(($ListeTFAppliances | ? id -eq $Service.appliance.id).ResourceName).id"
                                        template="data.rtms_template.$(($ListeTFTemplates | ? id -eq $Service.template.id).ResourceName).id"
                                        }
        if ($Service.nicename) { $TFService | Add-Member -NotePropertyName 'nice_name' -NotePropertyValue $Service.nicename }
        if ($RTMSHost.name) { $TFService | Add-Member -NotePropertyName 'host' -NotePropertyValue "rtms_host.$($RTMSHost.name).id" }
        if ($Service.plugin.id) { $TFService | Add-Member -NotePropertyName 'plugin' -NotePropertyValue "data.rtms_plugin.$(($ListeTFPlugins | ? id -eq $Service.plugin.id).ResourceName).id" }
        if ($null -ne $Service.ismonitored) { $TFService | Add-Member -NotePropertyName 'is_monitored' -NotePropertyValue $Service.ismonitored }
        if ($Service.keywords) { $TFService | Add-Member -NotePropertyName 'keywords' -NotePropertyValue $Service.keywords }
        if ($Service.severity) { $TFService | Add-Member -NotePropertyName 'severity' -NotePropertyValue $Service.severity }
        if ($null -ne $Service.onlynotifyifcritical) { $TFService | Add-Member -NotePropertyName 'only_notify_if_critical' -NotePropertyValue $Service.onlynotifyifcritical }
        if ($Service.normalcheckinterval) { $TFService | Add-Member -NotePropertyName 'normal_check_interval' -NotePropertyValue $Service.normalcheckinterval }
        if ($Service.retrycheckinterval) { $TFService | Add-Member -NotePropertyName 'retry_check_interval ' -NotePropertyValue $Service.retrycheckinterval }
        if ($Service.maxcheckattempts) { $TFService | Add-Member -NotePropertyName 'max_check_attempts' -NotePropertyValue $Service.maxcheckattempts }
        if ($Service.timeperiod) { $TFService | Add-Member -NotePropertyName 'time_period' -NotePropertyValue "data.rtms_timeperiod.$(($ListeTFTimePeriods | ? id -eq $Service.timeperiod.id).ResourceName).id" }
        if ($Service.checkperiod) { $TFService | Add-Member -NotePropertyName 'check_period' -NotePropertyValue "data.rtms_checkperiod.$(($ListeTFCheckPeriods | ? id -eq $Service.checkperiod.id).ResourceName).id" }
        if ($Service.ticketcatalogsitems) { 
            $MatchingTypo = ($ListeTFCatalogsItems | ?{($_.id -join ',') -eq ($Service.ticketCatalogsItems.id -join ',')})
            $TFService | Add-Member -NotePropertyName 'ticket_catalogs_items' -NotePropertyValue "data.rtms_typology.$($MatchingTypo.ResourceName).id"
        }
        if ($Service.autoprocessing) { $TFService | Add-Member -NotePropertyName 'auto_processing' -NotePropertyValue $Service.autoprocessing }
        if ($Service.responsibleteam) { $TFService | Add-Member -NotePropertyName 'responsible_team' -NotePropertyValue "data.rtms_team.$(($ListeTFTeams | ? id -eq $Service.responsibleTeam.id).ResourceName).id" }
        if ($null -ne $Service.notificationsenabled) { $TFService | Add-Member -NotePropertyName 'notifications_enabled' -NotePropertyValue $Service.notificationsenabled }
        if ($Service.help) { $TFService | Add-Member -NotePropertyName 'ignore_changes' -NotePropertyValue 'ignore_changes'
            #$IgnoreChange = [PSCustomObject]@{ignore_changes=@("help")}
            #$TFService | Add-Member -NotePropertyName 'lifecycle' -NotePropertyValue $IgnoreChange
        }

        Add-Content -Path $TFMainFilePath -Value (Convert-ToTerraformResource $TFService -ResourceType rtms_monitoring_service)
        Add-Content -Path $TFImportCmdFilePath -Value ('terraform import "rtms_monitoring_service.' + $TFService.ResourceName + '" ' + $Service.id)
    }
}

# Bricolage de secours pour gérer le paramètre help
$content = [System.IO.File]::ReadAllText($TFMainFilePath).Replace('ignore_changes = "ignore_changes"',"lifecycle {
    ignore_changes = [
      help,
    ]
  }")
[System.IO.File]::WriteAllText($TFMainFilePath, $content)

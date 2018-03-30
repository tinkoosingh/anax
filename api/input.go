package api

import (
	"fmt"
	"github.com/open-horizon/anax/microservice"
	"github.com/open-horizon/anax/persistence"
	"reflect"
	"strconv"
)

type Configstate struct {
	State          *string `json:"state"`
	LastUpdateTime *uint64 `json:"last_update_time,omitempty"`
}

func (c *Configstate) String() string {
	if c == nil {
		return "Configstate: not set"
	} else {
		return fmt.Sprintf("State: %v, Time: %v", *c.State, *c.LastUpdateTime)
	}
}

type HorizonDevice struct {
	Id                 *string      `json:"id"`
	Org                *string      `json:"organization"`
	Pattern            *string      `json:"pattern"` // a simple name, not prefixed with the org
	Name               *string      `json:"name,omitempty"`
	Token              *string      `json:"token,omitempty"`
	TokenLastValidTime *uint64      `json:"token_last_valid_time,omitempty"`
	TokenValid         *bool        `json:"token_valid,omitempty"`
	HA                 *bool        `json:"ha,omitempty"`
	Config             *Configstate `json:"configstate,omitempty"`
	ServiceBased       *bool        `json:"serviceBased,omitempty"`  // The device is service based if this flag is on, but the flag being off could mean that service or workload based is not yet known.
	WorkloadBased      *bool        `json:"workloadBased,omitempty"` // The device is workload based if this flag is on, but the flag being off could mean that service or workload based is not yet known.
}

func (h HorizonDevice) String() string {

	id := "not set"
	if h.Id != nil {
		id = *h.Id
	}

	org := "not set"
	if h.Org != nil {
		org = *h.Org
	}

	pat := "not set"
	if h.Pattern != nil {
		pat = *h.Pattern
	}

	name := "not set"
	if h.Name != nil {
		name = *h.Name
	}

	cred := "not set"
	if h.Token != nil && *h.Token != "" {
		cred = "set"
	}

	tlvt := uint64(0)
	if h.TokenLastValidTime != nil {
		tlvt = *h.TokenLastValidTime
	}

	tv := false
	if h.TokenValid != nil {
		tv = *h.TokenValid
	}

	ha := false
	if h.HA != nil {
		ha = *h.HA
	}

	sb := false
	if h.ServiceBased != nil {
		sb = *h.ServiceBased
	}

	wb := false
	if h.WorkloadBased != nil {
		wb = *h.WorkloadBased
	}

	return fmt.Sprintf("Id: %v, Org: %v, Pattern: %v, Name: %v, Token: [%v], TokenLastValidTime: %v, TokenValid: %v, HA: %v, ServiceBased: %v, WorkloadBased: %v, %v", id, org, pat, name, cred, tlvt, tv, ha, sb, wb, h.Config)
}

// This is a type conversion function but note that the token field within the persistent
// is explicitly omitted so that it's not exposed in the API.
func ConvertFromPersistentHorizonDevice(pDevice *persistence.ExchangeDevice) *HorizonDevice {
	return &HorizonDevice{
		Id:                 &pDevice.Id,
		Org:                &pDevice.Org,
		Pattern:            &pDevice.Pattern,
		Name:               &pDevice.Name,
		TokenValid:         &pDevice.TokenValid,
		TokenLastValidTime: &pDevice.TokenLastValidTime,
		HA:                 &pDevice.HA,
		Config: &Configstate{
			State:          &pDevice.Config.State,
			LastUpdateTime: &pDevice.Config.LastUpdateTime,
		},
		ServiceBased:  &pDevice.ServiceBased,
		WorkloadBased: &pDevice.WorkloadBased,
	}
}

type Attribute struct {
	Id          *string                 `json:"id"`
	Type        *string                 `json:"type"`
	SensorUrls  *[]string               `json:"sensor_urls"`
	Label       *string                 `json:"label"`
	Publishable *bool                   `json:"publishable"`
	HostOnly    *bool                   `json:"host_only"`
	Mappings    *map[string]interface{} `json:"mappings"`
}

func (a Attribute) String() string {
	// function to make sure the nil pointers get printed without 'invalid memory address' error
	getString := func(v interface{}) string {
		if reflect.ValueOf(v).IsNil() {
			return "<nil>"
		} else {
			return fmt.Sprintf("%v", reflect.Indirect(reflect.ValueOf(v)))
		}
	}

	return fmt.Sprintf("Id: %v, Type: %v, SensorUrls: %v, Label: %v, Publishable: %v, HostOnly: %v, Mappings: %v",
		getString(a.Id), getString(a.Type), getString(a.SensorUrls), getString(a.Label), getString(a.Publishable), getString(a.HostOnly), getString(a.Mappings))
}

func NewAttribute(t string, sURLs []string, l string, publishable bool, hostOnly bool, mappings map[string]interface{}) *Attribute {
	return &Attribute{
		Type:        &t,
		SensorUrls:  &sURLs,
		Label:       &l,
		Publishable: &publishable,
		HostOnly:    &hostOnly,
		Mappings:    &mappings,
	}
}

// uses pointers for members b/c it allows nil-checking at deserialization; !Important!: the json field names here must not change w/out changing the error messages returned from the API, they are not programmatically determined
type MicroService struct {
	SensorUrl     *string      `json:"sensor_url"`     // uniquely identifying
	SensorOrg     *string      `json:"sensor_org"`     // The org that holds the ms definition
	SensorName    *string      `json:"sensor_name"`    // may not be uniquely identifying
	SensorArch    *string      `json:"sensor_arch"`    // the arch of the microservice defined in the exchange
	SensorVersion *string      `json:"sensor_version"` // added for ms split. It is only used for microsevice. If it is omitted, old behavior is asumed.
	AutoUpgrade   *bool        `json:"auto_upgrade"`   // added for ms split. The default is true. If the sensor (microservice) should be automatically upgraded when new versions become available.
	ActiveUpgrade *bool        `json:"active_upgrade"` // added for ms split. The default is false. If horizon should actively terminate agreements when new versions become available (active) or wait for all the associated agreements terminated before making upgrade.
	Attributes    *[]Attribute `json:"attributes"`
}

func (s *MicroService) String() string {
	sURL := ""
	sOrg := ""
	sName := ""
	sArch := ""
	sVersion := ""
	auto_upgrade := ""
	active_upgrade := ""

	if s.SensorUrl != nil {
		sURL = *s.SensorUrl
	}

	if s.SensorOrg != nil {
		sOrg = *s.SensorOrg
	}

	if s.SensorName != nil {
		sName = *s.SensorName
	}

	if s.SensorArch != nil {
		sArch = *s.SensorArch
	}

	if s.SensorVersion != nil {
		sVersion = *s.SensorVersion
	}

	if s.AutoUpgrade != nil {
		auto_upgrade = strconv.FormatBool(*s.AutoUpgrade)
	}

	if s.ActiveUpgrade != nil {
		active_upgrade = strconv.FormatBool(*s.ActiveUpgrade)
	}

	return fmt.Sprintf("SensorUrl: %v, SensorOrg: %v, SensorName: %v, SensorArch: %v, SensorVersion: %v, AutoUpgrade: %v, ActiveUpgrade: %v, Attributes: %v", sURL, sOrg, sName, sArch, sVersion, auto_upgrade, active_upgrade, s.Attributes)
}

// Constructor used to create microservice objects for programmatic creation of microservices.
func NewMicroService(url string, org string, name string, arch string, v string) *MicroService {
	autoUpgrade := microservice.MS_DEFAULT_AUTOUPGRADE
	activeUpgrade := microservice.MS_DEFAULT_ACTIVEUPGRADE

	return &MicroService{
		SensorUrl:     &url,
		SensorOrg:     &org,
		SensorName:    &name,
		SensorArch:    &arch,
		SensorVersion: &v,
		AutoUpgrade:   &autoUpgrade,
		ActiveUpgrade: &activeUpgrade,
		Attributes:    &[]Attribute{},
	}
}

// uses pointers for members b/c it allows nil-checking at deserialization; !Important!: the json field names here must not change w/out changing the error messages returned from the API, they are not programmatically determined
type Service struct {
	Url           *string      `json:"url"`            // The URL of the service definition.
	Org           *string      `json:"organization"`   // The org that holds the service definition.
	Name          *string      `json:"name"`           // Optional, may not be uniquely identifying.
	Arch          *string      `json:"arch"`           // The arch of the service to be configured, could be a synonym.
	VersionRange  *string      `json:"versionRange"`   // The version range that the configuration applies to.
	AutoUpgrade   *bool        `json:"auto_upgrade"`   // The default is true. If the service should be automatically upgraded when a new version becomes available.
	ActiveUpgrade *bool        `json:"active_upgrade"` // The default is false. If horizon should actively terminate agreements when new versions become available (active) or wait for all the associated agreements to terminate before upgrading.
	Attributes    *[]Attribute `json:"attributes"`
}

func (s *Service) String() string {
	sURL := ""
	sOrg := ""
	sName := ""
	sArch := ""
	sVersion := ""
	auto_upgrade := ""
	active_upgrade := ""

	if s.Url != nil {
		sURL = *s.Url
	}

	if s.Org != nil {
		sOrg = *s.Org
	}

	if s.Name != nil {
		sName = *s.Name
	}

	if s.Arch != nil {
		sArch = *s.Arch
	}

	if s.VersionRange != nil {
		sVersion = *s.VersionRange
	}

	if s.AutoUpgrade != nil {
		auto_upgrade = strconv.FormatBool(*s.AutoUpgrade)
	}

	if s.ActiveUpgrade != nil {
		active_upgrade = strconv.FormatBool(*s.ActiveUpgrade)
	}

	return fmt.Sprintf("Url: %v, Org: %v, Name: %v, Arch: %v, VersionRange: %v, AutoUpgrade: %v, ActiveUpgrade: %v, Attributes: %v", sURL, sOrg, sName, sArch, sVersion, auto_upgrade, active_upgrade, s.Attributes)
}

// Constructor used to create service objects for programmatic creation of services.
func NewService(url string, org string, name string, arch string, v string) *Service {
	autoUpgrade := microservice.MS_DEFAULT_AUTOUPGRADE
	activeUpgrade := microservice.MS_DEFAULT_ACTIVEUPGRADE

	return &Service{
		Url:           &url,
		Org:           &org,
		Name:          &name,
		Arch:          &arch,
		VersionRange:  &v,
		AutoUpgrade:   &autoUpgrade,
		ActiveUpgrade: &activeUpgrade,
		Attributes:    &[]Attribute{},
	}
}

// This section is for handling the workloadConfig API input
type WorkloadConfig struct {
	WorkloadURL string      `json:"workload_url"`
	Org         string      `json:"organization"`
	Version     string      `json:"workload_version"` // This is a version range
	Attributes  []Attribute `json:"attributes"`
}

func (w WorkloadConfig) String() string {
	return fmt.Sprintf("WorkloadURL: %v, "+
		"Org: %v, "+
		"Version: %v, "+
		"Attributes: %v",
		w.WorkloadURL, w.Org, w.Version, w.Attributes)
}

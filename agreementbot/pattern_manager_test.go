// +build unit

package agreementbot

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/open-horizon/anax/exchange"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("v", "7")
	// no need to parse flags, that's done by test framework
}

func Test_pattern_entry_success1(t *testing.T) {

	lab := "label"

	p := &exchange.Pattern{
		Label:              lab,
		Description:        "desc",
		Public:             true,
		Workloads:          []exchange.WorkloadReference{},
		AgreementProtocols: []exchange.AgreementProtocol{},
	}

	if np, err := NewPatternEntry(p); err != nil {
		t.Errorf("Error %v creating new pattern entry from %v", err, *p)
	} else if np.Pattern.Label != "label" {
		t.Errorf("Error: label should be %v but is %v", lab, np.Pattern.Label)
	} else if len(np.Hash) != 32 {
		t.Errorf("Error: hash should be length %v", 32)
	} else {
		t.Log(np)
	}

}

func Test_pattern_manager_success1(t *testing.T) {

	if np := NewPatternManager(); np == nil {
		t.Errorf("Error: pattern manager not created")
	} else {
		t.Log(np)
	}

}

// No existing served patterns, no new served patterns
func Test_pattern_manager_setpatterns0(t *testing.T) {

	policyPath := "/tmp/servedpatterntest/"
	servedPatterns := map[string]exchange.ServedPattern{}

	if np := NewPatternManager(); np == nil {
		t.Errorf("Error: pattern manager not created")
	} else if err := np.SetCurrentPatterns(servedPatterns, policyPath); err != nil {
		t.Errorf("Error %v consuming served patterns %v", err, servedPatterns)
	} else if len(np.OrgPatterns) != 0 {
		t.Errorf("Error: should have 0 org in the PatternManager, have %v", len(np.OrgPatterns))
	} else {
		t.Log(np)
	}

}

// Add a new served org and pattern
func Test_pattern_manager_setpatterns1(t *testing.T) {

	policyPath := "/tmp/servedpatterntest/"
	servedPatterns := map[string]exchange.ServedPattern{
		"myorg1_pattern1": {
			Org:     "myorg1",
			Pattern: "pattern1",
		},
	}

	if np := NewPatternManager(); np == nil {
		t.Errorf("Error: pattern manager not created")
	} else if err := np.SetCurrentPatterns(servedPatterns, policyPath); err != nil {
		t.Errorf("Error %v consuming served patterns %v", err, servedPatterns)
	} else if len(np.OrgPatterns) != 1 {
		t.Errorf("Error: should have 1 org in the PatternManager, have %v", len(np.OrgPatterns))
	} else {
		t.Log(np)
	}

}

// Remove an org and pattern, replace with a new org and pattern
func Test_pattern_manager_setpatterns2(t *testing.T) {

	policyPath := "/tmp/servedpatterntest/"
	myorg1 := "myorg1"
	myorg2 := "myorg2"
	pattern1 := "pattern1"
	pattern2 := "pattern2"

	servedPatterns1 := map[string]exchange.ServedPattern{
		"myorg1_pattern1": {
			Org:     myorg1,
			Pattern: pattern1,
		},
	}

	servedPatterns2 := map[string]exchange.ServedPattern{
		"myorg2_pattern2": {
			Org:     myorg2,
			Pattern: pattern2,
		},
	}

	definedPatterns1 := map[string]exchange.Pattern{
		"myorg1/pattern1": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test1",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "1.0.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
	}

	definedPatterns2 := map[string]exchange.Pattern{
		"myorg2/pattern2": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test2",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "1.5.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
	}

	// setup test
	if err := cleanTestDir(policyPath); err != nil {
		t.Errorf(err.Error())
	}

	// run test
	if np := NewPatternManager(); np == nil {
		t.Errorf("Error: pattern manager not created")
	} else if err := np.SetCurrentPatterns(servedPatterns1, policyPath); err != nil {
		t.Errorf("Error %v consuming served patterns %v", err, servedPatterns1)
	} else if err := np.UpdatePatternPolicies(myorg1, definedPatterns1, policyPath); err != nil {
		t.Errorf("Error: error updating pattern policies, %v", err)
	} else if len(np.OrgPatterns) != 1 {
		t.Errorf("Error: should have 1 org in the PatternManager, have %v", len(np.OrgPatterns))
	} else if !np.hasOrg(myorg1) {
		t.Errorf("Error: PM should have org %v but doesnt, has %v", myorg1, np)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg1][pattern1].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg1, pattern1, err)
	} else if err := np.SetCurrentPatterns(servedPatterns2, policyPath); err != nil {
		t.Errorf("Error %v consuming served patterns %v", err, servedPatterns2)
	} else if err := np.UpdatePatternPolicies(myorg2, definedPatterns2, policyPath); err != nil {
		t.Errorf("Error: error updating pattern policies, %v", err)
	} else if len(np.OrgPatterns) != 1 {
		t.Errorf("Error: should have 1 org in the PatternManager, have %v", len(np.OrgPatterns))
	} else if !np.hasOrg(myorg2) {
		t.Errorf("Error: PM should have org %v but doesnt, has %v", myorg2, np)
	} else if np.hasOrg(myorg1) {
		t.Errorf("Error: PM should NOT have org %v but does %v", myorg1, np)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg2][pattern2].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg2, pattern2, err)
	} else if files, err := getPolicyFiles(policyPath + myorg1); err != nil {
		t.Errorf(err.Error())
	} else if len(files) != 0 {
		t.Errorf("Error: found policy files for %v, %v", myorg1, files)
	} else {
		t.Log(np)
	}

}

// Remove an org with multiple patterns, add a pattern to existing org
func Test_pattern_manager_setpatterns3(t *testing.T) {

	policyPath := "/tmp/servedpatterntest/"
	myorg1 := "myorg1"
	myorg2 := "myorg2"
	pattern1 := "pattern1"
	pattern2 := "pattern2"

	servedPatterns1 := map[string]exchange.ServedPattern{
		"myorg1_pattern1": {
			Org:     myorg1,
			Pattern: pattern1,
		},
		"myorg1_pattern2": {
			Org:     myorg1,
			Pattern: pattern2,
		},
		"myorg2_pattern2": {
			Org:     myorg2,
			Pattern: pattern2,
		},
	}

	servedPatterns2 := map[string]exchange.ServedPattern{
		"myorg2_pattern1": {
			Org:     myorg2,
			Pattern: pattern1,
		},
		"myorg2_pattern2": {
			Org:     myorg2,
			Pattern: pattern2,
		},
	}

	definedPatterns1 := map[string]exchange.Pattern{
		"myorg1/pattern1": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test1",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "1.0.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
		"myorg1/pattern2": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test1",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "2.0.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
	}

	definedPatterns2 := map[string]exchange.Pattern{
		"myorg2/pattern1": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test2",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "1.4.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
		"myorg2/pattern2": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test2",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "1.5.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
	}

	// setup test
	if err := cleanTestDir(policyPath); err != nil {
		t.Errorf(err.Error())
	}

	// run test
	if np := NewPatternManager(); np == nil {
		t.Errorf("Error: pattern manager not created")
	} else if err := np.SetCurrentPatterns(servedPatterns1, policyPath); err != nil {
		t.Errorf("Error %v consuming served patterns %v", err, servedPatterns1)
	} else if err := np.UpdatePatternPolicies(myorg1, definedPatterns1, policyPath); err != nil {
		t.Errorf("Error: error updating pattern policies, %v", err)
	} else if err := np.UpdatePatternPolicies(myorg2, definedPatterns2, policyPath); err != nil {
		t.Errorf("Error: error updating pattern policies, %v", err)
	} else if len(np.OrgPatterns) != 2 {
		t.Errorf("Error: should have 2 orgs in the PatternManager, have %v", len(np.OrgPatterns))
	} else if !np.hasOrg(myorg1) {
		t.Errorf("Error: PM should have org %v but doesnt, has %v", myorg1, np)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg1][pattern1].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg1, pattern1, err)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg1][pattern2].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg1, pattern2, err)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg2][pattern2].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg2, pattern2, err)
	} else if err := np.SetCurrentPatterns(servedPatterns2, policyPath); err != nil {
		t.Errorf("Error %v consuming served patterns %v", err, servedPatterns2)
	} else if err := np.UpdatePatternPolicies(myorg2, definedPatterns2, policyPath); err != nil {
		t.Errorf("Error: error updating pattern policies, %v", err)
	} else if len(np.OrgPatterns) != 1 {
		t.Errorf("Error: should have 1 org in the PatternManager, have %v", len(np.OrgPatterns))
	} else if !np.hasOrg(myorg2) {
		t.Errorf("Error: PM should have org %v but doesnt, has %v", myorg2, np)
	} else if np.hasOrg(myorg1) {
		t.Errorf("Error: PM should NOT have org %v but does %v", myorg1, np)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg2][pattern1].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg2, pattern1, err)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg2][pattern2].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg2, pattern2, err)
	} else if files, err := getPolicyFiles(policyPath + myorg1); err != nil {
		t.Errorf(err.Error())
	} else if len(files) != 0 {
		t.Errorf("Error: found policy files for %v, %v", myorg1, files)
	} else {
		t.Log(np)
	}

}

// // Remove a pattern but org stays around, add a pattern to existing org
func Test_pattern_manager_setpatterns4(t *testing.T) {

	policyPath := "/tmp/servedpatterntest/"
	myorg1 := "myorg1"
	myorg2 := "myorg2"
	pattern1 := "pattern1"
	pattern2 := "pattern2"

	servedPatterns1 := map[string]exchange.ServedPattern{
		"myorg1_pattern1": {
			Org:     myorg1,
			Pattern: pattern1,
		},
		"myorg1_pattern2": {
			Org:     myorg1,
			Pattern: pattern2,
		},
		"myorg2_pattern2": {
			Org:     myorg2,
			Pattern: pattern2,
		},
	}

	servedPatterns2 := map[string]exchange.ServedPattern{
		"myorg1_pattern1": {
			Org:     myorg1,
			Pattern: pattern1,
		},
		"myorg2_pattern1": {
			Org:     myorg2,
			Pattern: pattern1,
		},
		"myorg2_pattern2": {
			Org:     myorg2,
			Pattern: pattern2,
		},
	}

	definedPatterns1 := map[string]exchange.Pattern{
		"myorg1/pattern1": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test1",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "1.0.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
		"myorg1/pattern2": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test1",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "2.0.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
	}

	definedPatterns2 := map[string]exchange.Pattern{
		"myorg2/pattern1": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test2",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "1.4.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
		"myorg2/pattern2": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test2",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "1.5.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
	}

	// setup the test
	if err := cleanTestDir(policyPath); err != nil {
		t.Errorf(err.Error())
	}

	// run the test
	if np := NewPatternManager(); np == nil {
		t.Errorf("Error: pattern manager not created")
	} else if err := np.SetCurrentPatterns(servedPatterns1, policyPath); err != nil {
		t.Errorf("Error %v consuming served patterns %v", err, servedPatterns1)
	} else if err := np.UpdatePatternPolicies(myorg1, definedPatterns1, policyPath); err != nil {
		t.Errorf("Error: error updating pattern policies, %v", err)
	} else if err := np.UpdatePatternPolicies(myorg2, definedPatterns2, policyPath); err != nil {
		t.Errorf("Error: error updating pattern policies, %v", err)
	} else if len(np.OrgPatterns) != 2 {
		t.Errorf("Error: should have 2 orgs in the PatternManager, have %v", len(np.OrgPatterns))
	} else if !np.hasOrg(myorg1) {
		t.Errorf("Error: PM should have org %v but doesnt, has %v", myorg1, np)
	} else if !np.hasOrg(myorg2) {
		t.Errorf("Error: PM should have org %v but doesnt, has %v", myorg2, np)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg1][pattern1].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg1, pattern1, err)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg1][pattern2].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg1, pattern2, err)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg2][pattern2].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg2, pattern2, err)
	} else if err := np.SetCurrentPatterns(servedPatterns2, policyPath); err != nil {
		t.Errorf("Error %v consuming served patterns %v", err, servedPatterns2)
	} else if err := np.UpdatePatternPolicies(myorg1, definedPatterns1, policyPath); err != nil {
		t.Errorf("Error: error updating pattern policies, %v", err)
	} else if err := np.UpdatePatternPolicies(myorg2, definedPatterns2, policyPath); err != nil {
		t.Errorf("Error: error updating pattern policies, %v", err)
	} else if len(np.OrgPatterns) != 2 {
		t.Errorf("Error: should have 2 org in the PatternManager, have %v", len(np.OrgPatterns))
	} else if !np.hasOrg(myorg2) {
		t.Errorf("Error: PM should have org %v but doesnt, has %v", myorg2, np)
	} else if !np.hasOrg(myorg1) {
		t.Errorf("Error: PM should have org %v but doesnt, has %v", myorg1, np)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg1][pattern1].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg1, pattern1, err)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg2][pattern1].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg2, pattern1, err)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg2][pattern2].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg2, pattern2, err)
	} else {
		t.Log(np)
	}
}

// UpdatePatternPolicies removes the pattern, org and the policy files
func Test_pattern_manager_setpatterns5(t *testing.T) {

	policyPath := "/tmp/servedpatterntest/"
	myorg1 := "myorg1"
	myorg2 := "myorg2"
	pattern1 := "pattern1"
	pattern2 := "pattern2"

	servedPatterns1 := map[string]exchange.ServedPattern{
		"myorg1_pattern1": {
			Org:     myorg1,
			Pattern: pattern1,
		},
		"myorg1_pattern2": {
			Org:     myorg1,
			Pattern: pattern2,
		},
		"myorg2_pattern2": {
			Org:     myorg2,
			Pattern: pattern2,
		},
	}

	definedPatterns1 := map[string]exchange.Pattern{
		"myorg1/pattern1": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test1",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "1.0.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
		"myorg1/pattern2": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test1",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "2.0.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
	}

	definedPatterns2 := map[string]exchange.Pattern{
		"myorg2/pattern1": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test2",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "1.4.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
		"myorg2/pattern2": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test2",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "1.5.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
	}

	definedPatterns11 := map[string]exchange.Pattern{
		"myorg1/pattern1": exchange.Pattern{
			Label:       "label",
			Description: "description",
			Public:      false,
			Workloads: []exchange.WorkloadReference{
				{
					WorkloadURL:  "http://mydomain.com/workload/test1",
					WorkloadOrg:  "testorg",
					WorkloadArch: "amd64",
					WorkloadVersions: []exchange.WorkloadChoice{
						{
							Version: "1.0.0",
						},
					},
				},
			},
			AgreementProtocols: []exchange.AgreementProtocol{
				{Name: "Basic"},
			},
		},
	}

	// setup the test
	if err := cleanTestDir(policyPath); err != nil {
		t.Errorf(err.Error())
	}

	// run the test
	if np := NewPatternManager(); np == nil {
		t.Errorf("Error: pattern manager not created")
	} else if err := np.SetCurrentPatterns(servedPatterns1, policyPath); err != nil {
		t.Errorf("Error %v consuming served patterns %v", err, servedPatterns1)
	} else if err := np.UpdatePatternPolicies(myorg1, definedPatterns1, policyPath); err != nil {
		t.Errorf("Error: error updating pattern policies, %v", err)
	} else if err := np.UpdatePatternPolicies(myorg2, definedPatterns2, policyPath); err != nil {
		t.Errorf("Error: error updating pattern policies, %v", err)
	} else if len(np.OrgPatterns) != 2 {
		t.Errorf("Error: should have 2 orgs in the PatternManager, have %v", len(np.OrgPatterns))
	} else if !np.hasOrg(myorg1) {
		t.Errorf("Error: PM should have org %v but doesnt, has %v", myorg1, np)
	} else if !np.hasOrg(myorg2) {
		t.Errorf("Error: PM should have org %v but doesnt, has %v", myorg2, np)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg1][pattern1].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg1, pattern1, err)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg1][pattern2].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg1, pattern2, err)
	} else if err := getPatternEntryFiles(np.OrgPatterns[myorg2][pattern2].PolicyFileNames); err != nil {
		t.Errorf("Error getting pattern entry files for %v %v, %v", myorg2, pattern2, err)
	} else {
		files_delete := np.OrgPatterns[myorg1][pattern2].PolicyFileNames
		if err := np.UpdatePatternPolicies(myorg1, definedPatterns11, policyPath); err != nil {
			t.Errorf("Error: error updating pattern policies, %v", err)
		} else if err := getPatternEntryFiles(files_delete); err == nil {
			t.Errorf("Should return error but got nil for checking policy files %v", files_delete)
		} else if len(np.OrgPatterns[myorg1]) != 1 {
			t.Errorf("Error: PM should have 1 pattern for org %v but got %v", myorg1, np.OrgPatterns[myorg1])
		} else {
			files_delete1 := np.OrgPatterns[myorg1][pattern1].PolicyFileNames
			if err := np.UpdatePatternPolicies(myorg1, make(map[string]exchange.Pattern), policyPath); err != nil {
				t.Errorf("Error: error updating pattern policies, %v", err)
			} else if np.hasOrg(myorg1) {
				t.Errorf("Error: org %v should have deleted but not.", myorg1)
			} else if err := getPatternEntryFiles(files_delete1); err == nil {
				t.Errorf("Should return error but got nil for checking policy files %v", files_delete1)
			} else if !np.hasOrg(myorg2) {
				t.Errorf("Error: org %v should be left but not.", myorg2)
			} else if err := getPatternEntryFiles(np.OrgPatterns[myorg2][pattern2].PolicyFileNames); err != nil {
				t.Errorf("Error getting pattern entry files for %v %v, %v", myorg2, pattern2, err)
			} else {
				t.Log(np)
			}
		}
	}
}

// Utility functions
// Clean up the test directory
func cleanTestDir(policyPath string) error {
	if _, err := os.Stat(policyPath); !os.IsNotExist(err) {
		if err := os.RemoveAll(policyPath); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(policyPath, 0764); err != nil {
		return err
	}
	return nil
}

// Check for policy files referenced by the pattern manager entries
func getPatternEntryFiles(files []string) error {
	for _, filename := range files {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return errors.New(fmt.Sprintf("File %v does not exist", filename))
		}
	}
	return nil
}

// Check for policy files that shouldnt have been left behind
func getPolicyFiles(homePath string) ([]os.FileInfo, error) {
	res := make([]os.FileInfo, 0, 10)

	if files, err := ioutil.ReadDir(homePath); err != nil {
		return nil, err
	} else {
		for _, fileInfo := range files {
			if strings.HasSuffix(fileInfo.Name(), ".policy") && !fileInfo.IsDir() {
				res = append(res, fileInfo)
			}
		}
		return res, nil
	}
}

// test if different order in the struct could change the hash.
func Test_pattern_manager_hashPattern(t *testing.T) {

	p_exp := getTestPattern()
	h_exp, _ := hashPattern(&p_exp)

	p_exp2 := getTestPattern2()
	h_exp2, _ := hashPattern(&p_exp2)

	if !bytes.Equal(h_exp, h_exp2) {
		t.Errorf("Error: Hashed are different. Hash1=%v, Hash2=%v", h_exp, h_exp2)
	}
}

// test large data
func Test_pattern_manager_setpatterns6(t *testing.T) {

	policyPath := "/tmp/servedpatterntest/"

	// setup the test
	if err := cleanTestDir(policyPath); err != nil {
		t.Errorf(err.Error())
	}

	servedPatterns := map[string]exchange.ServedPattern{
		"org1_EdgeType":           {Org: "org1", Pattern: "EdgeType", LastUpdated: "2018-05-14T19:20:27.187Z[UTC]"},
		"org2_edgegateway":        {Org: "org2", Pattern: "edgegateway", LastUpdated: "2018-04-25T15:10:12.153Z[UTC]"},
		"org3_EdgeType":           {Org: "org3", Pattern: "EdgeType", LastUpdated: "2018-05-21T14:40:51.017Z[UTC]"},
		"org4_EdgeType":           {Org: "org4", Pattern: "EdgeType", LastUpdated: "2018-05-21T14:28:50.608Z[UTC]"},
		"org5_EdgeType":           {Org: "org5", Pattern: "EdgeType", LastUpdated: "2018-05-18T21:13:09.358Z[UTC]"},
		"org6_EdgeType":           {Org: "org6", Pattern: "EdgeType", LastUpdated: "2018-04-17T14:44:00.957Z[UTC]"},
		"org7_EdgeType":           {Org: "org7", Pattern: "EdgeType", LastUpdated: "2018-05-10T12:12:20.640Z[UTC]"},
		"org8_EdgeType":           {Org: "org8", Pattern: "EdgeType", LastUpdated: "2018-05-04T19:13:23.210Z[UTC]"},
		"org9_myGatewayType":      {Org: "org9", Pattern: "myGatewayType", LastUpdated: "2018-05-16T14:19:25.557Z[UTC]"},
		"org10_EdgeType":          {Org: "org10", Pattern: "EdgeType", LastUpdated: "2018-05-04T18:31:01.533Z[UTC]"},
		"org11_myanothertypeEdge": {Org: "org11", Pattern: "myanothertypeEdge", LastUpdated: "2018-05-15T17:58:46.386Z[UTC]"},
		"org11_EdgeType":          {Org: "org11", Pattern: "EdgeType", LastUpdated: "2018-04-24T19:43:16.427Z[UTC]"},
		"org12_EdgeType":          {Org: "org12", Pattern: "EdgeType", LastUpdated: "2018-04-24T19:43:16.427Z[UTC]"},
		"org13_EdgeType":          {Org: "org13", Pattern: "EdgeType", LastUpdated: "2018-05-03T16:40:38.945Z[UTC]"},
		"org14_EdgeType":          {Org: "org14", Pattern: "EdgeType", LastUpdated: "2018-04-24T12:16:47.278Z[UTC]"},
		"org15_EdgeType":          {Org: "org15", Pattern: "EdgeType", LastUpdated: "2018-04-17T14:07:43.350Z[UTC]"},
		"org16_EdgeType":          {Org: "org16", Pattern: "EdgeType", LastUpdated: "2018-04-18T18:22:34.237Z[UTC]"},
		"org17_EdgeType":          {Org: "org17", Pattern: "EdgeType", LastUpdated: "2018-05-18T19:26:10.097Z[UTC]"},
		"org18_EdgeType":          {Org: "org18", Pattern: "EdgeType", LastUpdated: "2018-05-11T21:14:11.998Z[UTC]"},
		"org19_EdgeType":          {Org: "org19", Pattern: "EdgeType", LastUpdated: "2018-04-19T13:52:13.210Z[UTC]"},
		"org20_EdgeType":          {Org: "org20", Pattern: "EdgeType", LastUpdated: "2018-04-23T12:12:01.337Z[UTC]"},
		"org21_EdgeType":          {Org: "org21", Pattern: "EdgeType", LastUpdated: "2018-04-18T14:29:20.840Z[UTC]"},
		"org22_p11":               {Org: "org22", Pattern: "p11", LastUpdated: "2018-05-07T19:31:24.801Z[UTC]"},
		"org22_p12":               {Org: "org22", Pattern: "p12", LastUpdated: "2018-05-18T13:42:11.294Z[UTC]"},
		"org22_p13":               {Org: "org22", Pattern: "p13", LastUpdated: "2018-05-02T19:53:09.428Z[UTC]"},
		"org22_p14":               {Org: "org22", Pattern: "p14", LastUpdated: "2018-05-14T15:02:49.802Z[UTC]"},
		"org22_p15":               {Org: "org22", Pattern: "p15", LastUpdated: "2018-05-16T19:37:46.886Z[UTC]"},
		"org22_p16":               {Org: "org22", Pattern: "p16", LastUpdated: "2018-05-16T20:19:02.775Z[UTC]"},
		"org22_p17":               {Org: "org22", Pattern: "p17", LastUpdated: "2018-05-16T20:19:02.775Z[UTC]"},
		"org22_p21":               {Org: "org22", Pattern: "p21", LastUpdated: "2018-05-17T14:05:47.301Z[UTC]"},
		"org22_p22":               {Org: "org22", Pattern: "p22", LastUpdated: "2018-05-14T14:56:11.403Z[UTC]"},
		"org22_p23":               {Org: "org22", Pattern: "p23", LastUpdated: "2018-05-07T19:59:17.033Z[UTC]"},
		"org23_myanothertypeEdge": {Org: "org23", Pattern: "myanothertypeEdge", LastUpdated: "2018-05-15T14:48:05.986Z[UTC]"},
		"org24_EdgeType":          {Org: "org24", Pattern: "EdgeType", LastUpdated: "2018-05-10T16:47:55.533Z[UTC]"},
		"org25_EdgeType":          {Org: "org25", Pattern: "EdgeType", LastUpdated: "2018-05-18T22:04:00.370Z[UTC]"},
	}

	org_pattern_map := map[string][]string{
		"org1":  {"EdgeType"},
		"org2":  {"edgegateway"},
		"org3":  {"EdgeType"},
		"org4":  {"EdgeType"},
		"org5":  {"EdgeType"},
		"org6":  {"EdgeType"},
		"org7":  {"EdgeType"},
		"org8":  {"EdgeType"},
		"org9":  {"myGatewayType"},
		"org10": {"EdgeType"},
		"org11": {"EdgeType", "myanothertypeEdge"},
		// no org13
		"org12": {"EdgeType"},
		"org14": {"EdgeType"},
		"org15": {"EdgeType"},
		"org16": {"EdgeType"},
		"org17": {"EdgeType"},
		"org18": {"EdgeType"},
		"org19": {"EdgeType"},
		"org20": {"EdgeType"},
		"org21": {"EdgeType"},
		// it does not contain pattern p21, p22 and p23.
		"org22": {"p11", "p12", "p13", "p14", "p15", "p16", "p17", "p33"},
		"org23": {"myanothertypeEdge"},
		"org24": {"EdgeType"},
		"org25": {"EdgeType"},
	}

	np := NewPatternManager()
	if np == nil {
		t.Errorf("Error: pattern manager not created")
	}

	p_exp := getTestPattern()
	h_exp, _ := hashPattern(&p_exp)
	for i := 0; i < 3; i++ {
		err := np.SetCurrentPatterns(servedPatterns, policyPath)
		if err != nil {
			t.Errorf("Error %v consuming served patterns %v", err, servedPatterns)
		}

		for org, ids := range org_pattern_map {
			definedPatterns := make(map[string]exchange.Pattern)
			for _, id := range ids {
				p := getTestPattern()
				h, _ := hashPattern(&p)
				if !bytes.Equal(h_exp, h) {
					t.Errorf("Error: Hashes are different. Hash1=%v, Hash2=%v", h_exp, h)
				}
				definedPatterns[fmt.Sprintf("%v/%v", org, id)] = p
			}
			err := np.UpdatePatternPolicies(org, definedPatterns, policyPath)
			if err != nil {
				t.Errorf("Error: error updating pattern policies, %v", err)
			} else if !np.hasOrg(org) {
				t.Errorf("Error: The pattern manager should container org %v but does not.", org)
			} else {
				for _, id := range ids {
					if _, ok := servedPatterns[fmt.Sprintf("%v_%v", org, id)]; ok {
						if !np.hasPattern(org, id) {
							t.Errorf("Error: The pattern manager should container pattern %v/%v but does not.", org, id)
						}
					}
				}
			}
		}
	}
}

func getTestPattern() exchange.Pattern {
	return exchange.Pattern{
		Owner:       "u1/u1",
		Label:       "Pattern",
		Description: "Pattern for the service version of Core",
		Public:      true,
		Workloads:   []exchange.WorkloadReference{},
		Services: []exchange.ServiceReference{
			{
				ServiceURL:  "https://internetofthings.ibmcloud.com/services/core-iot",
				ServiceOrg:  "IBM",
				ServiceArch: "amd64",
				ServiceVersions: []exchange.WorkloadChoice{
					{
						Version:                      "3.0.0",
						Priority:                     exchange.WorkloadPriority{0, 0, 0, 0},
						Upgrade:                      exchange.UpgradePolicy{"", ""},
						DeploymentOverrides:          "",
						DeploymentOverridesSignature: "ng/uu...",
					},
				},
				DataVerify: exchange.DataVerification{false, "", "", "", 0, 0, exchange.Meter{0, "", 0}},
				NodeH:      exchange.NodeHealth{600, 120},
			},

			{
				ServiceURL:  "https://internetofthings.ibmcloud.com/services/core",
				ServiceOrg:  "IBM",
				ServiceArch: "arm64",
				ServiceVersions: []exchange.WorkloadChoice{
					{
						Version:                      "3.0.0",
						Priority:                     exchange.WorkloadPriority{0, 0, 0, 0},
						Upgrade:                      exchange.UpgradePolicy{"", ""},
						DeploymentOverrides:          "",
						DeploymentOverridesSignature: "N4gkO...",
					},
				},
				DataVerify: exchange.DataVerification{false, "", "", "", 0, 0, exchange.Meter{0, "", 0}},
				NodeH:      exchange.NodeHealth{600, 120},
			},

			{
				ServiceURL:  "https://internetofthings.ibmcloud.com/services/core",
				ServiceOrg:  "IBM",
				ServiceArch: "arm",
				ServiceVersions: []exchange.WorkloadChoice{
					{
						Version:                      "3.0.0",
						Priority:                     exchange.WorkloadPriority{0, 0, 0, 0},
						Upgrade:                      exchange.UpgradePolicy{"", ""},
						DeploymentOverrides:          "",
						DeploymentOverridesSignature: "p2Rwa...",
					},
				},
				DataVerify: exchange.DataVerification{false, "", "", "", 0, 0, exchange.Meter{0, "", 0}},
				NodeH:      exchange.NodeHealth{600, 120},
			},
		},
		AgreementProtocols: []exchange.AgreementProtocol{
			{"Basic", 0, []exchange.Blockchain{}},
		},
	}
}

func getTestPattern2() exchange.Pattern {
	return exchange.Pattern{
		AgreementProtocols: []exchange.AgreementProtocol{
			{"Basic", 0, []exchange.Blockchain{}},
		},
		Description: "Pattern for the service version of Core",
		Public:      true,
		Owner:       "u1/u1",
		Label:       "Pattern",
		Workloads:   []exchange.WorkloadReference{},
		Services: []exchange.ServiceReference{
			{
				ServiceURL:  "https://internetofthings.ibmcloud.com/services/core-iot",
				ServiceOrg:  "IBM",
				ServiceArch: "amd64",
				ServiceVersions: []exchange.WorkloadChoice{
					{
						Version:                      "3.0.0",
						Priority:                     exchange.WorkloadPriority{0, 0, 0, 0},
						Upgrade:                      exchange.UpgradePolicy{"", ""},
						DeploymentOverrides:          "",
						DeploymentOverridesSignature: "ng/uu...",
					},
				},
				DataVerify: exchange.DataVerification{false, "", "", "", 0, 0, exchange.Meter{0, "", 0}},
				NodeH:      exchange.NodeHealth{600, 120},
			},

			{
				ServiceURL:  "https://internetofthings.ibmcloud.com/services/core",
				ServiceOrg:  "IBM",
				ServiceArch: "arm64",
				ServiceVersions: []exchange.WorkloadChoice{
					{
						Version:                      "3.0.0",
						Priority:                     exchange.WorkloadPriority{0, 0, 0, 0},
						Upgrade:                      exchange.UpgradePolicy{"", ""},
						DeploymentOverrides:          "",
						DeploymentOverridesSignature: "N4gkO...",
					},
				},
				DataVerify: exchange.DataVerification{false, "", "", "", 0, 0, exchange.Meter{0, "", 0}},
				NodeH:      exchange.NodeHealth{600, 120},
			},

			{
				ServiceURL:  "https://internetofthings.ibmcloud.com/services/core",
				ServiceOrg:  "IBM",
				ServiceArch: "arm",
				ServiceVersions: []exchange.WorkloadChoice{
					{
						Version:                      "3.0.0",
						Priority:                     exchange.WorkloadPriority{0, 0, 0, 0},
						Upgrade:                      exchange.UpgradePolicy{"", ""},
						DeploymentOverrides:          "",
						DeploymentOverridesSignature: "p2Rwa...",
					},
				},
				DataVerify: exchange.DataVerification{false, "", "", "", 0, 0, exchange.Meter{0, "", 0}},
				NodeH:      exchange.NodeHealth{600, 120},
			},
		},
	}
}

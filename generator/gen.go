package generator

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2016 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"pkg.re/essentialkaos/ek.v3/fsutil"
	"pkg.re/essentialkaos/ek.v3/knf"
	"pkg.re/essentialkaos/ek.v3/path"
	"pkg.re/essentialkaos/ek.v3/system"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const MOCK_TEMPLATE = `@DESCRIPTION
# Enter simple description for this mock

@REQUEST
GET /test.json?random=12345

@RESPONSE
# Add response body here

@CODE
200

@HEADERS
Content-Type:application/json
`

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	DATA_RULE_DIR         = "data:rule-dir"
	ACCESS_USER           = "access:user"
	ACCESS_GROUP          = "access:group"
	ACCESS_MOCK_PERMS     = "access:mock-perms"
	ACCESS_MOCK_DIR_PERMS = "access:mock-dir-perms"
	TEMPLATE_PATH         = "template:path"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Make(name string) error {
	if !fsutil.IsWritable(knf.GetS(DATA_RULE_DIR)) {
		return fmt.Errorf("Directory %s must be writable.", knf.GetS(DATA_RULE_DIR))
	}

	if name == "" {
		return errors.New("You must difine mock file name (service1/mock1 for example)")
	}

	if !strings.Contains(name, "/") {
		return errors.New("You must difine mock file name as <service-id>/<mock-name>.")
	}

	template := knf.GetS(TEMPLATE_PATH)
	ruleDir := knf.GetS(DATA_RULE_DIR)
	dirName := path.Dir(name)
	fullPath := path.Join(ruleDir, name)

	if !strings.Contains(fullPath, ".mock") {
		fullPath += ".mock"
	}

	if fsutil.IsExist(fullPath) {
		return fmt.Errorf("File %s already exist\n", fullPath)
	}

	if template == "" || !fsutil.CheckPerms("FRS", template) {
		return createMock(MOCK_TEMPLATE, dirName, fullPath)
	}

	templData, err := ioutil.ReadFile(template)

	if err != nil {
		return fmt.Errorf("Can't read template content from %s: %v", template, err)
	}

	return createMock(string(templData), dirName, fullPath)
}

func createMock(content, dirName, fullPath string) error {
	serviceDir := path.Join(knf.GetS(DATA_RULE_DIR), dirName)

	err := os.MkdirAll(serviceDir, 0755)

	if err != nil {
		return fmt.Errorf("Can't create directory %s", serviceDir)
	}

	mf, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return fmt.Errorf("Can't create file %s", fullPath)
	}

	defer mf.Close()

	mf.WriteString(content)

	updatePerms(serviceDir, fullPath)

	return nil
}

func updatePerms(dirPath, mockPath string) {
	if knf.GetS(ACCESS_USER) != "" || knf.GetS(ACCESS_GROUP) != "" {
		dirOwnerUID, dirOwnerGID, _ := fsutil.GetOwner(dirPath)
		mockOwnerUID, mockOwnerGID, _ := fsutil.GetOwner(mockPath)

		if knf.GetS(ACCESS_USER) != "" {
			userInfo, err := system.LookupUser(knf.GetS(ACCESS_USER))

			if err == nil {
				dirOwnerUID = userInfo.UID
				mockOwnerUID = userInfo.UID
			}
		}

		if knf.GetS(ACCESS_GROUP) != "" {
			groupInfo, err := system.LookupGroup(knf.GetS(ACCESS_GROUP))

			if err == nil {
				dirOwnerGID = groupInfo.GID
				mockOwnerGID = groupInfo.GID
			}
		}

		os.Chown(dirPath, dirOwnerUID, dirOwnerGID)
		os.Chown(mockPath, mockOwnerUID, mockOwnerGID)
	}

	os.Chmod(dirPath, knf.GetM(ACCESS_MOCK_DIR_PERMS))
	os.Chmod(mockPath, knf.GetM(ACCESS_MOCK_PERMS))
}

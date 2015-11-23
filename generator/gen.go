package generator

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2015 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/essentialkaos/ek/fsutil"
	"github.com/essentialkaos/ek/system"

	"github.com/essentialkaos/mockka/core"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const MockTemplate = `@DESCRIPTION
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

func Make(name string) error {
	if !fsutil.IsWritable(core.Config.GetS(core.ConfMainRuleDir)) {
		return fmt.Errorf("Directory %s must be writable.", core.Config.GetS(core.ConfMainRuleDir))
	}

	if name == "" {
		return errors.New("You must difine mock file name (service1/mock1 for example)")
	}

	if !strings.Contains(name, "/") {
		return errors.New("You must difine mock file name as <service-id>/<mock-name>.")
	}

	template := core.Config.GetS(core.ConfTemplatePath)
	ruleDir := core.Config.GetS(core.ConfMainRuleDir)
	dirName := path.Dir(name)
	fullPath := ruleDir + "/" + name

	if !strings.Contains(fullPath, ".mock") {
		fullPath += ".mock"
	}

	if fsutil.IsExist(fullPath) {
		return fmt.Errorf("File %s already exist\n", fullPath)
	}

	if template == "" || !fsutil.CheckPerms("FRS", template) {
		return createMock(MockTemplate, dirName, fullPath)
	}

	templData, err := ioutil.ReadFile(template)

	if err != nil {
		return fmt.Errorf("Can't read template content from %s - %s", template, err.Error())
	}

	return createMock(string(templData), dirName, fullPath)
}

func createMock(content, dirName, fullPath string) error {
	serviceDir := core.Config.GetS(core.ConfMainRuleDir) + "/" + dirName

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
	if core.Config.GetS(core.ConfAccessUser) != "" || core.Config.GetS(core.ConfAccessGroup) != "" {
		dirOwnerUID, dirOwnerGID, _ := fsutil.GetOwner(dirPath)
		mockOwnerUID, mockOwnerGID, _ := fsutil.GetOwner(mockPath)

		if core.Config.GetS(core.ConfAccessUser) != "" {
			userInfo, err := system.LookupUser(core.Config.GetS(core.ConfAccessUser))

			if err == nil {
				dirOwnerUID = userInfo.UID
				mockOwnerUID = userInfo.UID
			}
		}

		if core.Config.GetS(core.ConfAccessGroup) != "" {
			groupInfo, err := system.LookupGroup(core.Config.GetS(core.ConfAccessGroup))

			if err == nil {
				dirOwnerGID = groupInfo.GID
				mockOwnerGID = groupInfo.GID
			}
		}

		os.Chown(dirPath, dirOwnerUID, dirOwnerGID)
		os.Chown(mockPath, mockOwnerUID, mockOwnerGID)
	}

	os.Chmod(dirPath, core.Config.GetM(core.ConfAccessDirPerms))
	os.Chmod(mockPath, core.Config.GetM(core.ConfAccessMockPerms))
}

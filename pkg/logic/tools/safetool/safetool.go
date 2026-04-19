package safetool

import (
	"fmt"
	"runtime/debug"

	"msa/pkg/model"

	log "github.com/sirupsen/logrus"
)

// SafeExecute provides unified panic recovery protection.
// Wraps tool entry functions to prevent panics from interrupting the session.
func SafeExecute(toolName string, params string, fn func() (string, error)) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Tool [%s] panic recovered: %v\n%s", toolName, r, debug.Stack())
			result = model.NewErrorResult(fmt.Sprintf("工具内部错误: %v", r))
			err = nil
		}
	}()
	result, err = fn()
	log.Infof("Tool [%s] result: %s", toolName, result)
	if err != nil {
		log.Errorf("Tool [%s] error: %v", toolName, err)
		return "", err
	}
	return result, nil
}

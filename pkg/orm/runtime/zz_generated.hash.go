// Code generated by codegen. DO NOT EDIT!!!

package runtime

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func (obj AppInstance) Sha256() string {
	data, _ := json.Marshal(obj)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func (obj Host) Sha256() string {
	data, _ := json.Marshal(obj)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func (obj Job) Sha256() string {
	data, _ := json.Marshal(obj)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

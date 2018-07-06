package jscfg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func ToJson(val interface{}) string {
	data, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func WriteJson(filename string, val interface{}) error {
	data, err := json.MarshalIndent(val, "  ", "  ")
	if err != nil {
		return fmt.Errorf("WriteJson failed: %v %v", filename, err)
	}
	return ioutil.WriteFile(filename, data, 0660)
}

func ReadJson(filename string, val interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("ReadJson failed: %T %v", val, err)
	}
	if err = json.Unmarshal(data, val); err != nil {
		return fmt.Errorf("ReadJson failed: %T %v %v", val, filename, err)
	}
	return nil
}

func ReadJsonByDataStr(jDataStr string, val interface{}) error {
	data := []byte(jDataStr)
	if err := json.Unmarshal(data, val); err != nil {
		return fmt.Errorf("ReadJson failed: %T %v %v", val, jDataStr, err)
	}
	return nil
}

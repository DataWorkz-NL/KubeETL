package main

import (
	"io/ioutil"

	"sigs.k8s.io/yaml"
)

func removeCRDValidation(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	crd := make(map[string]interface{})
	err = yaml.Unmarshal(data, &crd)
	if err != nil {
		panic(err)
	}
	spec := crd["spec"].(map[string]interface{})
	versions := spec["versions"].([]interface{})
	version := versions[0].(map[string]interface{})
	properties := version["schema"].(map[string]interface{})["openAPIV3Schema"].(map[string]interface{})["properties"].(map[string]interface{})
	for k := range properties {
		if k == "spec" || k == "status" {
			properties[k] = map[string]interface{}{"type": "object", "x-kubernetes-preserve-unknown-fields": true}
		}
	}
	data, err = yaml.Marshal(crd)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(filename, data, 0o666)
	if err != nil {
		panic(err)
	}
}

package mapper

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"testing"

	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/stretchr/testify/assert"
)

func TestMessageMapperMap_Ok(t *testing.T) {
	defaultMessageMappper := DefaultMessageMapper{}

	methodeCPH, _ := ioutil.ReadFile("test_resources/methode_cph_update.json")
	methodeCPHJSON := jsonStringToMap(methodeCPH)

	mcp, err := defaultMessageMappper.Map(methodeCPH)
	assert.NoError(t, err, "No error should thrown on correct methode message.")

	assert.Equal(t, methodeCPHJSON["uuid"], mcp.UUID)
	assert.Equal(t, methodeCPHJSON["type"], mcp.Type)
	assert.Equal(t, methodeCPHJSON["workflowStatus"], mcp.WorkflowStatus)
	assert.Equal(t, methodeCPHJSON["attributes"], mcp.AttributesXML)
	assert.Equal(t, methodeCPHJSON["systemAttributes"], mcp.SystemAttributes)
	assert.Equal(t, methodeCPHJSON["usageTickets"], mcp.UsageTickets)
	assert.Equal(t, createMethodeBody(methodeCPHJSON["value"].(string)), mcp.Body)
}

func TestMessageMapperMapDelete_Ok(t *testing.T) {
	defaultMessageMappper := DefaultMessageMapper{}

	methodeCPH, _ := ioutil.ReadFile("test_resources/methode_cph_delete.json")
	methodeCPHJSON := jsonStringToMap(methodeCPH)

	mcp, err := defaultMessageMappper.Map(methodeCPH)
	assert.NoError(t, err, "No error should thrown on correct methode message.")

	assert.Equal(t, methodeCPHJSON["uuid"], mcp.UUID)
	assert.Equal(t, methodeCPHJSON["type"], mcp.Type)
	assert.Equal(t, methodeCPHJSON["workflowStatus"], mcp.WorkflowStatus)
	assert.Equal(t, methodeCPHJSON["attributes"], mcp.AttributesXML)
	assert.Equal(t, methodeCPHJSON["systemAttributes"], mcp.SystemAttributes)
	assert.Equal(t, methodeCPHJSON["usageTickets"], mcp.UsageTickets)
	assert.Equal(t, createMethodeBody(methodeCPHJSON["value"].(string)), mcp.Body)
}

func TestMessageMapperMapWrongType_ThrowsError(t *testing.T) {
	defaultMessageMappper := DefaultMessageMapper{}

	methodeCPH, _ := ioutil.ReadFile("test_resources/methode_cph_wrong_type.json")

	_, err := defaultMessageMappper.Map(methodeCPH)
	assert.Error(t, err, "An error should thrown on wrong type for methode message.")
}

func TestMessageMapperMapWrongSourceCode_ThrowsError(t *testing.T) {
	defaultMessageMappper := DefaultMessageMapper{}

	methodeCPH, _ := ioutil.ReadFile("test_resources/methode_cph_wrong_source_code.json")

	_, err := defaultMessageMappper.Map(methodeCPH)
	assert.Error(t, err, "An error should thrown on wrong type for methode message.")
}

func jsonStringToMap(jsonString []byte) map[string]interface{} {
	var jsonMap map[string]interface{}
	json.Unmarshal(jsonString, &jsonMap)
	return jsonMap
}

func createMethodeBody(methodeBodyXMLBase64 string) model.MethodeBody {
	methodeBodyXML, _ := base64.StdEncoding.DecodeString(methodeBodyXMLBase64)
	var body model.MethodeBody
	xml.Unmarshal([]byte(methodeBodyXML), &body)
	return body
}

package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/goccy/go-json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Example usage:
//   client := NewClientFromEnv()
//   actions, _ := client.GetEMQXListActions()
//   _ = client.CreateEMQXAction("action_name", "desc", "write_syntax", actions)
//   _ = client.CreateEMQXRule("id1", "rule_name", "desc", `SELECT * FROM "topic"`, "action_name")
//   _ = client.ProcessYAMLAndSync()

type Client struct {
	BaseURL  string
	User     string
	Password string
	Client   *http.Client
	Headers  map[string]string
	Logger   *log.Logger
}

// NewClientFromEnv builds a client using environment variables (as in the Python script).
func NewClientFromEnv() *Client {
	host := os.Getenv("EMQX_HOST")
	port := os.Getenv("EMQX_API_PORT")

	emqxUserToken := os.Getenv("EMQX_USER_TOKEN")
	emqxToken := os.Getenv("EMQX_TOKEN")

	base := fmt.Sprintf("http://%s:%s", host, port)

	logger := log.New(os.Stdout, "", log.LstdFlags)

	return &Client{
		BaseURL:  base,
		User:     emqxUserToken,
		Password: emqxToken,
		Client:   &http.Client{},
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
		Logger: logger,
	}
}

// helper to perform a request with basic auth and headers
func (c *Client) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.User, c.Password)
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	return c.Client.Do(req)
}

// GetEMQXListActions returns the raw parsed JSON array of actions.
func (c *Client) GetEMQXListActions() ([]map[string]interface{}, error) {
	url := c.BaseURL + "/api/v5/actions"
	resp, err := c.doRequest(http.MethodGet, url, nil)
	if err != nil {
		c.Logger.Printf("GetEMQXListActions request error: %v", err)
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	c.Logger.Printf("--- Current Configured actions ---\nStatus Code: %d\nResponse: %s\n", resp.StatusCode, string(body))

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var arr []map[string]interface{}
	if err := json.Unmarshal(body, &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

// CreateEMQXAction creates or updates an influxdb action. Returns true on success.
func (c *Client) CreateEMQXAction(actionName, description, influxWriteSyntax string, currentActions []map[string]interface{}) (bool, error) {
	url := c.BaseURL + "/api/v5/actions"
	payload := map[string]interface{}{
		"connector":   "Influx1",
		"description": description,
		"enable":      true,
		"name":        actionName,
		"parameters": map[string]interface{}{
			"precision":    "ms",
			"write_syntax": influxWriteSyntax,
		},
		"resource_opts": map[string]interface{}{
			"health_check_interval": "30s",
		},
		"type": "influxdb",
	}

	// find if action exists
	var found map[string]interface{}
	for _, a := range currentActions {
		if n, ok := a["name"].(string); ok && n == actionName {
			found = a
			break
		}
	}

	var resp *http.Response
	var err error
	if found != nil {
		// copy created_at/last_modified_at if present
		if v, ok := found["created_at"]; ok {
			payload["created_at"] = v
		}
		if v, ok := found["last_modified_at"]; ok {
			payload["last_modified_at"] = v
		}
		updateURL := url + "/influxdb:" + actionName

		// remove name and type for update (matching Python)
		delete(payload, "name")
		delete(payload, "type")

		bs, _ := json.Marshal(payload)
		c.Logger.Printf("--- Updating Action: %s --> %s %s ---", actionName, updateURL, string(bs))
		resp, err = c.doRequest(http.MethodPut, updateURL, bytes.NewReader(bs))
	} else {
		bs, _ := json.Marshal(payload)
		c.Logger.Printf("--- Creating Action: %s ---", actionName)
		resp, err = c.doRequest(http.MethodPost, url, bytes.NewReader(bs))
	}
	if err != nil {
		c.Logger.Printf("CreateEMQXAction request error: %v", err)
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	c.Logger.Printf("Status Code: %d\nResponse: %s\n", resp.StatusCode, string(body))
	return resp.StatusCode >= 200 && resp.StatusCode < 300, nil
}

// CreateEMQXRule creates a rule linking a topic to an action. Returns true on success.
func (c *Client) CreateEMQXRule(ruleID, ruleName, description, sql, actionName string) (bool, error) {
	url := c.BaseURL + "/api/v5/rules"
	payload := map[string]interface{}{
		"sql":         sql,
		"actions":     []string{fmt.Sprintf("influxdb:%s", actionName)},
		"description": description,
		"enable":      true,
		"metadata":    map[string]interface{}{},
		"id":          ruleID,
		"name":        ruleName,
	}
	bs, _ := json.Marshal(payload)
	c.Logger.Printf("--- Creating Rule: %s ---", ruleName)
	resp, err := c.doRequest(http.MethodPost, url, bytes.NewReader(bs))
	if err != nil {
		c.Logger.Printf("CreateEMQXRule request error: %v", err)
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	c.Logger.Printf("Status Code: %d\nResponse: %s\n", resp.StatusCode, string(body))

	// If the resource already exists, attempt to update it with PUT
	if resp.StatusCode != http.StatusCreated {
		c.Logger.Printf("Rule %s already exists, attempting to update (PUT)", ruleID)
		putURL := url + "/" + ruleID
		// try PUT to update the existing rule
		respPut, errPut := c.doRequest(http.MethodPut, putURL, bytes.NewReader(bs))
		if errPut != nil {
			c.Logger.Printf("CreateEMQXRule PUT request error: %v", errPut)
			return false, errPut
		}
		defer func() { _ = respPut.Body.Close() }()
		bodyPut, _ := io.ReadAll(respPut.Body)
		c.Logger.Printf("PUT Status Code: %d\nPUT Response: %s\n", respPut.StatusCode, string(bodyPut))
		if respPut.StatusCode >= 200 && respPut.StatusCode < 300 {
			return true, nil
		}
		return false, fmt.Errorf("PUT failed: status %d: %s", respPut.StatusCode, string(bodyPut))
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, nil
	}
	return false, fmt.Errorf("POST failed: status %d: %s", resp.StatusCode, string(body))
}

// ProcessYAMLAndSync reads the YAML pointed by env YAML_FILE_PATH and creates actions+rules.
// It mirrors the Python control flow but uses dynamic maps to avoid a large struct model.
func (c *Client) ProcessYAMLAndSync(experiment bson.M) error {
	actionsList, _ := c.GetEMQXListActions()

	nonAlpha := regexp.MustCompile(`[^a-z0-9]`)

	// extract experimentId from the passed experiment (GetCompleteExperimentById sets this)
	experimentId := getString(experiment, "experimentId")

	// devices were previously passed directly; now pull them from the experiment
	devices, _ := experiment["devices"].(bson.M)

	for deviceName, dv := range devices {
		deviceMap, ok := dv.(bson.M)
		if !ok {
			continue
		}
		deviceShort := strings.ToLower(getString(deviceMap, "shortName"))
		services, _ := deviceMap["services"].([]primitive.M)
		for _, s := range services {
			smap := s
			charList, _ := smap["characteristics"].([]primitive.M)
			for _, ch := range charList {
				chm := ch
				if sp, ok := chm["structParser"]; ok {
					name := strings.ToLower(getString(chm, "name"))
					clean := nonAlpha.ReplaceAllString(name, "")
					mqttTopic := getString(chm, "mqttTopic")
					if deviceShort == "" || clean == "" || mqttTopic == "" {
						continue
					}
					measureName := deviceShort + "_" + clean

					// build fields list
					fields := []primitive.M{}
					if farr, ok := sp.(primitive.M)["fields"].(primitive.A); ok {
						for _, fi := range farr {
							fimap := fi.(bson.M)
							fields = append(fields, fimap)
						}
					}
					// add fixed fields
					fields = append(fields, primitive.M{"name": "gatewayBattery"})
					fields = append(fields, primitive.M{"name": "rssi"})

					influxParts := []string{}
					for _, f := range fields {
						n := getString(f, "name")
						influxParts = append(influxParts, fmt.Sprintf("%s=${payload.%s}i", n, n))
					}
					// include experimentId as a fixed tag if available
					tagPrefix := measureName
					if experimentId != "" {
						tagPrefix = fmt.Sprintf("%s,experimentId=%s", measureName, experimentId)
					}
					writeSyntax := fmt.Sprintf("%s,appTagName=${payload.APP_TAG_NAME},deviceAddress=${payload.deviceAddress},deviceName=${payload.deviceName},gatewayName=${payload.gatewayName} %s",
						tagPrefix, strings.Join(influxParts, ","))
					sql := fmt.Sprintf(`SELECT * FROM "%s"`, mqttTopic)

					actionName := "action_" + measureName + "_" + experimentId
					actionDesc := fmt.Sprintf("InfluxDB action for %s - %s", deviceName, getString(chm, "name"))
					_, _ = c.CreateEMQXAction(actionName, actionDesc, writeSyntax, actionsList)

					ruleName := "rule_" + measureName + "_" + experimentId
					ruleID := "rule_id_" + measureName + "_" + experimentId
					ruleDesc := fmt.Sprintf("Rule for %s - %s", deviceName, getString(chm, "name"))
					_, _ = c.CreateEMQXRule(ruleID, ruleName, ruleDesc, sql, actionName)
				}
			}
		}

		// movesense_whiteboard measures
		if mw, ok := deviceMap["movesense_whiteboard"].(map[string]interface{}); ok {
			if measures, ok := mw["measures"].([]interface{}); ok {
				for _, m := range measures {
					mm, _ := m.(map[string]interface{})
					mname := strings.ToLower(getString(mm, "name"))
					clean := nonAlpha.ReplaceAllString(mname, "")
					mqttTopic := getString(mm, "mqttTopic")
					if deviceShort == "" || clean == "" || mqttTopic == "" {
						continue
					}
					measureName := deviceShort + "_" + clean
					actionName := "action_" + measureName
					ruleName := "rule_" + measureName
					ruleID := "rule_id_" + measureName
					ruleDesc := fmt.Sprintf("Rule for %s - %s", deviceName, getString(mm, "name"))
					actionDesc := fmt.Sprintf("InfluxDB action for %s - %s", deviceName, getString(mm, "name"))

					// jsonPayloadParser
					if jp, ok := mm["jsonPayloadParser"].(map[string]interface{}); ok {
						fields := []map[string]interface{}{}
						if farr, ok := jp["fields"].([]interface{}); ok {
							for _, fi := range farr {
								if fm, ok := fi.(map[string]interface{}); ok {
									fields = append(fields, fm)
								}
							}
						}
						// add fixed field
						fields = append(fields, map[string]interface{}{"name": "gatewayBattery", "type": "integer"})
						useJq := false
						if uj, ok := jp["use_jq"].(bool); ok {
							useJq = uj
						}

						selectParts := []string{"payload.deviceName as deviceName", "payload.gatewayName as gatewayName"}
						influxParts := []string{}
						for _, f := range fields {
							fname := getString(f, "name")
							fpath := getString(f, "path")
							ftype := getString(f, "type")
							if useJq && fpath != "" {
								selectParts = append(selectParts, fmt.Sprintf("first(jq('.%s', payload)) as %s", fpath, fname))
							} else {
								if fpath != "" {
									selectParts = append(selectParts, fmt.Sprintf("payload.%s as %s", fpath, fname))
								} else {
									selectParts = append(selectParts, fmt.Sprintf("payload as %s", fname))
								}
							}
							if ftype == "integer" {
								influxParts = append(influxParts, fmt.Sprintf("%s=${%s}i", fname, fname))
							} else {
								influxParts = append(influxParts, fmt.Sprintf("%s=${%s}", fname, fname))
							}
						}
						sql := fmt.Sprintf("SELECT %s FROM \"%s\"", strings.Join(selectParts, ", "), mqttTopic)
						writeSyntax := fmt.Sprintf("%s,deviceName=${deviceName},gatewayName=${gatewayName} %s", measureName, strings.Join(influxParts, ","))
						_, _ = c.CreateEMQXAction(actionName, actionDesc, writeSyntax, actionsList)
						_, _ = c.CreateEMQXRule(ruleID, ruleName, ruleDesc, sql, actionName)
					} else if ja, ok := mm["jsonArrayParser"].(map[string]interface{}); ok {
						// jsonArrayParser
						arrayPath := getString(ja, "arrayPath")
						fields := []map[string]interface{}{}
						if farr, ok := ja["fields"].([]interface{}); ok {
							for _, fi := range farr {
								if fm, ok := fi.(map[string]interface{}); ok {
									fields = append(fields, fm)
								}
							}
						}
						arrayAlias := "sample_item"
						doParts := []string{"payload.deviceName as deviceName", "payload.gatewayName as gatewayName"}
						influxParts := []string{}
						for _, f := range fields {
							fname := getString(f, "name")
							ftype := getString(f, "type")
							fpath := getString(f, "path")
							if fpath != "" {
								doParts = append(doParts, fmt.Sprintf("%s.%s as %s", arrayAlias, fpath, fname))
							} else {
								doParts = append(doParts, fmt.Sprintf("%s as %s", arrayAlias, fname))
							}
							if ftype == "integer" {
								influxParts = append(influxParts, fmt.Sprintf("%s=${%s}i", fname, fname))
							} else {
								influxParts = append(influxParts, fmt.Sprintf("%s=${%s}", fname, fname))
							}
						}
						sql := fmt.Sprintf("FOREACH payload.%s as %s DO %s FROM \"%s\"", arrayPath, arrayAlias, strings.Join(doParts, ", "), mqttTopic)
						writeSyntax := fmt.Sprintf("%s,deviceName=${deviceName},gatewayName=${gatewayName} %s", measureName, strings.Join(influxParts, ","))
						_, _ = c.CreateEMQXAction(actionName, actionDesc, writeSyntax, actionsList)
						_, _ = c.CreateEMQXRule(ruleID, ruleName, ruleDesc, sql, actionName)
					} else if smp, ok := mm["SingleMeasurementParser"].([]interface{}); ok {
						// SingleMeasurementParser - fields are list of maps
						selectParts := []string{"payload.deviceName as deviceName", "payload.gatewayName as gatewayName"}
						influxParts := []string{}
						for _, fi := range smp {
							if fm, ok := fi.(map[string]interface{}); ok {
								fname := getString(fm, "name")
								fpath := getString(fm, "path")
								ftype := getString(fm, "type")
								if fpath != "" {
									selectParts = append(selectParts, fmt.Sprintf("payload.%s as %s", fpath, fname))
								} else {
									selectParts = append(selectParts, fmt.Sprintf("payload as %s", fname))
								}
								if ftype == "integer" {
									influxParts = append(influxParts, fmt.Sprintf("%s=${%s}i", fname, fname))
								} else {
									influxParts = append(influxParts, fmt.Sprintf("%s=${%s}", fname, fname))
								}
							}
						}
						sql := fmt.Sprintf("SELECT %s FROM \"%s\"", strings.Join(selectParts, ", "), mqttTopic)
						writeSyntax := fmt.Sprintf("%s,deviceName=${deviceName},gatewayName=${gatewayName} %s", measureName, strings.Join(influxParts, ","))
						_, _ = c.CreateEMQXAction(actionName, actionDesc, writeSyntax, actionsList)
						_, _ = c.CreateEMQXRule(ruleID, ruleName, ruleDesc, sql, actionName)
					}
				}
			}
		}
	}

	return nil
}

// helper to safely extract string fields from map
func getString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok && v != nil {
		switch t := v.(type) {
		case string:
			return t
		case fmt.Stringer:
			return t.String()
		default:
			// attempt JSON encode -> string
			bs, err := json.Marshal(v)
			if err == nil {
				return string(bs)
			}
		}
	}
	return ""
}

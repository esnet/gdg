package secure

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/esnet/gdg/pkg/config/domain"
	"github.com/stretchr/testify/assert"
)

const inputJson = `
[
	{
		"name": "discord",
		"orgId": 1,
		"receivers": [
			{
				"settings": {
					"url": "https://www.discord.com?q=hello",
					"use_discord_username": false
				},
				"type": "discord",
				"uid": "fdxmqkyb5gl4xb"
			}
		]
	},
{
		"name": "anotherEntry",
		"orgId": 2,
		"receivers": [
			{
				"settings": {
					"url": "https://www.discord.com?q=world",
					"use_discord_username": false
				},
				"type": "discord",
				"uid": "fdxmqkyb5gl4xb"
			}
		]
	}
	{
		"name": "slack",
		"orgId": 1,
		"receivers": [
			{
				"settings": {
					"recipient": "testing",
					"token": "woot"
				},
				"type": "slack",
				"uid": "aeov0rrgij7r4a"
			}
		]
	}
]
   `

var (
	encoderFn = func(s string) (string, error) {
		return base64.StdEncoding.EncodeToString([]byte(s)), nil
	}
	decoderFn = func(s string) (string, error) {
		data, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
)

func TestUpdateJson(t *testing.T) {
	encoder := PluginCipherEncoder{
		secureFields: map[string][]string{
			"alerting": {
				"#.receivers.#.settings.url",
				"#.receivers.#.settings.password",
				"#.receivers.#.settings.token",
			},
		},
	}
	assert := assert.New(t)
	result := encoder.updateJson(domain.AlertingResource, []byte(inputJson), encoderFn)
	strResult := string(result)
	assert.False(strings.Contains(strResult, "www.discord.com"))
	assert.False(strings.Contains(strResult, "hello"))
	assert.False(strings.Contains(strResult, "world"))
	assert.False(strings.Contains(strResult, "woot"))
	// Revert the changes
	result = encoder.updateJson(domain.AlertingResource, result, decoderFn)
	strResult = string(result)
	assert.True(strings.Contains(strResult, "www.discord.com"))
	assert.True(strings.Contains(strResult, "hello"))
	assert.True(strings.Contains(strResult, "world"))
	assert.True(strings.Contains(strResult, "woot"))
	// unsupported entity
	result = encoder.updateJson(domain.DashboardResource, []byte(inputJson), encoderFn)
	strResult = string(result)
	assert.Equal(strResult, inputJson)
}

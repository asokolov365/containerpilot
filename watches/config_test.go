package watches

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asokolov365/containerpilot/surveillee"
	"github.com/asokolov365/containerpilot/tests"
	"github.com/asokolov365/containerpilot/tests/mocks"
)

var noop = &mocks.NoopDiscoveryBackend{}

func TestWatchesParse(t *testing.T) {
	data, _ := ioutil.ReadFile(fmt.Sprintf("./testdata/%s.json5", t.Name()))
	testCfg := tests.DecodeRawToSlice(string(data))
	survSvcs := surveillee.NewServices(noop, nil, nil)
	watches, err := NewConfigs(testCfg, survSvcs)
	if err != nil {
		t.Fatal(err)
	}
	assert := assert.New(t)
	assert.Equal(watches[0].serviceName, "upstreamA", "config for serviceName")
	assert.Equal(watches[0].Name, "watch.upstreamA", "config for Name")
	assert.Equal(watches[0].Poll, 11, "config for Poll")
	assert.Equal(watches[0].Tag, "dev", "config for Tag")
	assert.Equal(watches[0].DC, "", "config for DC")

	assert.Equal(watches[1].serviceName, "upstreamB", "config for serviceName")
	assert.Equal(watches[1].Name, "watch.upstreamB", "config for Name")
	assert.Equal(watches[1].Poll, 79, "config for Poll")
	assert.Equal(watches[1].DC, "us-east-1", "config for DC")
}

func TestWatchesConfigError(t *testing.T) {

	expectErr := func(test, errMsg string) {
		testCfg := tests.DecodeRawToSlice(test)
		_, err := NewConfigs(testCfg, surveillee.NewServices(nil, nil, nil))
		assert.EqualError(t, err, errMsg)
	}
	expectErr(
		`[{name: ""}]`,
		"'name' must not be blank")
	expectErr(
		`[{name: "myName"}]`,
		"watch[myName].interval must be > 0")
	expectErr(
		`[{name: "myName", interval: 10}]`,
		"watch[myName].source is consul but consul config is not defined")
	expectErr(
		`[{name: "myName", interval: "xx"}]`,
		"Watch configuration error: 1 error(s) decoding:\n\n* cannot parse '[0].interval' as int: strconv.ParseInt: parsing \"xx\": invalid syntax")
	expectErr(
		`[{name: "myName", source: "vault", interval: 10}]`,
		"watch[myName].source is vault but vault config is not defined")
}

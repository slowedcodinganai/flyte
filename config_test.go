/*
Copyright (C) 2018 Expedia Group.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type envVars map[string]string

func (e envVars) lookupEnv(key string) (string, bool) {
	val, contains := e[key]
	return val, contains
}

func newflyteEnvVars() envVars {
	return envVars{
		portEnvName:                              "80",
		tlsCertPathEnvName:                       "/path/to/tls/cert",
		tlsKeyPathEnvName:                        "/path/to/tls/key",
		mgoHostEnvName:                           "mongo:27017",
		authPolicyPathEnvName:                    "/path/to/authpolicy",
		oidcIssuerURLName:                        "dex:5559",
		oidcIssuerClientIDName:                   "example-app",
		flyteTTLEnvName:                          "86400",
		shouldDeleteDeadPacksEnvName:             "false",
		deleteDeadPacksTimeEnvName:               "10:00",
		packGracePeriodUntilDeadInSecondsEnvName: "500000",
	}
}

func TestConfigShouldReadFlyteEnvVars(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = newflyteEnvVars().lookupEnv

	c := NewConfig()

	assert.Equal(t, "80", c.Port)
	assert.Equal(t, "/path/to/tls/cert", c.TLSCertPath)
	assert.Equal(t, "/path/to/tls/key", c.TLSKeyPath)
	assert.Equal(t, "mongo:27017", c.MongoHost)
	assert.Equal(t, "/path/to/authpolicy", c.AuthPolicyPath)
	assert.Equal(t, "dex:5559", c.OidcIssuerURL)
	assert.Equal(t, "example-app", c.OidcIssuerClientID)
	assert.Equal(t, 86400, c.FlyteTTL)
	assert.Equal(t, false, c.ShouldDeleteDeadPacks)
	assert.Equal(t, "10:00", c.DeleteDeadPacksTime)
	assert.Equal(t, 500000, c.PackGracePeriodUntilDeadInSeconds)
}

func TestConfigShouldDefaultMongoHostIfNotSetAsEnvVar(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	// remove mongo host from env vars
	flyteEnvVars := newflyteEnvVars()
	delete(flyteEnvVars, mgoHostEnvName)
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()
	assert.Equal(t, "localhost:27017", c.MongoHost)
}

func TestConfigShouldDefaultPortBasedOnTLSEnabled(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	// remove port from env vars
	flyteEnvVars := newflyteEnvVars()
	delete(flyteEnvVars, portEnvName)
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()

	// secure port default
	assert.Equal(t, "8443", c.Port)
}

func TestConfigShouldDefaultPortBasedOnTLSDisabled(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	// remove port from env vars & tls vars (turns off tls)
	flyteEnvVars := newflyteEnvVars()
	delete(flyteEnvVars, portEnvName)
	delete(flyteEnvVars, tlsCertPathEnvName)
	delete(flyteEnvVars, tlsKeyPathEnvName)
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()

	// standard http port default
	assert.Equal(t, "8080", c.Port)
}

func TestConfigShouldSetDefaultTTL(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	// remove default flyte data ttl from env vars
	flyteEnvVars := newflyteEnvVars()
	delete(flyteEnvVars, flyteTTLEnvName)
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()

	// default flyte data ttl
	assert.Equal(t, oneYearInSeconds, c.FlyteTTL)
}

func TestConfigShouldSetDefaultTTLAndLogOnStringToIntConversionError(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	flyteEnvVars := newflyteEnvVars()
	flyteEnvVars[flyteTTLEnvName] = "this-string-cannot-be-converted-to-int!!!"
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()

	// default flyte data ttl
	assert.Equal(t, 31557600, c.FlyteTTL)
}

func TestConfigShouldEnableTLSOnlyIfBothKeyAndCertProvided(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	// remove tls key env var
	flyteEnvVars := newflyteEnvVars()
	delete(flyteEnvVars, tlsKeyPathEnvName)
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()

	assert.False(t, c.requireTLS())
}

func TestConfigShouldEnableAuthIfOIDCIssuerIsProvidedAndPolicyPathIsProvided(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	// policy path, oidc issuer uri/config id is set
	flyteEnvVars := newflyteEnvVars()
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()

	assert.True(t, c.requireAuth())
}

func TestConfigShouldNotEnableAuthIfPolicyPathNotProvided(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	// remove policy path env var
	flyteEnvVars := newflyteEnvVars()
	delete(flyteEnvVars, authPolicyPathEnvName)
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()

	assert.False(t, c.requireAuth())
}

func TestConfigShouldNotEnableAuthIfOIDCIssuerUriNotProvided(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	// remove oidc issuer uri env var
	flyteEnvVars := newflyteEnvVars()
	delete(flyteEnvVars, oidcIssuerURLName)
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()

	assert.False(t, c.requireAuth())
}

func TestConfigShouldNotEnableAuthIfOIDCIssuerConfigIdNotProvided(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	// remove oidc issuer config id env var
	flyteEnvVars := newflyteEnvVars()
	delete(flyteEnvVars, oidcIssuerClientIDName)
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()

	assert.False(t, c.requireAuth())
}

func TestConfigShouldSetDefaultForShouldDeleteDeadPacksIfNotSetAsEnvVar(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	// remove shouldDeleteDeadPacksEnvName from env vars
	flyteEnvVars := newflyteEnvVars()
	delete(flyteEnvVars, shouldDeleteDeadPacksEnvName)
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()
	assert.Equal(t, false, c.ShouldDeleteDeadPacks)
}

func TestConfigShouldSetDefaultAndLogErrorForShouldDeleteDeadPacksIfEnvVarValueSetIsNotBoolean(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	// change shouldDeleteDeadPacksEnvName in env vars to non boolean value
	flyteEnvVars := newflyteEnvVars()
	flyteEnvVars[shouldDeleteDeadPacksEnvName] = "6372gyd"
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()
	assert.Equal(t, false, c.ShouldDeleteDeadPacks)
}

func TestConfigShouldSetDefaultDeleteDeadPacksTimeIfNotSetAsEnvVar(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	// remove deleteDeadPacksTimeEnvName from env vars
	flyteEnvVars := newflyteEnvVars()
	delete(flyteEnvVars, deleteDeadPacksTimeEnvName)
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()
	assert.Equal(t, defaultDeleteDeadPacksTime, c.DeleteDeadPacksTime)
}

var invalidTimes = []struct {
	invalidValue string
	errorMsg     string
}{
	{"1100", "FLYTE_DELETE_DEAD_PACKS_AT_HH_COLON_MM env is invalid, using default 23:00, error: parsing time \"1100\" as \"15:04\": cannot parse \"00\" as \":\""},
	{"AA:00", "FLYTE_DELETE_DEAD_PACKS_AT_HH_COLON_MM env is invalid, using default 23:00, error: parsing time \"AA:00\" as \"15:04\": cannot parse \"AA:00\" as \"15\""},
	{"23:AA", "FLYTE_DELETE_DEAD_PACKS_AT_HH_COLON_MM env is invalid, using default 23:00, error: parsing time \"23:AA\" as \"15:04\": cannot parse \"AA\" as \"04\""},
	{"24:00", "FLYTE_DELETE_DEAD_PACKS_AT_HH_COLON_MM env is invalid, using default 23:00, error: parsing time \"24:00\": hour out of range"},
	{"23:60", "FLYTE_DELETE_DEAD_PACKS_AT_HH_COLON_MM env is invalid, using default 23:00, error: parsing time \"23:60\": minute out of range"},
}

func TestConfigShouldSetDefaultDeleteDeadPacksTimeAndLogErrorIfTimeIsInInvalidFormat(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)

	for _, tt := range invalidTimes {
		// change deleteDeadPacksTimeEnvName in env vars
		flyteEnvVars := newflyteEnvVars()
		flyteEnvVars[deleteDeadPacksTimeEnvName] = tt.invalidValue
		lookupEnv = flyteEnvVars.lookupEnv

		c := NewConfig()
		assert.Equal(t, defaultDeleteDeadPacksTime, c.DeleteDeadPacksTime)
	}
}

func TestConfigShouldSetDefaultPackGracePeriodUntilDeadInSeconds(t *testing.T) {
	defer func(oldFileExists func(string) bool) { fileExists = oldFileExists }(fileExists)
	fileExists = func(string) bool { return true }

	// remove default flyte pack grace period from env vars
	flyteEnvVars := newflyteEnvVars()
	delete(flyteEnvVars, packGracePeriodUntilDeadInSecondsEnvName)
	defer func(oldGetEnv func(string) (string, bool)) { lookupEnv = oldGetEnv }(lookupEnv)
	lookupEnv = flyteEnvVars.lookupEnv

	c := NewConfig()

	// default flyte pack grace period in seconds
	assert.Equal(t, oneWeekInSeconds, c.PackGracePeriodUntilDeadInSeconds)
}

package config

import "testing"

func TestSetupAppConfig(t *testing.T) {
	app := SetupAppConfig(false)

	if app == nil {
		t.Error("Error in creating App Config")
	}
}

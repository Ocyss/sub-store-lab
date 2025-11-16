package static

import _ "embed"

//go:embed script.js
var ScriptJs string

//go:embed mihomo-config.yaml
var MihomoConfigYamlByte []byte

//go:embed override.yaml
var OverrideYamlByte []byte

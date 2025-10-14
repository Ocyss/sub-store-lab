package static

import _ "embed"

//go:embed script.js
var ScriptJs string

//go:embed clash.yml
var ClashYml []byte

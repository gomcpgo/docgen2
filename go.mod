module github.com/savant/mcp-servers/docgen2

go 1.23

require (
	github.com/gomcpgo/mcp v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/gomcpgo/mcp => ../mcp

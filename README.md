# modbus-cli

This is a basic tool to query modbus elements.

Modbus is a protocol designed for communicating between electronic devices designed in the 1970s that continues to be used today despite modern alternatives existing.

Modbus can be found in a wide variety of devices, my particular use case is around solar power systems which often have an undocumented modbus interface designed to be used with the manufacturer's add-on devices, which are universally awful.

This tool allows you to query a single address or address block at a time, and output the results in a few formats. You could use this within a script to collect many data points however it is more intended to be a tool for scanning around and discovering those data points to be collected by another more appropriate tool.

# Installation

If you have a Go compiler installed, simply
`go get github.com/squash/modbus-cli`
This will download, complile, and install to your GOROOT/bin directory.

# Using modbus-cli

Run modbus-cli with -h for a list of options.

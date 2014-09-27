rm -rf test_config && mkdir test_config
echo $'#!/bin/bash\n\n'export PROTO_PATH='"'.:$(pwd | sed "s/\/github\.com.*//g")'"' > test_config/config
echo $'package config\n\n'const ProtoPath string = '"'.:$(pwd | sed "s/\/github\.com.*//g")'"' > test_config/config.go

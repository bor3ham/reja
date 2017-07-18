#!/bin/ash
rm **/*easyjson.go
easyjson -all -no_std_marshalers \
	schema/result.go \
	schema/instance_pointer.go \
	schema/page.go

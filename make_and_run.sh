#!/bin/bash
goimports -w . && go fmt && go build && ./wedge

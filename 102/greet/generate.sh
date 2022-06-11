#!/bin/bash

protoc pb/greet.proto --go_out=plugins=grpc:./pb

#!/bin/bash

python -m grpc_tools.protoc -I protobufs --python_out=./recommendations --grpc_python_out=./recommendations protobufs/recommendations.proto
python -m grpc_tools.protoc -I protobufs --python_out=./marketplace --grpc_python_out=./marketplace protobufs/recommendations.proto

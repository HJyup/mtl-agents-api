#!/usr/bin/env python3
import os
import subprocess
import sys

ROOT_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..'))
PROTO_DIR = os.path.join(ROOT_DIR, 'common', 'api')
OUTPUT_DIR = os.path.join(ROOT_DIR, 'agent', 'protos')

def generate_proto_files():
    print(f"Generating protobuf files from {PROTO_DIR} to {OUTPUT_DIR}")

    os.makedirs(OUTPUT_DIR, exist_ok=True)

    init_file = os.path.join(OUTPUT_DIR, '__init__.py')
    if not os.path.exists(init_file):
        with open(init_file, 'w') as f:
            f.write("# Generated protobuf modules\n")
    
    proto_files = [
        os.path.join(PROTO_DIR, 'agent.proto'),
        os.path.join(PROTO_DIR, 'config.proto'),
    ]
    
    for proto_file in proto_files:
        if not os.path.exists(proto_file):
            print(f"Error: Proto file {proto_file} does not exist.")
            sys.exit(1)
        
        cmd = [
            'python3', '-m', 'grpc_tools.protoc',
            f'--proto_path={PROTO_DIR}',
            f'--python_out={OUTPUT_DIR}',
            f'--grpc_python_out={OUTPUT_DIR}',
            proto_file
        ]
        
        try:
            subprocess.run(cmd, check=True)
            print(f"Successfully generated Python files for {proto_file}")
        except subprocess.CalledProcessError as e:
            print(f"Error generating protobuf files: {e}")
            sys.exit(1)

    with open(init_file, 'w') as f:
        f.write("# Generated protobuf modules\n")
        f.write("from agent.protos import agent_pb2, agent_pb2_grpc\n")
        f.write("from agent.protos import config_pb2, config_pb2_grpc\n")

if __name__ == "__main__":
    generate_proto_files() 
#!/usr/bin/env python3
import os
import subprocess
import sys
import re

ROOT_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..'))
PROTO_DIR = os.path.join(ROOT_DIR, 'common', 'api')
OUTPUT_DIR = os.path.join(ROOT_DIR, 'agent', 'protos')
PACKAGE = 'agent.protos'

def generate_proto_files():
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    with open(os.path.join(OUTPUT_DIR, '__init__.py'), 'w') as f:
        f.write("# Generated protobuf modules\n")

    proto_files = [
        os.path.join(PROTO_DIR, 'agent.proto'),
        os.path.join(PROTO_DIR, 'config.proto'),
    ]

    for proto_file in proto_files:
        if not os.path.exists(proto_file):
            print(f"Proto file {proto_file} not found")
            sys.exit(1)

        proto_name = os.path.basename(proto_file).split('.')[0]

        cmd = [
            'python3', '-m', 'grpc_tools.protoc',
            f'-I{PROTO_DIR}',
            f'--python_out={OUTPUT_DIR}',
            f'--pyi_out={OUTPUT_DIR}',
            f'--grpc_python_out={OUTPUT_DIR}',
            proto_file
        ]

        try:
            result = subprocess.run(cmd, check=True, capture_output=True, text=True)
            print(f"Successfully generated {proto_name} protos")
            if result.stdout:
                print(f"Output: {result.stdout}")

            fix_imports(proto_name)
        except subprocess.CalledProcessError as e:
            print(f"Error: {e}")
            print(f"STDOUT: {e.stdout}")
            print(f"STDERR: {e.stderr}")
            sys.exit(1)

def fix_imports(proto_name):
    """Fix imports in generated files to use the custom package path."""
    pb2_file = os.path.join(OUTPUT_DIR, f"{proto_name}_pb2.py")
    grpc_file = os.path.join(OUTPUT_DIR, f"{proto_name}_pb2_grpc.py")

    if os.path.exists(grpc_file):
        with open(grpc_file, 'r') as f:
            content = f.read()

        modified_content = re.sub(
            r'import ' + proto_name + r'_pb2 as ' + proto_name + r'__pb2',
            f'from {PACKAGE} import {proto_name}_pb2 as {proto_name}__pb2',
            content
        )

        with open(grpc_file, 'w') as f:
            f.write(modified_content)
            print(f"Fixed imports in {grpc_file}")

    if os.path.exists(pb2_file):
        print(f"Generated {pb2_file}")

if __name__ == "__main__":
    generate_proto_files()
import yaml
import os

def generate_workflow():
    with open("deploy_config.yaml", "r") as f:
        config = yaml.safe_load(f)

    services = config.get("services", {})
    service_paths = set()
    for name, details in services.items():
        path = details.get("path")
        if path:
            dir_path = os.path.dirname(path)
            service_paths.add(dir_path)

    workflow_content = f"""name: Backend Tests (Generated)

on:
  push:
    paths:
      - 'services/**'
      - 'pkg/**'
      - 'e2e_tests/backend_integration/**'
      - '.github/workflows/backend_tests.yml'
  pull_request:
    paths:
      - 'services/**'
      - 'pkg/**'
      - 'e2e_tests/backend_integration/**'
      - '.github/workflows/backend_tests.yml'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Check out repository
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Verify Dependencies
        run: go mod download

      - name: Skip External Dependencies
        run: echo "Skipping Prometheus and Jaeger as requested for basic backend tests."

"""

    workflow_content += "      - name: Run Unit Tests\n        run: |\n"
    for path in sorted(service_paths):
        workflow_content += f"          echo \"Testing {path}...\"\n"
        workflow_content += f"          (cd {path} && go test ./...)\n"

    workflow_content += """
      - name: Run Integration Tests
        run: |
          go test ./e2e_tests/backend_integration/...
"""

    os.makedirs(".github/workflows", exist_ok=True)
    with open(".github/workflows/backend_tests.yml", "w") as f:
        f.write(workflow_content)

    print("Generated .github/workflows/backend_tests.yml successfully.")

if __name__ == "__main__":
    generate_workflow()

import os
import sys
import yaml

CONFIG_FILE = 'deploy_config.yaml'
PROCFILE = 'Procfile'
DOCKER_COMPOSE_FILE = 'docker-compose.test.yml'

def load_config():
    with open(CONFIG_FILE, 'r') as f:
        return yaml.safe_load(f)

def is_config_newer(target_file):
    if not os.path.exists(target_file):
        return True
    return os.path.getmtime(CONFIG_FILE) > os.path.getmtime(target_file)

def generate_procfile(config):
    lines = []

    for name, dep in config.get('external_dependencies', {}).items():
        cmd = dep.get('command')
        if cmd:
            lines.append(f"{name}: {cmd}")

    for name, service in config.get('services', {}).items():
        path = service.get('path')
        if path:
            lines.append(f"{name}: OTEL_EXPORTER_OTLP_INSECURE=true go run {path}")

    with open(PROCFILE, 'w') as f:
        f.write('\n'.join(lines) + '\n')
    print(f"Generated {PROCFILE}")

def generate_docker_compose(config):
    compose = {
        'version': '3.8',
        'services': {}
    }

    for name, dep in config.get('external_dependencies', {}).items():
        image = dep.get('docker_image')
        if not image:
            continue
        svc = {
            'image': image,
            'ports': dep.get('ports', []),
        }
        compose['services'][name] = svc

    for name, service in config.get('services', {}).items():
        path = service.get('path')
        # Here we define the service definition assuming we would mount the code and use an image with go
        svc = {
            'image': 'golang:1.21',
            'volumes': ['.:/app'],
            'working_dir': '/app',
            'environment': {
                'OTEL_EXPORTER_OTLP_INSECURE': 'true',
            },
            'command': f"go run {path}",
            'ports': [f"{service['port']}:{service['port']}"],
            'depends_on': list(config.get('external_dependencies', {}).keys())
        }
        # In this simplistic docker-compose, we mount everything. In reality we might use a Dockerfile.
        # But this is fine for tests.
        compose['services'][name] = svc

    with open(DOCKER_COMPOSE_FILE, 'w') as f:
        yaml.dump(compose, f, sort_keys=False)
    print(f"Generated {DOCKER_COMPOSE_FILE}")

def main():
    if not os.path.exists(CONFIG_FILE):
        print(f"Error: {CONFIG_FILE} not found.")
        sys.exit(1)

    config = load_config()

    if is_config_newer(PROCFILE):
        generate_procfile(config)
    else:
        print(f"{PROCFILE} is up to date.")

    if is_config_newer(DOCKER_COMPOSE_FILE):
        generate_docker_compose(config)
    else:
        print(f"{DOCKER_COMPOSE_FILE} is up to date.")

if __name__ == '__main__':
    main()

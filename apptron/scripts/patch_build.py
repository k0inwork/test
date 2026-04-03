import re, sys, os

def process_file(filepath, callback):
    if not os.path.exists(filepath):
        print(f"Skipping {filepath} (not found)")
        return
    try:
        with open(filepath, "r") as f:
            content = f.read()
        new_content = callback(content)
        with open(filepath, "w") as f:
            f.write(new_content)
        print(f"Patched {filepath}")
    except Exception as e:
        print(f"Error patching {filepath}: {e}")

def patch_dockerfile(content):
    # Strip the problematic stages to bypass i386 alpine dependencies
    content = re.sub(r'FROM --platform=\$LINUX_386.*?FROM alpine:\$ALPINE_VERSION AS bundle-base',
                     'FROM alpine:$ALPINE_VERSION AS bundle-base',
                     content, flags=re.DOTALL)

    # Remove 'bundle-sys' block completely
    content = re.sub(r'FROM bundle-base AS bundle-sys.*?FROM golang:\$GO_VERSION-alpine AS worker-build',
                     'FROM golang:$GO_VERSION-alpine AS worker-build',
                     content, flags=re.DOTALL)

    # Replace copying from bundle-sys with local bundles
    content = re.sub(r'COPY --from=bundle-sys /bundles/\* /bundles/', 'COPY assets/bundles/* /bundles/', content)
    return content

def patch_makefile(content, pum_admin_assets):
    # 1. Update vscode extraction target to prevent curl errors and unzip natively
    # Use the local cache if available, else download cleanly.
    content = content.replace('curl -sL $(VSCODE_URL) -o assets/vscode.zip',
                              f'if [ -f "{pum_admin_assets}/vscode.zip" ]; then cp "{pum_admin_assets}/vscode.zip" assets/vscode.zip; else curl -sL $(VSCODE_URL) -o assets/vscode.zip; fi')

    # 2. Prevent Docker from failing when downloading wanix by simply pulling from github release directly
    content = content.replace('$(DOCKER_CMD) rm -f apptron-wanix', f'cp "{pum_admin_assets}/wanix.min.js" assets/wanix.min.js')
    content = content.replace('$(DOCKER_CMD) pull --platform linux/amd64 ghcr.io/tractordev/wanix:runtime', f'cp "{pum_admin_assets}/wanix.min.js" assets/wanix.js')

    # Remove docker commands that are no longer needed
    content = re.sub(r'^\s*\$\(DOCKER_CMD\) create --name apptron-wanix.*$', '', content, flags=re.MULTILINE)
    content = re.sub(r'^\s*\$\(DOCKER_CMD\) cp apptron-wanix.*$', '', content, flags=re.MULTILINE)

    return content

if __name__ == "__main__":
    pum_admin_assets = os.getenv("PUM_ADMIN_ASSETS", "")
    process_file("worker/Dockerfile", patch_dockerfile)
    process_file("Makefile", lambda c: patch_makefile(c, pum_admin_assets))

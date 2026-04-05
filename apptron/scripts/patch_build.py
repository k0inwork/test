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
    # Completely replace the Dockerfile for local/CI Phase 2 testing to bypass
    # Docker Hub rate limits and local Alpine overlayfs errors, and drastically speed up builds.
    # We compile the worker locally in run_phase2.sh and just COPY it here.
    return """FROM scratch AS worker
COPY assets/bundles/* /bundles/
COPY bin/worker /worker
EXPOSE 8080
CMD ["/worker"]
"""

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
    process_file("Dockerfile", patch_dockerfile)
    process_file("Makefile", lambda c: patch_makefile(c, pum_admin_assets))

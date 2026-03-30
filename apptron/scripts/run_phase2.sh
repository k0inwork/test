#!/bin/bash
set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../.." && pwd)
APPTRON_DIR="$REPO_ROOT/apptron/apptron"

if [ ! -d "$APPTRON_DIR" ]; then
    echo "Apptron worker source not found. Cloning tractordev/apptron..."
    git clone --depth 1 https://github.com/tractordev/apptron "$APPTRON_DIR"
fi

echo "Setting up Apptron worker..."
cd "$APPTRON_DIR"

if [ ! -f .env.local ]; then
    cp .env.example .env.local
fi

# In Phase 2 mock mode, we want the frontend to skip login as well
# by matching the `authUrl === "/auth"` condition in apptron.js
if grep -q "AUTH_URL=https://ad6044b5-53c2-4cb5-8542-9fdaef75f771.hanko.io" .env.local; then
    echo "Updating .env.local to mock frontend auth..."
    sed -i.bak 's|AUTH_URL=https://ad6044b5-53c2-4cb5-8542-9fdaef75f771.hanko.io|AUTH_URL=/auth|g' .env.local
fi

# Ensure clean state to apply latest patches properly
git restore worker/src/auth.ts assets/lib/apptron.js assets/signin.html worker/src/worker.ts 2>/dev/null || true

# Mock auth dynamically in the worker source
echo "Mocking Auth in worker/src/auth.ts..."

# Also patch the frontend to skip login
echo "Patching frontend assets to bypass Hanko login..."
cat << 'EOF' > patch_auth.py
import re, sys

with open("assets/lib/apptron.js", "r") as f:
    content = f.read()

pattern = r'(if \(!getMeta\("auth-url"\)\) \{\n\s*throw new Error\("auth-url meta tag not found"\);\n\s*\})'
replacement = """const authUrl = getMeta("auth-url");
if (!authUrl) {
    throw new Error("auth-url meta tag not found");
}

if (authUrl === "/auth") {
    document.cookie = "hanko=header.eyJzdWIiOiIxIiwidXNlcm5hbWUiOiJhZG1pbiJ9.signature; path=/";
    const mockSession = { is_valid: true, claims: { username: "admin" } };
    window.session = mockSession;
    auth = {
        getSessionToken: () => "header.eyJzdWIiOiIxIiwidXNlcm5hbWUiOiJhZG1pbiJ9.signature",
        session: { get: () => ({ jwt: "header.eyJzdWIiOiIxIiwidXNlcm5hbWUiOiJhZG1pbiJ9.signature" }) },
        getUser: async () => ({ id: "1", username: "admin", email: "admin@example.com" }),
        validatedSession: Promise.resolve(mockSession),
        validateSession: async () => mockSession,
        onUserDeleted: () => {}, onSessionCreated: () => {}, onSessionExpired: () => {},
        onUserLoggedOut: () => {}, onBeforeStateChange: () => {}, onAfterStateChange: () => {}
    };
    auth.validatedSession.then(session => {
        if (session.is_valid) console.log("valid mock session for user", session.claims.username);
    });
    return auth;
}"""

content = re.sub(pattern, replacement, content)
content = content.replace('register(getMeta("auth-url")', 'register(authUrl')

with open("assets/lib/apptron.js", "w") as f:
    f.write(content)
EOF
python3 patch_auth.py
rm patch_auth.py


# Also patch signin.html, worker.ts, and auth.ts reliably using Python
cat << 'EOF' > patch_others.py
import re, sys

def process_file(filepath, callback):
    try:
        with open(filepath, "r") as f:
            content = f.read()
        new_content = callback(content)
        with open(filepath, "w") as f:
            f.write(new_content)
    except Exception as e:
        print(f"Error patching {filepath}: {e}"); sys.exit(1)

# 2. Patch signin.html
def patch_signin(content):
    return content.replace('<hanko-auth></hanko-auth>', """<script>
if (document.querySelector("meta[name='auth-url']")?.content === "/auth") {
    document.write("<p>Bypassing login...</p>");
} else {
    document.write("<hanko-auth></hanko-auth>");
}
</script>""")

# 3. Patch worker.ts
def patch_worker(content):
    pattern = r'(if \(url\.pathname === "/" && req\.method === "GET"\) \{)(.*?)(^\s*\})'
    replacement = r"""    if (url.pathname === "/" && req.method === "GET") {
        await ensureSystemDirs(req, env);
        if (await validateToken(env.AUTH_URL, ctx.tokenRaw)) {
            url.pathname = "/dashboard";
            return Response.redirect(url.toString(), 307);
        }
        return redirectToSignin(env, url);
    }"""
    content = re.sub(pattern, replacement, content, flags=re.MULTILINE|re.DOTALL)

    # Add ctx.userDomain check for /edit/ block and fix missing user["id"] check
    pattern2 = r'(\/\/ <username>\.apptron\.dev\/<mode>\/<env-name>\s*)(if \(url\.pathname\.startsWith\("/edit/"\)\|\|url\.pathname\.startsWith\("/console/"\)\) \{)'
    replacement2 = r'\1if (ctx.userDomain && (url.pathname.startsWith("/edit/")||url.pathname.startsWith("/console/"))) {'
    content = re.sub(pattern2, replacement2, content)

    pattern3 = r'(const user = await req\.json\(\);\s*)(const usrResp = await putdir\(req, env, `/usr/\$\{user\["user_id"\]\}`,\s*\{\s*"username": user\["username"\],\s*\}\);)'
    replacement3 = r'\1const userId = user["id"] || user["user_id"];\n            const usrResp = await putdir(req, env, `/usr/${userId}`, {\n                "username": user["username"],\n            });'
    content = re.sub(pattern3, replacement3, content)

    pattern4 = r'("uuid": user\["user_id"\],)'
    replacement4 = r'"uuid": userId,'
    content = re.sub(pattern4, replacement4, content)

    return content

# 4. Patch auth.ts
def patch_auth(content):
    pattern = r'(export async function validateToken\(.*?\): Promise<boolean> \{)(.*?)(
\})'
    replacement = r'''export async function validateToken(hankoApiUrl: string, token: string): Promise<boolean> {
// MOCK AUTH: Always return true
return true;
}'''
    return re.sub(pattern, replacement, content, flags=re.MULTILINE|re.DOTALL)

# 5. Patch bundle loading in frontend
def patch_bundle(content):
    return content.replace('getBundle("/bundles/sys.tar.gz");', 'getBundle("/bundles/sys.tar.gz");
    getBundle("/bundles/pum.tar.gz");')

# Execute
process_file("assets/signin.html", patch_signin)
process_file("assets/signin.html", patch_bundle)
process_file("assets/dashboard.html", patch_bundle)
process_file("worker/src/worker.ts", patch_worker)
process_file("worker/src/auth.ts", patch_auth)
EOF
python3 patch_others.py
rm patch_others.py

# Patch boot.go to fix nil OPFS panic
echo "Patching boot.go to filter nil OPFS..."
cat << 'EOF' > patch_boot.py
import re, sys

with open("boot.go", "r") as f:
    content = f.read()

replacement = """	webMod := web.New(k)
	for name, subfs := range webMod {
		if subfs == nil {
			delete(webMod, name)
		}
	}
	k.AddModule("#web", webMod)"""

content = content.replace('k.AddModule("#web", web.New(k))', replacement)

with open("boot.go", "w") as f:
    f.write(content)
EOF
python3 patch_boot.py
rm patch_boot.py

# Always start from a clean Makefile to ensure patches apply correctly on multiple runs
git checkout Makefile 2>/dev/null || true
git checkout Dockerfile 2>/dev/null || true

# Patch the worker Dockerfile so it uses the host's pre-built bundles rather than rebuilding them
# This fixes busybox tar stripping './' prefixes AND prevents a 5-minute Docker rebuild on every wrangler startup
echo "Patching Dockerfile to copy pre-built bundles..."
cat << 'EOF' > patch_dockerfile.py
import re

with open("Dockerfile", "r") as f:
    content = f.read()

# Replace the worker container copying bundles from multi-stage builds
# with copying them directly from the host's assets/bundles/ directory
content = re.sub(r'COPY --from=bundle-sys /bundles/\* /bundles/', 'COPY assets/bundles/sys.tar.gz /bundles/sys.tar.gz', content)

with open("Dockerfile", "w") as f:
    f.write(content)
EOF
python3 patch_dockerfile.py
rm patch_dockerfile.py

# To properly and cleanly resolve missing components
echo "Patching Makefile to download dependencies cleanly..."

# Use absolute path for copying to ensure it works reliably regardless of working directory
PUM_ADMIN_ASSETS="$REPO_ROOT/apptron/cmd/pum-admin/assets"

# 1. Update vscode extraction target to prevent curl errors and unzip natively
# Use the local cache if available, else download cleanly.
sed -i.bak "s|curl -sL \$(VSCODE_URL) -o assets/vscode.zip|if [ -f \"$PUM_ADMIN_ASSETS/vscode.zip\" ]; then cp \"$PUM_ADMIN_ASSETS/vscode.zip\" assets/vscode.zip; else curl -sL \$(VSCODE_URL) -o assets/vscode.zip; fi|g" Makefile

# 2. Prevent Docker from failing when downloading wanix by simply pulling from github release directly, instead of extracting docker blobs
sed -i.bak "s|\$(DOCKER_CMD) rm -f apptron-wanix|cp \"$PUM_ADMIN_ASSETS/wanix.min.js\" assets/wanix.min.js|g" Makefile
sed -i.bak "s|\$(DOCKER_CMD) pull --platform linux/amd64 ghcr.io/tractordev/wanix:runtime|cp \"$PUM_ADMIN_ASSETS/wanix.min.js\" assets/wanix.js|g" Makefile
sed -i.bak '/$(DOCKER_CMD) create --name apptron-wanix.*/d' Makefile
sed -i.bak '/$(DOCKER_CMD) cp apptron-wanix.*/d' Makefile

# Copy the checked-in base sys.tar.gz bundle from the repo to the worker assets BEFORE make all
echo "Ensuring base sys.tar.gz is present from repository..."
mkdir -p assets/bundles
if [ ! -f "assets/bundles/sys.tar.gz" ]; then
    cp "$REPO_ROOT/apptron/assets/bundles/sys.tar.gz" "assets/bundles/sys.tar.gz"
fi

# Only run make clean and make all if assets/wanix.wasm is missing
if [ ! -f "assets/wanix.wasm" ]; then
    echo "Base Apptron assets not found. Running make clean and make all..."
    make clean
    # Ensure wrangler is installed locally in the worker package
    echo "Installing worker dependencies..."
    cd worker && npm ci && cd ..

    # make clean deletes sys.tar.gz, so copy it back before make all runs
    mkdir -p assets/bundles
    cp "$REPO_ROOT/apptron/assets/bundles/sys.tar.gz" "assets/bundles/sys.tar.gz"

    make all
else
    echo "Base Apptron assets found. Skipping make all to speed up startup."
    echo "Installing worker dependencies..."
    cd worker && npm ci && cd ..
fi

# Check if CLI needs rebuilding by comparing timestamps
CLI_CHANGED=false
if [ ! -f "$REPO_ROOT/build/distro/bundle/bin/pum.wasm" ]; then
    CLI_CHANGED=true
elif [ -n "$(find "$REPO_ROOT/apptron/services/pum-cli" -newer "$REPO_ROOT/build/distro/bundle/bin/pum.wasm" 2>/dev/null | head -n 1)" ]; then
    CLI_CHANGED=true
fi


if [ "$CLI_CHANGED" = true ] || [ ! -f "assets/bundles/pum.tar.gz" ]; then
    echo "pum-cli source changed or pum.tar.gz missing, building..."
    (cd "$REPO_ROOT" && bash "apptron/scripts/build_distro.sh")
fi

echo "Re-packing sys.tar.gz using Python tarfile to guarantee precise ./ paths for Wanix..."
TMP_BUNDLE_DIR=$(mktemp -d)
tar -xzf "assets/bundles/sys.tar.gz" -C "$TMP_BUNDLE_DIR"
TARGET_TAR="$(pwd)/assets/bundles/sys.tar.gz"

cat << 'PY_TAR' > repack.py
import tarfile
import os

src = "$TMP_BUNDLE_DIR"
target = "$TARGET_TAR"

with tarfile.open(target, "w:gz") as tar:
    # Add root directory explicitly
    ti = tarfile.TarInfo(name="./")
    ti.type = tarfile.DIRTYPE
    ti.mode = 0o755
    tar.addfile(ti)

    for root, dirs, files in os.walk(src):
        for d in dirs:
            full_path = os.path.join(root, d)
            rel_path = os.path.relpath(full_path, src)
            arcname = "./" + rel_path + "/"
            tar.add(full_path, arcname=arcname, recursive=False)
        for f in files:
            full_path = os.path.join(root, f)
            rel_path = os.path.relpath(full_path, src)
            arcname = "./" + rel_path
            tar.add(full_path, arcname=arcname, recursive=False)
PY_TAR

sed -i "s|\$TMP_BUNDLE_DIR|$TMP_BUNDLE_DIR|g" repack.py
sed -i "s|\$TARGET_TAR|$TARGET_TAR|g" repack.py
python3 repack.py
rm repack.py
rm -rf "$TMP_BUNDLE_DIR"

echo "Copying sys.tar.gz and pum.tar.gz into worker assets..."

mkdir -p assets/bundles
cp "$REPO_ROOT/apptron/assets/bundles/sys.tar.gz" "assets/bundles/sys.tar.gz"
cp "$REPO_ROOT/apptron/cmd/pum-admin/assets/bundles/pum.tar.gz" "assets/bundles/pum.tar.gz"

echo "Starting Phase 2 Worker in dev mode..."
echo "Mock Auth is ENABLED in apptron/worker/src/auth.ts."
echo "You can access the environment at http://localhost:8788"


echo "Clearing Wrangler cache..."
rm -rf .wrangler
docker image prune -f --filter "label=org.opencontainers.image.title=worker"

export CI=true
export WRANGLER_SEND_METRICS=false
cd worker && npx wrangler dev --port=8788 --log-level=none

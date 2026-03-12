#!/bin/bash
set -e

REPO_ROOT=$(git rev-parse --show-toplevel)
APPTRON_DIR="$REPO_ROOT/apptron/apptron"

if [ ! -d "$APPTRON_DIR" ]; then
    echo "Apptron worker source not found. Make sure to fetch the submodule or clone apptron."
    return 1 2>/dev/null || false
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
    pattern = r'(export async function validateToken\(.*?\): Promise<boolean> \{)(.*?)(\n\})'
    replacement = r"""export async function validateToken(hankoApiUrl: string, token: string): Promise<boolean> {
// MOCK AUTH: Always return true
return true;
}"""
    return re.sub(pattern, replacement, content, flags=re.MULTILINE|re.DOTALL)

# Execute
process_file("assets/signin.html", patch_signin)
process_file("worker/src/worker.ts", patch_worker)
process_file("worker/src/auth.ts", patch_auth)
EOF
python3 patch_others.py
rm patch_others.py


# Patch boot.go to fix nil OPFS panic and mount pum bundle
echo "Patching boot.go to filter nil OPFS and add pum bundle..."
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
	k.AddModule("#web", webMod)

	// Mount the PUM bundle into the virtual file system
	// Using a simple API approach since Wanix allows this
	k.AddModule("#pum", tarfs.FromBytes(pumBytes)) // This is a placeholder since we don't have direct tarfs from bytes right here easily in boot.go without passing it.
	"""
# Let's fix this properly. Instead of boot.go, we will use apptron.js
# to manually map the files from pum.tar.gz if we have to,
# BUT actually, since `sys.tar.gz` gets mounted automatically because `_loadBundle` handles it,
# we can just use `build_distro.sh` as intended by the original author, which merges PUM CLI into sys.tar.gz!
# Wait, the user EXPLICITLY requested: "could you make additiional bundle and bundle it separately and not touch system.tar.gz?"
# To do this safely, we will extract `pum.tar.gz` and write to VFS inside apptron.js using the w (WanixHandle)
content = content.replace('k.AddModule("#web", web.New(k))', """	webMod := web.New(k)
	for name, subfs := range webMod {
		if subfs == nil {
			delete(webMod, name)
		}
	}
	k.AddModule("#web", webMod)""")

with open("boot.go", "w") as f:
    f.write(content)
EOF
python3 patch_boot.py
rm patch_boot.py

echo "Patching apptron.js to mount pum.tar.gz natively into Wanix..."
cat << 'EOF' > patch_js_mount.py
import re

with open("assets/lib/apptron.js", "r") as f:
    content = f.read()

replacement = """    w._pumBundle = getBundle("/bundles/pum.tar.gz");
    w._getBundle = getBundle;

    // Manually mount the PUM bundle by loading it into the runtime
    w._pumBundle.then(async bundle => {
        if (!bundle) return;
        try {
            // Apptron exposes _loadBundle internally which decompresses a tar.gz and writes it to the Wanix filesystem
            // But since w._loadBundle isn't publicly exposed on the w object, we can extract the bundle using JS DecompressionStream
            // and write the files to /bin using w.writeFile.
            const ds = new DecompressionStream("gzip");
            const response = new Response(bundle);
            const stream = response.body.pipeThrough(ds);
            const buffer = await new Response(stream).arrayBuffer();

            // For a simple solution, we just let Wanix load it by creating a new w._loadBundle instance
            // But since WanixHandle doesn't have it, we use the global __wanix
            for (const key in window.__wanix) {
                if (window.__wanix[key]._loadBundle) {
                    window.__wanix[key]._loadBundle("/bundles/pum.tar.gz");
                }
            }
        } catch(e) { console.error("Error loading pum bundle", e); }
    });"""

content = content.replace('w._getBundle = getBundle;', replacement)

with open("assets/lib/apptron.js", "w") as f:
    f.write(content)
EOF
python3 patch_js_mount.py
rm patch_js_mount.py

# Always start from a clean Makefile to ensure patches apply correctly on multiple runs
git checkout Makefile 2>/dev/null || true

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

make clean

# Build our bundles before making the apptron worker
echo "Building distributions..."
# We run build_pum_bundle.sh FIRST so pum.tar.gz is built, but NOT build_distro.sh!
# We want to keep the original sys.tar.gz from Apptron intact and NOT overwrite it!
bash "$REPO_ROOT/apptron/scripts/build_pum_bundle.sh"

# Copy the generated custom bundle to the apptron worker build space
mkdir -p assets/bundles
cp "$PUM_ADMIN_ASSETS/bundles/pum.tar.gz" assets/bundles/

# Ensure wrangler is installed locally in the worker package
echo "Installing worker dependencies..."
cd worker && npm ci && cd ..

make all

echo "Starting Phase 2 Worker in dev mode..."
echo "Mock Auth is ENABLED in apptron/worker/src/auth.ts."
echo "You can access the environment at http://localhost:8788"
export CI=true
export WRANGLER_SEND_METRICS=false
cd worker && npx wrangler dev --port=8788 --log-level=none

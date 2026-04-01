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

def patch_auth_js(content):
    pattern = r'(if \(!getMeta\("auth-url"\)\) \{\n\s*throw new Error\("auth-url meta tag not found"\);\n\s*\}\))'
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
    return content.replace('register(getMeta("auth-url")', 'register(authUrl')

def patch_signin_html(content):
    return content.replace('<hanko-auth></hanko-auth>', """<script>
if (document.querySelector("meta[name='auth-url']")?.content === "/auth") {
    document.write("<p>Bypassing login...</p>");
} else {
    document.write("<hanko-auth></hanko-auth>");
}
</script>""")

def patch_worker_ts(content):
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

def patch_auth_ts(content):
    pattern = r'(export async function validateToken\(.*?\): Promise<boolean> \{)(.*?)(\n\})'
    replacement = r"""export async function validateToken(hankoApiUrl: string, token: string): Promise<boolean> {
// MOCK AUTH: Always return true
return true;
}"""
    return re.sub(pattern, replacement, content, flags=re.MULTILINE|re.DOTALL)

def patch_boot_go(content):
    replacement = """	webMod := web.New(k)
	for name, subfs := range webMod {
		if subfs == nil {
			delete(webMod, name)
		}
	}
	k.AddModule("#web", webMod)"""
    return content.replace('k.AddModule("#web", web.New(k))', replacement)

def patch_bundle(content):
    return content.replace('getBundle("/bundles/sys.tar.gz");', 'getBundle("/bundles/sys.tar.gz");\n    getBundle("/bundles/pum.tar.gz");')

if __name__ == "__main__":
    process_file("assets/lib/apptron.js", patch_auth_js)
    process_file("assets/signin.html", patch_signin_html)
    process_file("worker/src/worker.ts", patch_worker_ts)
    process_file("worker/src/auth.ts", patch_auth_ts)
    process_file("boot.go", patch_boot_go)
    process_file("assets/signin.html", patch_bundle)
    process_file("assets/dashboard.html", patch_bundle)

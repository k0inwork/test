import os
import glob
import re

for filepath in glob.glob("services/*/main.go"):
    with open(filepath, "r") as f:
        content = f.read()

    # Find the Capabilities block: Capabilities: []string{...}
    match = re.search(r'Capabilities:\s*\[\]string\{([^}]+)\}', content)
    if not match:
        continue

    caps_str = match.group(1)
    caps = [c.strip().strip('"') for c in caps_str.split(',') if c.strip()]

    # We want to replace it with []logging.CapabilityRegistration{ ... }
    new_caps = "Capabilities: []logging.CapabilityRegistration{\n"
    for cap in caps:
        # Default endpoint mapping
        ep = "/" + cap
        if cap in ['graphql']:
            ep = "/query"
        elif cap in ['auth']:
            ep = "/login"
        elif cap in ['audit']:
            ep = "/activitylist"
        elif cap in ['terminal', 'external-modules', 'external-data', 'inventory', 'network', 'compatibility', 'services', 'gws']:
            ep = "/"

        new_caps += f'\t\t\t{{Name: "{cap}", Endpoints: []string{{"{ep}"}}}},\n'
    new_caps += "\t\t}"

    content = content.replace(match.group(0), new_caps)

    # Some files might not have imported logging or might have weird spacing.

    with open(filepath, "w") as f:
        f.write(content)

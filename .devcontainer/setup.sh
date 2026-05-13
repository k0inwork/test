#!/bin/bash
set -e

echo "=== Installing Bun ==="
curl -fsSL https://bun.sh/install | bash
export BUN_INSTALL="$HOME/.bun"
export PATH="$BUN_INSTALL/bin:$PATH"
echo 'export BUN_INSTALL="$HOME/.bun"' >> ~/.bashrc
echo 'export PATH="$BUN_INSTALL/bin:$PATH"' >> ~/.bashrc

echo "=== Installing OpenClaude ==="
npm install -g @anthropic-ai/claude-code 2>/dev/null || true
bun install -g openclaude 2>/dev/null || npm install -g openclaude 2>/dev/null || {
    echo "Trying direct git install..."
    cd /tmp
    git clone https://github.com/Gitlawb/openclaude.git
    cd openclaude
    bun install
    bun link
    cd ~
}

echo "=== Configuring z.ai provider ==="
mkdir -p ~/.claude

# Write provider profile for z.ai/GLM
cat > ~/.claude/.openclaude-profile.json << 'PROFILE'
{
  "provider": "openai-compatible",
  "name": "z.ai GLM-5.1",
  "model": "glm-5.1",
  "base_url": "https://api.z.ai/api/coding/paas/v4",
  "api_key_env": "ZAI_API_KEY"
}
PROFILE

# Set env vars for the provider
cat >> ~/.bashrc << 'ENVVARS'

# z.ai / OpenClaude config
export ANTHROPIC_BASE_URL="https://api.z.ai/api/coding/paas/v4"
export ANTHROPIC_API_KEY="${ZAI_API_KEY}"
export OPENCLAUDE_MODEL="glm-5.1"
ENVVARS

echo "=== Setup complete ==="
echo "OpenClaude is ready with z.ai GLM-5.1"

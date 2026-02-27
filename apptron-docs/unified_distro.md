# The Unified PUM Admin Distro

## Overview
The Unified PUM Admin Distro provides a "single-click" installation experience for network administrators. It bundles the Apptron environment with a built-in Bridge Agent, ensuring seamless connectivity to the management network without manual configuration.

## Architecture
The distro is powered by the **Unified Runner** (`pum-admin`), a native Go application that performs the following tasks:
1.  **Embedded Web Server**: Serves the Apptron web assets (HTML, JS, WASM) and the customized Alpine bundle (`sys.tar.gz`).
2.  **Integrated Bridge Agent**: Automatically joins the Apptron virtual network and bridges it to the local physical network.
3.  **Auto-Configuration**: Injects the necessary session tokens and gateway IPs into the web environment so the user doesn't have to perform any setup.

## Component Flow
1.  Admin runs `./pum-admin`.
2.  The runner starts a local web server (e.g., on port 8080).
3.  The runner starts the Bridge Agent, which establishes a tunnel to the Apptron virtual network.
4.  The Admin opens `http://localhost:8080` in their browser.
5.  Apptron boots, detects the local gateway, and routes traffic destined for the management subnet through the integrated bridge.

## Invisible Rerouting
By pre-configuring the `resolv.conf` and routing tables within the custom `sys.tar.gz`, we ensure that traffic to specific CIDRs (e.g., your management network) is automatically sent to the Bridge Agent's virtual IP.

## Installation & Packaging
The entire environment is packaged as a single executable for Linux/Windows/macOS, including all necessary web assets embedded using Go's `embed` package.

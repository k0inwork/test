# Introduction: The Admin Command Center Vision

## Overview
The Product Unit Management (PUM) system is evolving from a traditional web-based management platform into a high-performance, unified "Command Center". This vision is realized through a modern hybrid architecture that combines the scalability of Go-based microservices with a dual-frontend strategy: standard Go-based web templates for accessibility and an Apptron-powered rich client for the full "Command Center" experience.

## The Vision
Our goal is to provide network administrators with a single, cohesive environment where they can monitor, diagnose, and remediate network issues with unprecedented speed and flexibility.

### Key Pillars
1.  **Distributed Power**: A backend of specialized Go microservices handles complex logic, state management, and high-volume data processing.
2.  **Hybrid Frontend Strategy**:
    -   **Go Web Frontend**: Standard HTML templates served via Gin for lightweight, zero-dependency access to core data.
    -   **Apptron Rich Client**: A dedicated runtime environment capable of hosting high-performance, local-first management applications.
3.  **Dynamic & Secure**: Features are discovered in real-time via service **Capabilities**, with access controlled by a granular role-to-capability mapping.
4.  **WASM/WASI Integration**: Advanced management tools are delivered as WebAssembly (WASM) binaries within the Apptron client, providing near-native execution speed directly in the browser.
5.  **Interface Synergy**: Seamless transitions between high-level visual Dashboards (WebViews) and powerful, keyboard-driven CLI/TUI tools.
6.  **Local-First, Cloud-Synced**: The client environment feels like a local machine, using WASM/WASI for local operations while bridging to the microservices for global state and device control.

## Target Audience
This documentation suite is intended for **internal stakeholders**, including developers, architects, and product managers, to provide a clear roadmap and architectural foundation for the ongoing transition.

# GUI to CLI/TUI Action Mapping Architecture

## Vision
To ensure consistency and power-user efficiency, every action available in the Web GUI (like "Reboot Device" or "Update VLAN") should be mirrored as a CLI command and a TUI interaction. This "UI Mirroring" strategy allows network admins to choose their preferred interface without losing functionality.

## The Action Registry Pattern
We avoid duplicating logic by using an **Action Registry**.

1.  **Core Logic**: Business logic is encapsulated in "Action" structs.
2.  **Shared Interface**: Each Action defines its input parameters, validation rules, and execution logic.
3.  **Multiple Frontends**:
    - **Web GUI**: Maps the Action to a button/form.
    - **CLI**: Maps the Action to a Cobra command.
    - **TUI**: Maps the Action to a Bubbletea list item or form.

## Architecture

```
[Shared Action Registry]
      |
      +--> [Web GUI (Gin Handler)] -> REST API
      |
      +--> [pum-cli (Cobra)] -> CLI command
      |
      +--> [pum-cli (Bubbletea)] -> TUI menu
```

## Handling 1:1 Mapping vs. Customization
While most actions follow a 1:1 mapping, the architecture allows for interface-specific enhancements:

- **CLI-specific**: Batch processing (e.g., `pum reboot --all`) which might not have a direct single-click equivalent in the GUI.
- **TUI-specific**: Real-time visual feedback and interactive wizards for complex multi-step configurations.
- **GUI-specific**: Geographical mapping and rich visual analytics.

## Implementation Details
Actions are defined in `pkg/actions`. A typical action includes:
- `Name`: Machine-readable ID (e.g., "node.reboot")
- `Params`: A struct defining required inputs.
- `Execute()`: The actual implementation that calls the backend microservices.

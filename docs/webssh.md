# WebSSH Terminal System

The PUM system includes a full-featured web-based SSH terminal, allowing administrators to manage remote devices directly from their browser.

## 1. Architecture

The WebSSH system is built using:
- **`django-webssh`**: A Django app providing the terminal UI and views.
- **`asyncssh`**: A Python library for asynchronous SSHv2.
- **Django Channels**: Manages the persistent WebSocket connection between the browser and the backend.

## 2. Key Features

### 2.1. Seamless Authentication
The system uses the user's current session to authenticate the SSH connection. The user's password (cached in `request.session['stored_password']` during login) is automatically used, providing a Single Sign-On (SSO) experience for remote management.

### 2.2. Session Recording
When a session starts, the system can record the entire interaction:
- **Persistence:** Raw terminal output and timing data are saved to `/var/log/arm/<uuid>.rec`.
- **Optimization:** Only "diffs" of the terminal screen are recorded when possible to save space.
- **Background Task:** A periodic background task (`call_periodic`) flushes the data to disk every 3 seconds to ensure no data is lost even if the session is interrupted.

### 2.3. Session Playback (Replay)
The system supports replaying recorded sessions:
- **Mechanism:** The `WebSSH` consumer reads the `.rec` file and "re-plays" the events back to the browser using the original timing data.
- **Interactive Playback:** Users can pause the playback via mouse events (Status 3).

## 3. Auditing Integration
Every SSH session is automatically logged in the `ActivityLog`:
- **Data Logged:** Username, Target Host, Source IP, and the filename of the session recording.
- **Compliance:** This provides a complete audit trail of all commands executed by administrators on remote hardware.

# ssh-config-view

A terminal-based TUI (Text User Interface) application for viewing and managing your SSH config file (`~/.ssh/config`). Built with [Bubbletea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), and [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Features

- **View** all SSH hosts defined in your `~/.ssh/config` file in a scrollable list.
- **Search** hosts by name using the `--search` or `-s` flag.
- **Add** new SSH host entries interactively.
- **Edit** existing SSH host entries.
- **Delete** SSH host entries with confirmation.
- **Persist** changes directly to your `~/.ssh/config` file.
- **Loading screen** for a smooth startup experience.

## Installation

1. **Clone the repository:**
   ```sh
   git clone <repo-url>
   cd ssh-config-view
   ```
2. **Build the application:**
   ```sh
   go build -o ssh-config-view
   ```
3. **Run the application:**
   ```sh
   ./ssh-config-view
   ```

## Usage

By default, the app loads and displays all hosts from your `~/.ssh/config` file.

### Command-line options

- `--search <term>` or `-s <term>`: Filter hosts by name or HostName.

Example:
```sh
./ssh-config-view --search myserver
```

## Keybindings

- `q` or `Ctrl+C`: Quit the application
- `Enter`: View details of the selected host
- `a`: Add a new host
- In details view:
  - `e`: Edit the selected host
  - `d`: Delete the selected host (with confirmation)
  - Any key: Return to the list
- In edit mode:
  - `Tab`/`Down`: Move to next field
  - `Shift+Tab`/`Up`: Move to previous field
  - `Enter`: Save changes
  - `Esc`: Cancel editing
- In confirmation dialogs: Any key to continue

## Notes

- The application **writes directly to your `~/.ssh/config` file**. Please ensure you have a backup if your config is important.
- Only basic SSH config fields are supported: `Host`, `HostName`, `User`, `Port`, and `IdentityFile`.
- Comments and advanced SSH config options are ignored and not preserved.

## Dependencies

- [Bubbletea](https://github.com/charmbracelet/bubbletea)
- [Bubbles](https://github.com/charmbracelet/bubbles)
- [Lipgloss](https://github.com/charmbracelet/lipgloss)

Install dependencies with:
```sh
go mod tidy
```

## License

MIT 
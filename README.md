# Terragrunt Vision

Terragrunt Vision is a powerful command-line tool designed to manage and execute multiple Terragrunt configurations **in parallel** with an intuitive, real-time interface built using the Bubble Tea library. It provides a comprehensive view of execution status, supports multiple terragrunt commands, and enables efficient multi-environment management.

## Features

### Core Functionality
- **Parallel Execution**: Run terragrunt commands across multiple environments simultaneously with goroutine-based parallelism
- **Multi-Select**: Select multiple stacks with the spacebar and execute commands on all of them at once
- **Multiple Commands**: Support for `init`, `plan`, `apply`, `validate`, and `destroy` operations
- **Real-time Progress Tracking**: Visual indicators showing running, success, and failed executions with color-coded status
- **Smart Workspace Scanning**: Automatically discovers and organizes terragrunt files by project, region, and stack

### User Interface
- **Three-Panel Layout**:
  - Left: Selectable list of terragrunt stacks with status indicators
  - Center: Code viewer with syntax highlighting
  - Right: Real-time execution output
- **Advanced Filtering**: Filter by project, region, and stack with multi-criteria support
- **Status Bar**: Shows current command, execution state, and selection count
- **Visual Feedback**: Color-coded selection indicators and execution status icons

### Integration & Configuration
- **AWS Integration**: Automatic AWS credential retrieval with profile and region support
- **Configuration File Support**: Customize behavior via JSON configuration files
- **Unique Output Files**: Each execution saves output to uniquely named files for debugging

## Installation

### From Source

```bash
git clone https://github.com/caiovfernandes/terragrunt-vision.git
cd terragrunt-vision
go build -o terragrunt-vision
```

### Using Docker

```bash
docker build -t terragrunt-vision .
docker run -it -v $(pwd)/workspaces:/workspaces terragrunt-vision /workspaces
```

## Usage

### Basic Usage

Run Terragrunt Vision by specifying the root directory of your Terragrunt configurations:

```bash
./terragrunt-vision <root-directory>
```

Example:
```bash
./terragrunt-vision ./workspaces
```

### Expected Directory Structure

Terragrunt Vision expects your workspace to follow this structure:

```
workspaces/
├── project-name/
│   ├── us-east-1/
│   │   ├── vpc/
│   │   │   └── terragrunt.hcl
│   │   ├── ec2/
│   │   │   └── terragrunt.hcl
│   ├── us-west-2/
│   │   ├── vpc/
│   │   │   └── terragrunt.hcl
```

### Key Bindings

#### Main View
- **`space`**: Toggle selection on current item
- **`i`**: Set command to `init`
- **`p`**: Set command to `plan`
- **`a`**: Set command to `apply`
- **`v`**: Set command to `validate`
- **`d`**: Set command to `destroy`
- **`enter`**: Execute selected command on all selected items (or current item if none selected)
- **`n`**: Open filter view
- **`j` / `down`**: Move cursor down
- **`k` / `up`**: Move cursor up
- **`q` / `ctrl+c`**: Quit the application

#### Filter View
- **`enter`**: Apply filter and return to main view
- **`n`**: Cancel and return to main view
- **`j` / `down`**: Move cursor down
- **`k` / `up`**: Move cursor up
- **`q` / `ctrl+c`**: Quit the application

### Workflow Example

1. **Launch the tool**: `./terragrunt-vision ./workspaces`
2. **Select stacks**: Use arrow keys to navigate, press `space` to select multiple stacks
3. **Choose command**: Press `p` for plan, `a` for apply, etc.
4. **Execute**: Press `enter` to run the command on all selected stacks in parallel
5. **Monitor progress**: Watch real-time status updates with visual indicators
6. **Review output**: Click on any stack to see its detailed execution output

## Configuration

Terragrunt Vision supports optional JSON configuration files. Create a `.terragrunt-vision.json` file in your project directory, home directory, or `/etc/terragrunt-vision/`:

```json
{
  "aws_region": "us-east-2",
  "aws_profile": "default",
  "max_parallel_executions": 10,
  "auto_approve": false,
  "save_outputs": true,
  "output_directory": "terragrunt-outputs",
  "show_status_icons": true,
  "color_scheme": "default"
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `aws_region` | string | `us-east-2` | Default AWS region |
| `aws_profile` | string | `default` | AWS profile to use |
| `max_parallel_executions` | int | `10` | Maximum concurrent executions |
| `auto_approve` | bool | `false` | Auto-approve apply/destroy operations |
| `save_outputs` | bool | `true` | Save execution outputs to files |
| `output_directory` | string | `terragrunt-outputs` | Directory for output files |
| `show_status_icons` | bool | `true` | Display status icons in UI |
| `color_scheme` | string | `default` | UI color scheme (default/light/dark) |

## Environment Variables

- **`AWS_REGION`**: Override AWS region (takes precedence over config file)
- **`AWS_PROFILE`**: Override AWS profile (takes precedence over config file)

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./terragrunt
go test ./config
```

## Dependencies

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**: Interactive terminal UI framework
- **[Glamour](https://github.com/charmbracelet/glamour)**: Markdown rendering with syntax highlighting
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)**: Terminal styling and layout
- **[AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2)**: AWS credential management

## Architecture

### Parallel Execution Engine

Terragrunt Vision uses Go's goroutines and channels to execute multiple terragrunt commands simultaneously:

1. User selects multiple stacks and presses enter
2. Each stack execution launches in a separate goroutine
3. Results are streamed back via channels as they complete
4. UI updates in real-time showing progress and status
5. Execution outputs are saved to unique files per stack

### Package Structure

```
terragrunt-vision/
├── main.go              # Entry point
├── ui/                  # UI layer (Bubble Tea models and views)
├── terragrunt/          # Terragrunt workspace parsing and execution
├── config/              # Configuration file management
└── utils/               # Utility functions (AWS credentials)
```

## Troubleshooting

### Common Issues

**Issue**: "No terragrunt files found in workspace"
- **Solution**: Ensure your directory structure follows the expected format with a `workspaces` folder containing `project/region/stack/terragrunt.hcl` hierarchy

**Issue**: AWS credential errors
- **Solution**: Verify your AWS profile is configured correctly with `aws configure` or set `AWS_PROFILE` environment variable

**Issue**: Parallel executions failing
- **Solution**: Check the output files in your execution directory for detailed error messages. Consider reducing `max_parallel_executions` in config.

## Performance

- **Parallel Execution**: Up to 10 concurrent operations by default (configurable)
- **Efficient Scanning**: Fast workspace discovery with optimized file traversal
- **Low Memory Footprint**: Streams output instead of buffering everything in memory

## Roadmap

- [ ] Add support for custom terragrunt commands
- [ ] Implement execution history and replay
- [ ] Add graph visualization of dependencies
- [ ] Support for remote state inspection
- [ ] Add dry-run mode for apply/destroy
- [ ] Export execution results to JSON/YAML

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/caiovfernandes/terragrunt-vision/blob/main/LICENSE) file for details.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charmbracelet
- Inspired by The Elm Architecture
- Thanks to the Terragrunt community for feedback and testing
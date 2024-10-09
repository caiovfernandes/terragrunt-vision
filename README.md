# Terragrunt Runner

Terragrunt Runner is a command-line tool designed to manage and execute Terragrunt configurations with a user-friendly interface built using the Bubble Tea library. It provides a real-time view of the execution output and allows filtering and navigation through different projects, regions, and stacks.

## Features

- **Real-time Execution Output**: View the output of `terragrunt init` in real-time.
- **Interactive UI**: Navigate through projects, regions, and stacks using keyboard shortcuts.
- **Filtering**: Filter items based on region.
- **AWS Integration**: Automatically retrieves AWS credentials for executing Terragrunt commands.

## Installation

To install the Terragrunt Runner, clone the repository and build the project using Go:

```bash
git clone https://github.com/caiovfernandes/terragrunt-runner.git
cd terragrunt-runner
go build
```

## Usage

Run the Terragrunt Runner by specifying the root directory of your Terragrunt configurations:

```bash
./terragrunt-runner <root-directory>
```

### Key Bindings

- **`ctrl+c`**: Quit the application.
- **`enter`**: Execute `terragrunt init` for the selected item.
- **`n`**: Navigate to the next view.
- **`j` / `down`**: Move the cursor down.
- **`k` / `up`**: Move the cursor up.

## Dependencies

- **Bubble Tea**: Used for building the interactive terminal UI.
- **AWS SDK for Go**: Used for retrieving AWS credentials.

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/caiovfernandes/terragrunt-runner/blob/main/LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## Acknowledgments

This project uses the Bubble Tea library by Charmbracelet, which is inspired by The Elm Architecture and go-tea.([1](https://github.com/charmbracelet/bubbletea/tree/v1.0.0))
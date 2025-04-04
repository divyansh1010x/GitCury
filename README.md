<div align="center">

# GitCury

[<img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/go/go-original.svg" width="60">](https://go.dev/)
[<img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/git/git-original.svg" width="60">](https://git-scm.com/)

[![Open in Visual Studio Code](https://img.shields.io/badge/Open%20in%20VS%20Code-007ACC?logo=visual-studio-code&logoColor=white)](https://vscode.dev/)
[![Contributors](https://img.shields.io/github/contributors/lakshyajain-0291/GitCury)](https://github.com/lakshyajain-0291/GitCury/graphs/contributors)
[![Forks](https://img.shields.io/github/forks/lakshyajain-0291/GitCury?style=social)](https://github.com/lakshyajain-0291/GitCury/network/members)
[![Stars](https://img.shields.io/github/stars/lakshyajain-0291/GitCury?style=social)](https://github.com/lakshyajain-0291/GitCury/stargazers)

</div>

## Overview

GitCury is a Go-based tool that automates Git commit message generation using the Gemini API. It not only streamlines your commit process but now leverages a configurable set of root folders to limit Git operations to specific areas of your file system. This means that the output commit messages are grouped by the designated root folders in the generated `output.json`.

## Key Features

- **AI-Powered Commit Messages**  
  Generate meaningful commit messages using the Gemini API. Messages are created based on file changes and differences.

- **Root Folder Filtering**  
  Configure multiple root folders in the `config.json` so that Git operations are scoped to specific directories. This is ideal when working on multi-repository projects or segmented codebases.

- **Grouped Output**  
  Commit messages are stored in `output.json` grouped by root folder. The structure looks like:  
  ```
  {
    "root_folder1": {
      "file1": "commit message",
      "file2": "commit message"
    },
    "root_folder2": {
      "file3": "commit message",
      "file4": "commit message"
    },
    ...
  }
  ```

- **Batch and Single-File Operations**  
  Prepare, commit, and push operations can be performed either in batch (all root folders) or on specific folders.

- **Configurable Parameters**  
  Easily update configuration settings (such as the number of files to commit and the root folders) via a simple API endpoint.

## Installation

<details>
<summary>Step-by-step guide</summary>

1. **Clone the repository:**
   ```bash
   git clone https://github.com/lakshyajain-0291/GitCury.git
   cd GitCury
   ```

2. **Set up the configuration:**  
   Edit the `config.json` file to update your GEMINI API key and list the root folders where Git operations should take place. An example configuration:
   ```json
   {
     "GEMINI_API_KEY": "YOUR_API_KEY",
     "app_name": "GitCury",
     "numFilesToCommit": 5,
     "root_folders": [
       "/path/to/folder1",
       "/path/to/folder2"
     ],
     "version": "1.0.0"
   }
   ```

3. **Build the project:**
   ```bash
   go build -o gitcury main.go
   ```

4. **Run the application:**
   ```bash
   ./gitcury
   ```
</details>

## Usage & API Endpoints

- **Server Startup:**  
  Start the server using:
  ```bash
  go run main.go
  ```
  The server listens on port 8080.

- **Update / Retrieve Configuration:**  
  - **GET /config:** Returns the current configuration.
  - **POST /config:** Updates configuration settings (including root folders).

- **Prepare Commit Messages:**  
  - **GET /getallmsgs:** Prepares commit messages for all configured root folders.  
  - **GET /getonemsgs?rootFolder=/path/to/folder:** Prepares commit messages for a specific root folder.

- **Commit Operations:**  
  - **GET /commitall:** Commits files from all grouped folders as per the generated messages.
  - **GET /commitone?rootFolder=/path/to/folder:** Commits files for a single root folder.

- **Push Operations:**  
  - **GET /pushall?branch=branchName:** Pushes all committed changes to the specified branch.
  - **GET /pushone?rootFolder=/path/to/folder&branch=branchName:** Pushes changes from a specific root folder.

## Example Workflow

1. **Start the server:**
   ```bash
   go run main.go
   ```

2. **Prepare commit messages for changed files across all root folders:**
   ```bash
   curl -X GET http://localhost:8080/getallmsgs
   ```
   or for a single folder:
   ```bash
   curl -X GET "http://localhost:8080/getonemsgs?rootFolder=/path/to/folder"
   ```

3. **Commit the prepared files:**
   To commit all folders:
   ```bash
   curl -X GET http://localhost:8080/commitall
   ```
   or to commit one folder:
   ```bash
   curl -X GET "http://localhost:8080/commitone?rootFolder=/path/to/folder"
   ```

4. **Push the commits:**
   For all folders:
   ```bash
   curl -X GET "http://localhost:8080/pushall?branch=main"
   ```
   For a specific folder:
   ```bash
   curl -X GET "http://localhost:8080/pushone?rootFolder=/path/to/folder&branch=main"
   ```

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository.
2. Create a feature branch:
   ```bash
   git checkout -b feature/NewFeature
   ```
3. Make your changes and commit them:
   ```bash
   git commit -m 'Add new feature'
   ```
4. Push your branch and open a Pull Request.

## License

GitCury is open-source, released under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Gemini API for AI-powered commit message generation.
- The open source community for continuous support and inspiration.

Happy coding! 🚀
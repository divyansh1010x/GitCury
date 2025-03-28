<div align="center">

# GitCury

[<img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/go/go-original.svg" width="60">](https://go.dev/)
[<img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/git/git-original.svg" width="60">](https://git-scm.com/)

[![Open in Visual Studio Code](https://img.shields.io/badge/Open%20in%20VS%20Code-007ACC?logo=visual-studio-code&logoColor=white)](https://vscode.dev/)
[![Contributors](https://img.shields.io/github/contributors/lakshyajain-0291/GitCury)](https://github.com/lakshyajain-0291/GitCury/graphs/contributors)
[![Forks](https://img.shields.io/github/forks/lakshyajain-0291/GitCury?style=social)](https://github.com/lakshyajain-0291/GitCury/network/members)
[![Stars](https://img.shields.io/github/stars/lakshyajain-0291/GitCury?style=social)](https://github.com/lakshyajain-0291/GitCury/stargazers)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/lakshyajain-0291/GitCury)
[![License](https://img.shields.io/badge/license-MIT-blue)](https://github.com/lakshyajain-0291/GitCury/blob/main/LICENSE)

*Automate Git Commit Messages with AI-Powered Suggestions*

[Features](#key-features) • [Installation](#installation) • [Usage](#usage) • [Contributing](#contribution)

</div>

## 🌟 Overview

**GitCury** is a Go-based tool designed to streamline your Git workflow by automating commit message generation. Powered by the GEMINI API, GitCury analyzes file changes and generates concise, project-specific commit messages.

## 🚀 Key Features

- 🤖 **AI-Powered Commit Messages**: Generate meaningful commit messages using the GEMINI API.
- 📂 **File-Specific Suggestions**: Tailored messages based on file type and changes.
- 🔄 **Batch Commit Preparation**: Prepare multiple files for commit in one go.
- ⚙️ **Configurable Settings**: Easily customize the number of files to process and other parameters.
- 🛠️ **Seamless Git Integration**: Works directly with your Git repository.

## 🌈 Why GitCury?

- **Efficiency**: Save time by automating commit message creation.
- **Consistency**: Ensure uniform and meaningful commit messages.
- **Flexibility**: Easily configurable for different workflows.
- **Integration**: Works seamlessly with existing Git commands.

## 📋 Prerequisites

- Go (1.24.1 or higher)
- GEMINI API Key
- Git

## 🔧 Installation

<details>
<summary>Step-by-step guide</summary>

1. Clone the repository:
```bash
git clone https://github.com/lakshyajain-0291/GitCury.git
cd GitCury
```

2. Set up GEMINI API Key:
```bash
# Add your API key to config.json
echo '{"GEMINI_API_KEY":"YOUR_API_KEY"}' > config.json
```

3. Build the project:
```bash
go build -o gitcury main.go
```

4. Run the application:
```bash
./gitcury
```
</details>

## 🎮 Usage

- **Start the server**:
    ```bash
    go run main.go
    ```
- **Prepare commit messages**:
    Send a POST request to `/getmessages` with the number of files to process.
- **Commit prepared files**:
    Send a POST request to `/commit` to commit files with generated messages.
- **Update configuration**:
    Send a POST request to `/config` with new settings.

## 🔑 Example Workflow

1. Start the server:
     ```bash
     go run main.go
     ```
2. Prepare commit messages for changed files:
     ```bash
     curl -X POST http://localhost:8080/getmessages -d '{"numFilesToCommit": 5}'
     ```
3. Commit the prepared files:
     ```bash
     curl -X POST http://localhost:8080/commit
     ```

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/NewFeature`).
3. Commit your changes (`git commit -m 'Add new feature'`).
4. Push to the branch (`git push origin feature/NewFeature`).
5. Open a Pull Request.

## 📜 License

GitCury is open-source, released under the MIT License. See `LICENSE` for details.

## 🙏 Acknowledgments

- GEMINI API for AI-powered commit message generation.
- Open-source community for inspiration and support.

**Happy coding!** 🚀  
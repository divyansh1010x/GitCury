<div align="center">

# 🌟 GitCury: Your Git Companion 🚀

[<img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/go/go-original.svg" width="60">](https://go.dev/)
[<img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/git/git-original.svg" width="60">](https://git-scm.com/)

[![Open in Visual Studio Code](https://img.shields.io/badge/Open%20in%20VS%20Code-007ACC?logo=visual-studio-code&logoColor=white)](https://vscode.dev/)
[![Contributors](https://img.shields.io/github/contributors/lakshyajain-0291/GitCury)](https://github.com/lakshyajain-0291/GitCury/graphs/contributors)
[![Forks](https://img.shields.io/github/forks/lakshyajain-0291/GitCury?style=social)](https://github.com/lakshyajain-0291/GitCury/network/members)
[![Stars](https://img.shields.io/github/stars/lakshyajain-0291/GitCury?style=social)](https://github.com/lakshyajain-0291/GitCury/stargazers)

</div>

## 🎉 Overview

GitCury is your ultimate Git sidekick! Built with Go, it automates your Git workflow with AI-powered commit messages, root folder filtering, and CLI commands that make Git operations a breeze. Whether you're managing a single repo or juggling multiple projects, GitCury has your back. 🌈

## ✨ Key Features

- **🤖 AI-Powered Commit Messages**  
  Let the Gemini API craft meaningful commit messages for you based on file changes. No more staring at your terminal in despair!

- **📂 Root Folder Filtering**  
  Scope Git operations to specific directories by configuring root folders in `config.json`. Perfect for multi-repo projects!

- **📊 Grouped Output**  
  Commit messages are neatly organized in `output.json` by root folder. Example:
  ```json
  {
    "root_folder1": {
      "file1": "commit message",
      "file2": "commit message"
    },
    "root_folder2": {
      "file3": "commit message",
      "file4": "commit message"
    }
  }
  ```

- **⚡ Batch and Single-File Operations**  
  Perform Git operations across all root folders or focus on just one. Flexibility at its finest!

- **🛠️ Configurable Parameters**  
  Update settings like the number of files to commit or root folders via a simple API or CLI commands. No more manual edits!

- **🎁 Hidden Gems**  
  Explore the codebase to uncover advanced features and Easter eggs. Who doesn’t love surprises?

- **🖋️ CLI Commands for Everything**  
  GitCury's CLI is packed with commands to manage configurations, generate commit messages, commit changes, and push them—all with a single command. It's like magic, but real!

## 🚀 Installation

<details>
<summary>Follow these simple steps:</summary>

1. **Clone the repository:**
   ```bash
   git clone https://github.com/lakshyajain-0291/GitCury.git
   cd GitCury
   ```

2. **Build the project:**
   ```bash
   go build -o gitcury main.go
   ```

3. **Run the application:**
   ```bash
   ./gitcury
   ```

4. **Set up the configuration:**  
  Use GitCury's CLI to configure your GEMINI API key and root folders:

  - Set your GEMINI API key:
    ```bash
    gitcury config set --key GEMINI_API_KEY --value YOUR_API_KEY
    ```

  - Add root folders:
    ```bash
    gitcury config set --key root_folders --value /path/to/folder1,/path/to/folder2
    ```

  - Optionally, configure the number of files to commit:
    ```bash
    gitcury config set --key numFilesToCommit --value 5
    ```

  - Verify your configuration:
    ```bash
    gitcury config
    ```


</details>

## 🛠️ CLI Commands

GitCury comes with a powerful CLI to make your Git workflow seamless. Here are some of the key commands:

### **Configuration Management**
- View the current configuration:
  ```bash
  gitcury config
  ```
- Set a configuration key-value pair:
  ```bash
  gitcury config set --key <key> --value <value>
  ```
- Remove a configuration key:
  ```bash
  gitcury config remove --key <key>
  ```

### **Generate Commit Messages**
- Generate commit messages for all root folders:
  ```bash
  gitcury getmsgs --all
  ```
- Generate commit messages for a specific folder:
  ```bash
  gitcury getmsgs --root <folder>
  ```

### **Commit Changes**
- Commit all files with generated messages:
  ```bash
  gitcury commit --all
  ```
- Commit files in a specific root folder:
  ```bash
  gitcury commit --root <folder>
  ```

### **Push Changes**
- Push all changes to a branch:
  ```bash
  gitcury push --all --branch <branch>
  ```
- Push changes for a specific folder:
  ```bash
  gitcury push --root <folder> --branch <branch>
  ```

### **Output Management**
- View all generated commit messages:
  ```bash
  gitcury output --log
  ```
- Edit the output file:
  ```bash
  gitcury output --edit
  ```
- Clear all generated commit messages:
  ```bash
  gitcury output --delete
  ```

## 🎯 Example Workflow

1. **Generate commit messages:**
   ```bash
   gitcury getmsgs --all
   ```

2. **Commit changes:**
   ```bash
   gitcury commit --all
   ```

3. **Push commits:**
   ```bash
   gitcury push --all --branch main
   ```

4. **View logs:**
   ```bash
   gitcury output --log
   ```

## 🤝 Contributing

We ❤️ contributions! Here's how you can help:

1. Fork the repo.  
2. Create a feature branch:  
   ```bash
   git checkout -b feature/NewFeature
   ```
3. Commit your changes:  
   ```bash
   git commit -m "Add NewFeature"
   ```
4. Push your branch and open a Pull Request.

## 📜 License

GitCury is open-source and licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🌟 Acknowledgments

- Thanks to the Gemini API for powering commit message generation.  
- Shoutout to the open-source community for their inspiration and support.

Happy coding with GitCury! 🎉✨

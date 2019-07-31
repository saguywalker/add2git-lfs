# add2git-lfs

A static web application that let you upload files and push them on a specific folder, remote and branch.

## Installation

1. Download an executable file directly from https://github.com/saguywalker/add2git-lfs/releases
2. Add it to your path enviroment.
    ```bash
    echo 'export PATH=path/to/your-add2git-lfs:$PATH' >> ~/.bashrc
    ```

## Usage
```bash
##### Run the command in your working repository

# Upload files with deafult config (remote: origin, branch: master, folder: sample-files)
add2git-lfs

# Upload files with specific configuration
add2git-lfs -remote upstream -branch dev -folder etc
```
